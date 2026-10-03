package application

import (
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// filled answers a store holding Tired at 09:00 and 17:12 plus Headache at 12:30.
func filled(t *testing.T) (*fakeStore, History) {
	t.Helper()
	store := newFakeStore()
	recorder := NewRecorder(store, fixedClock{at(t, 18, 0)})
	for _, entry := range []struct {
		symptom      string
		hour, minute int
	}{{"Tired", 9, 0}, {"Headache", 12, 30}, {"Tired", 17, 12}} {
		when := at(t, entry.hour, entry.minute)
		if _, err := recorder.Record(RecordForm{Symptom: entry.symptom, OccurredAt: &when}); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	return store, NewHistory(store, london(t))
}

func TestNewestFirst(t *testing.T) {
	t.Parallel()
	_, history := filled(t)
	events, err := history.List(domain.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 3 || !events[0].OccurredAt.Equal(at(t, 17, 12)) ||
		!events[2].OccurredAt.Equal(at(t, 9, 0)) {
		t.Errorf("history = %+v, want newest first", events)
	}
}

func TestListRefusals(t *testing.T) {
	t.Parallel()
	store, history := filled(t)
	_, err := history.List(domain.Filter{
		From: domain.Date{Year: 2026, Month: time.October, Day: 1},
		To:   domain.Date{Year: 2026, Month: time.September, Day: 1},
	})
	wantIs(t, "reversed range", err, domain.ErrReversedRange)
	store.failReads = true
	_, err = history.List(domain.Filter{})
	wantIs(t, "failed read", err, ErrNotRead)
	_, err = history.Suggest("t")
	wantIs(t, "failed suggest", err, ErrNotRead)
	_, err = history.Symptoms()
	wantIs(t, "failed symptoms", err, ErrNotRead)
	_, err = history.Receipt(domain.Date{}, domain.Date{}, domain.Date{})
	wantIs(t, "failed receipt", err, ErrNotRead)
	wantIs(t, "failed rename read", history.Rename(1, "Exhausted"), ErrNotRenamed)
}

func TestSuggestFromTheStore(t *testing.T) {
	t.Parallel()
	_, history := filled(t)
	found, err := history.Suggest("ti")
	if err != nil || len(found) != 1 || found[0].Label != "Tired" {
		t.Errorf("Suggest(\"ti\") = %+v, %v", found, err)
	}
}

func TestSymptomsInLabelOrder(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	store.definitions = []domain.Definition{
		{ID: 3, Label: "tired"}, {ID: 1, Label: "Back pain"}, {ID: 2, Label: "back pain"},
	}
	symptoms, err := NewHistory(store, london(t)).Symptoms()
	if err != nil {
		t.Fatalf("Symptoms: %v", err)
	}
	if symptoms[0].ID != 1 || symptoms[1].ID != 2 || symptoms[2].ID != 3 {
		t.Errorf("order = %+v, want by label ignoring case, then by id", symptoms)
	}
}

func TestRenameReachesEveryEvent(t *testing.T) {
	t.Parallel()
	store, history := filled(t)
	tired, _ := domain.FindDefinition(store.definitions, "Tired")
	if err := history.Rename(tired.ID, "Exhausted"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	events, _ := history.List(domain.Filter{Definitions: []domain.DefinitionID{tired.ID}})
	for _, event := range events {
		if event.Symptom != "Exhausted" {
			t.Errorf("event still reads %q", event.Symptom)
		}
	}
	wantIs(t, "rename onto another", history.Rename(tired.ID, "headache"), domain.ErrDefinitionExists)
	store.failRename = true
	wantIs(t, "failed rename", history.Rename(tired.ID, "Weary"), ErrNotRenamed)
}

func TestReceiptFromTheStore(t *testing.T) {
	t.Parallel()
	_, history := filled(t)
	day := domain.Date{Year: 2026, Month: time.September, Day: 22}
	receipt, err := history.Receipt(day, day, domain.Date{})
	if err != nil {
		t.Fatalf("Receipt: %v", err)
	}
	if len(receipt.Groups) != 2 || receipt.Groups[0].Symptom != "Tired" {
		t.Errorf("groups = %+v, want Tired first (it occurred first)", receipt.Groups)
	}
}

func TestReceiptNamesTheAppointmentItCovers(t *testing.T) {
	t.Parallel()
	_, history := filled(t)
	day := domain.Date{Year: 2026, Month: time.September, Day: 22}
	receipt, err := history.Receipt(day, day, day)
	if err != nil || receipt.Appointment != day {
		t.Errorf("Receipt naming its own day = %+v, %v", receipt.Appointment, err)
	}
	earlier := domain.Date{Year: 2026, Month: time.September, Day: 2}
	_, err = history.Receipt(day, day, earlier)
	wantIs(t, "appointment outside the range", err, domain.ErrAppointmentOutsideRange)
	_, err = history.Receipt(domain.Date{}, day, domain.Date{})
	wantIs(t, "open range", err, domain.ErrOpenRange)
}
