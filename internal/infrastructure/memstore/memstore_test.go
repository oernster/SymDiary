package memstore

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/infrastructure/storetest"
)

// errFull is the failure the fake snapshots inject.
var errFull = errors.New("the phone's storage is full")

// fakeSnapshots keeps the saved bytes in memory, standing in for the browser.
type fakeSnapshots struct {
	saved              []byte
	failLoad, failSave bool
}

func (f *fakeSnapshots) Load() ([]byte, error) {
	if f.failLoad {
		return nil, errFull
	}
	return f.saved, nil
}

func (f *fakeSnapshots) Save(data []byte) error {
	if f.failSave {
		return errFull
	}
	f.saved = slices.Clone(data)
	return nil
}

// opened answers a store over fresh snapshots.
func opened(t *testing.T) (*Store, *fakeSnapshots) {
	t.Helper()
	snapshots := &fakeSnapshots{}
	store, err := Open(snapshots)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return store, snapshots
}

// tired answers an input for Tired at the hour.
func tired(hour int) application.EventInput {
	when := time.Date(2026, time.November, 16, hour, 0, 0, 0, time.UTC)
	return application.EventInput{Label: "Tired", Key: domain.KeyOf("Tired"), OccurredAt: when, RecordedAt: when}
}

func TestTheStoreContract(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T) application.Store {
		store, _ := opened(t)
		return store
	})
}

func TestTheRecordSurvivesReopening(t *testing.T) {
	t.Parallel()
	store, snapshots := opened(t)
	added, _ := store.AddEvent(tired(8))
	if err := store.MarkSent([]domain.EventID{added.ID}); err != nil {
		t.Fatalf("MarkSent: %v", err)
	}
	again, err := Open(snapshots)
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	held, err := again.Event(added.ID)
	if err != nil || !held.OccurredAt.Equal(added.OccurredAt) || held.Symptom != "Tired" {
		t.Errorf("after reopening = %+v, %v; want %+v", held, err, added)
	}
	sent, _ := again.SentIDs()
	batch, _ := again.LastBatch()
	if !slices.Equal(sent, []domain.EventID{added.ID}) || !slices.Equal(batch, sent) {
		t.Errorf("after reopening sent = %v, batch = %v", sent, batch)
	}
	next, _ := again.AddEvent(tired(9))
	if next.ID == added.ID {
		t.Error("a reopened record handed out an id already used")
	}
}

func TestAFailedSaveChangesNothing(t *testing.T) {
	t.Parallel()
	store, snapshots := opened(t)
	kept, _ := store.AddEvent(tired(8))
	before := slices.Clone(snapshots.saved)
	snapshots.failSave = true

	writes := map[string]func() error{
		"add": func() error { _, err := store.AddEvent(tired(9)); return err },
		"update": func() error {
			_, err := store.UpdateEvent(kept.ID, tired(10))
			return err
		},
		"delete": func() error { return store.DeleteEvents([]domain.EventID{kept.ID}) },
		"rename": func() error {
			return store.RenameDefinition(kept.Definition, "Exhausted", domain.KeyOf("Exhausted"))
		},
		"import": func() error { return store.Import(nil, []application.EventInput{tired(11)}) },
		"mark":   func() error { return store.MarkSent([]domain.EventID{kept.ID}) },
	}
	for name, write := range writes {
		if err := write(); !errors.Is(err, errFull) {
			t.Errorf("%s with the storage full = %v, want the reason (FR-101)", name, err)
		}
	}
	events, _ := store.Events()
	sent, _ := store.SentIDs()
	if len(events) != 1 || events[0] != kept || len(sent) != 0 || !slices.Equal(snapshots.saved, before) {
		t.Errorf("a refused save changed the record: %+v, sent %v", events, sent)
	}
}

func TestATimeTheSnapshotCannotHoldIsRefused(t *testing.T) {
	t.Parallel()
	// JSON holds years 0 to 9999 alone; a time beyond them cannot be saved, so
	// the write is refused and nothing changes, rather than a record that will
	// not open again.
	store, snapshots := opened(t)
	distant := tired(8)
	distant.OccurredAt = time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)
	if _, err := store.AddEvent(distant); err == nil {
		t.Error("an event in the year 10000 was saved")
	}
	if events, _ := store.Events(); len(events) != 0 || snapshots.saved != nil {
		t.Errorf("a refused write changed the record: %+v", events)
	}
}

func TestMarkSentReplacesTheBatchAndKeepsEveryID(t *testing.T) {
	t.Parallel()
	store, _ := opened(t)
	one, _ := store.AddEvent(tired(8))
	two, _ := store.AddEvent(tired(9))
	_ = store.MarkSent([]domain.EventID{one.ID})
	_ = store.MarkSent([]domain.EventID{one.ID, two.ID})
	sent, _ := store.SentIDs()
	batch, _ := store.LastBatch()
	want := []domain.EventID{one.ID, two.ID}
	if !slices.Equal(sent, want) || !slices.Equal(batch, want) {
		t.Errorf("sent = %v, batch = %v; want each id once and the last batch whole", sent, batch)
	}
}

func TestASnapshotThatCannotBeTrustedIsRefused(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		saved []byte
		want  error
	}{
		"newer":       {[]byte(`{"schema":2}`), ErrNewerSchema},
		"unversioned": {[]byte(`{"events":[]}`), ErrNoSchema},
		"orphan event": {[]byte(`{"schema":1,"nextId":3,"definitions":[],` +
			`"events":[{"id":2,"definition":1,"severity":""}]}`), ErrDamaged},
		"bad severity": {[]byte(`{"schema":1,"nextId":3,"definitions":[{"id":1,"label":"Tired"}],` +
			`"events":[{"id":2,"definition":1,"severity":"awful"}]}`), ErrDamaged},
	}
	for name, given := range cases {
		if _, err := Open(&fakeSnapshots{saved: given.saved}); !errors.Is(err, given.want) {
			t.Errorf("%s snapshot = %v, want %v", name, err, given.want)
		}
	}
	for name, saved := range map[string][]byte{
		"not JSON":      []byte(`not json`),
		"wrong shape":   []byte(`{"schema":1,"events":"many"}`),
		"unreadable id": []byte(`{"schema":"one"}`),
	} {
		if _, err := Open(&fakeSnapshots{saved: saved}); err == nil {
			t.Errorf("a %s snapshot was opened", name)
		}
	}
	if _, err := Open(&fakeSnapshots{failLoad: true}); !errors.Is(err, errFull) {
		t.Errorf("a refused load = %v, want the reason", err)
	}
}

func TestARenameMayChangeOnlyTheCase(t *testing.T) {
	t.Parallel()
	store, _ := opened(t)
	added, _ := store.AddEvent(tired(8))
	if err := store.RenameDefinition(added.Definition, "TIRED", domain.KeyOf("TIRED")); err != nil {
		t.Errorf("renaming to its own key in capitals = %v", err)
	}
	other, _ := store.AddEvent(application.EventInput{
		Label: "Headache", Key: domain.KeyOf("Headache"),
		OccurredAt: tired(9).OccurredAt, RecordedAt: tired(9).RecordedAt,
	})
	err := store.RenameDefinition(other.Definition, "tired", domain.KeyOf("tired"))
	if !errors.Is(err, ErrKeyHeld) {
		t.Errorf("renaming onto a held key = %v, want ErrKeyHeld", err)
	}
}
