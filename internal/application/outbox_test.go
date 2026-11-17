package application

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/oernster/symdiary/internal/domain"
)

// fakeSent remembers sends in memory, with a switch per operation to make it
// fail.
type fakeSent struct {
	sent, batch                   []domain.EventID
	failRead, failMark, failBatch bool
	// markCalls records that MarkSent was asked at all.
	markCalls bool
}

func (f *fakeSent) SentIDs() ([]domain.EventID, error) {
	if f.failRead {
		return nil, errBroken
	}
	return slices.Clone(f.sent), nil
}

func (f *fakeSent) LastBatch() ([]domain.EventID, error) {
	if f.failBatch {
		return nil, errBroken
	}
	return slices.Clone(f.batch), nil
}

func (f *fakeSent) MarkSent(ids []domain.EventID) error {
	f.markCalls = true
	if f.failMark {
		return errBroken
	}
	f.sent = append(f.sent, ids...)
	f.batch = slices.Clone(ids)
	return nil
}

// fakeEncoder encodes a record as the ids and labels it carries, which is
// enough to see what went into the file.
type fakeEncoder struct{ fail bool }

type encoded struct {
	Events      []domain.EventID
	Definitions []string
}

func (f fakeEncoder) Encode(record Record) ([]byte, error) {
	if f.fail {
		return nil, errBroken
	}
	var out encoded
	for _, event := range record.Events {
		out.Events = append(out.Events, event.ID)
	}
	for _, definition := range record.Definitions {
		out.Definitions = append(out.Definitions, definition.Label)
	}
	return json.Marshal(out)
}

// outboxOver answers an outbox over a store holding Tired at 08:00 and 09:00
// and Headache at 10:00, with the clock at 17:12.
func outboxOver(t *testing.T) (Outbox, *fakeStore, *fakeSent, []domain.EventID) {
	t.Helper()
	store, sent := newFakeStore(), &fakeSent{}
	var ids []domain.EventID
	for _, add := range []struct {
		label string
		hour  int
	}{{"Tired", 8}, {"Tired", 9}, {"Headache", 10}} {
		when := at(t, add.hour, 0)
		event, err := store.AddEvent(EventInput{
			Label: add.label, Key: domain.KeyOf(add.label), OccurredAt: when, RecordedAt: when,
		})
		if err != nil {
			t.Fatalf("adding %s: %v", add.label, err)
		}
		ids = append(ids, event.ID)
	}
	outbox := NewOutbox(store, sent, fakeEncoder{}, fixedClock{at(t, 17, 12)}, london(t))
	return outbox, store, sent, ids
}

func decode(t *testing.T, parcel Parcel) encoded {
	t.Helper()
	var out encoded
	if err := json.Unmarshal(parcel.Data, &out); err != nil {
		t.Fatalf("decoding the parcel: %v", err)
	}
	return out
}

func TestWaitingCountsTheUnsentAndNamesTheOldest(t *testing.T) {
	t.Parallel()
	outbox, _, sent, ids := outboxOver(t)
	sent.sent, sent.batch = ids[:1], ids[:1]

	waiting, err := outbox.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if waiting.Count != 2 || !waiting.Oldest.Equal(at(t, 9, 0)) || !waiting.HasBatch {
		t.Errorf("waiting = %+v; want 2 unsent, the oldest at 09:00, a batch held", waiting)
	}
}

func TestNothingWaitingBeforeTheFirstRecording(t *testing.T) {
	t.Parallel()
	outbox := NewOutbox(newFakeStore(), &fakeSent{}, fakeEncoder{}, fixedClock{at(t, 17, 12)}, london(t))
	waiting, err := outbox.Waiting()
	if err != nil || waiting.Count != 0 || !waiting.Oldest.IsZero() || waiting.HasBatch {
		t.Errorf("waiting = %+v, %v; want nothing at all", waiting, err)
	}
	wantIs(t, "Prepare with nothing unsent", outboxPrepareErr(outbox), ErrNothingToSend)
	_, err = outbox.Again()
	wantIs(t, "Again before any send", err, ErrNoBatch)
}

func outboxPrepareErr(outbox Outbox) error {
	_, err := outbox.Prepare()
	return err
}

func TestPrepareCarriesTheUnsentAndTheSymptomsTheyUse(t *testing.T) {
	t.Parallel()
	outbox, _, sent, ids := outboxOver(t)
	sent.sent = ids[2:] // Headache has gone

	parcel, err := outbox.Prepare()
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if parcel.Name != "SymDiary-phone-2026-09-22-1712.txt" {
		t.Errorf("name = %q", parcel.Name)
	}
	if !slices.Equal(parcel.Events, ids[:2]) {
		t.Errorf("parcel events = %v, want %v", parcel.Events, ids[:2])
	}
	carried := decode(t, parcel)
	if !slices.Equal(carried.Events, ids[:2]) || !slices.Equal(carried.Definitions, []string{"Tired"}) {
		t.Errorf("file holds %+v; want the two Tired events and Tired alone", carried)
	}
	if sent.markCalls {
		t.Error("Prepare marked events sent; only a completed share may (FR-090)")
	}
}

func TestConfirmMarksTheBatchAndAgainRepeatsIt(t *testing.T) {
	t.Parallel()
	outbox, store, _, ids := outboxOver(t)
	parcel, _ := outbox.Prepare()
	if err := outbox.Confirm(parcel.Events); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	wantIs(t, "Prepare once everything has gone", outboxPrepareErr(outbox), ErrNothingToSend)

	// A later recording is not part of the last batch (FR-095's acceptance).
	when := at(t, 11, 0)
	if _, err := store.AddEvent(EventInput{Label: "Nausea", Key: domain.KeyOf("Nausea"),
		OccurredAt: when, RecordedAt: when}); err != nil {
		t.Fatalf("adding Nausea: %v", err)
	}
	again, err := outbox.Again()
	if err != nil || !slices.Equal(again.Events, ids) {
		t.Errorf("Again = %v, %v; want the three events of the last batch", again.Events, err)
	}
	wantIs(t, "Confirm of nothing", outbox.Confirm(nil), ErrNothingToSend)
}

func TestAgainLeavesOutWhatWasDeletedSince(t *testing.T) {
	t.Parallel()
	outbox, store, sent, ids := outboxOver(t)
	sent.batch = ids[:1]
	store.events = slices.DeleteFunc(store.events, func(e domain.Event) bool { return e.ID == ids[0] })
	_, err := outbox.Again()
	wantIs(t, "Again of a batch whose events are all gone", err, ErrNoBatch)
}

func TestTheOutboxSaysWhyWhenItCannotRead(t *testing.T) {
	t.Parallel()
	outbox, store, sent, ids := outboxOver(t)
	sent.batch = ids

	sent.failRead = true
	_, err := outbox.Waiting()
	wantIs(t, "Waiting with sends unreadable", err, ErrNotRead)
	wantIs(t, "Prepare with sends unreadable", outboxPrepareErr(outbox), ErrNotRead)
	sent.failRead = false

	sent.failBatch = true
	_, err = outbox.Waiting()
	wantIs(t, "Waiting with the batch unreadable", err, ErrNotRead)
	_, err = outbox.Again()
	wantIs(t, "Again with the batch unreadable", err, ErrNotRead)
	sent.failBatch = false

	store.failReads = true
	wantIs(t, "Prepare with events unreadable", outboxPrepareErr(outbox), ErrNotRead)
	_, err = outbox.Again()
	wantIs(t, "Again with events unreadable", err, ErrNotRead)
}

func TestTheOutboxSaysWhyWhenItCannotWrite(t *testing.T) {
	t.Parallel()
	outbox, store, sent, ids := outboxOver(t)

	broken := NewOutbox(store, sent, fakeEncoder{fail: true}, fixedClock{at(t, 17, 12)}, london(t))
	wantIs(t, "Prepare with the encoding failing", outboxPrepareErr(broken), ErrNotExported)

	sent.failMark = true
	wantIs(t, "Confirm with the mark failing", outbox.Confirm(ids), ErrNotSent)
}

func TestPrepareSaysWhyWhenTheSymptomsCannotBeRead(t *testing.T) {
	t.Parallel()
	outbox, _, _, _ := outboxOver(t)
	failing := &definitionsFail{fakeStore: newFakeStore()}
	when := at(t, 8, 0)
	if _, err := failing.AddEvent(EventInput{Label: "Tired", Key: domain.KeyOf("Tired"),
		OccurredAt: when, RecordedAt: when}); err != nil {
		t.Fatalf("adding Tired: %v", err)
	}
	outbox.store = failing
	wantIs(t, "Prepare with definitions unreadable", outboxPrepareErr(outbox), ErrNotRead)
}

// definitionsFail reads events but not definitions, the one read the parcel
// makes after it has its events.
type definitionsFail struct{ *fakeStore }

func (definitionsFail) Definitions() ([]domain.Definition, error) { return nil, errBroken }

func TestTheDesktopNeverSends(t *testing.T) {
	t.Parallel()
	var desktop NothingSent
	sent, err := desktop.SentIDs()
	batch, batchErr := desktop.LastBatch()
	if sent != nil || err != nil || batch != nil || batchErr != nil {
		t.Errorf("NothingSent answered %v, %v, %v, %v; want nothing", sent, err, batch, batchErr)
	}
	wantIs(t, "MarkSent on the desktop", desktop.MarkSent([]domain.EventID{1}), ErrNothingToSend)
}
