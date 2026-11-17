// Package storetest holds the Store contract once, so every store is held to
// the same rules: the desktop's SQLite record and the phone recorder's
// snapshot record (Amendment 21). A store's own suite calls Run with a way to
// open an empty store; what only one store does stays in that store's suite.
package storetest

import (
	"errors"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// Opener answers an empty store for one test.
type Opener func(t *testing.T) application.Store

// summerOffset is British Summer Time's offset from UTC.
const summerOffset = time.Hour

// instant answers a fixed instant with a one-hour offset, as British Summer
// Time writes it.
func instant(hour, minute int) time.Time {
	return time.Date(2026, time.September, 22, hour, minute, 0, 0,
		time.FixedZone("BST", int(summerOffset.Seconds())))
}

// input answers an event input for a label, occurring and recorded at a time.
func input(label string, when time.Time) application.EventInput {
	return application.EventInput{
		Label: label, Key: domain.KeyOf(label), OccurredAt: when, RecordedAt: when,
	}
}

// Run holds a store to the contract application.Store states.
func Run(t *testing.T, open Opener) {
	t.Run("a new record is empty", func(t *testing.T) { emptyRecord(t, open(t)) })
	t.Run("a note and symptom round trip byte for byte", func(t *testing.T) { roundTrip(t, open(t)) })
	t.Run("a case variant reuses the held label", func(t *testing.T) { caseVariant(t, open(t)) })
	t.Run("an update cannot move recorded-at", func(t *testing.T) { updateKeepsRecordedAt(t, open(t)) })
	t.Run("a deletion is all or nothing", func(t *testing.T) { deleteAllOrNothing(t, open(t)) })
	t.Run("a rename", func(t *testing.T) { rename(t, open(t)) })
	t.Run("an import adds definitions and events", func(t *testing.T) { importBoth(t, open(t)) })
}

func emptyRecord(t *testing.T, store application.Store) {
	events, err := store.Events()
	if err != nil || len(events) != 0 {
		t.Errorf("a new record holds %v, %v", events, err)
	}
	definitions, err := store.Definitions()
	if err != nil || len(definitions) != 0 {
		t.Errorf("a new record holds definitions %v, %v", definitions, err)
	}
}

func roundTrip(t *testing.T, store application.Store) {
	note := "  Much worse than yesterday.  \n\tAnd a second line; ümlaut, emoji \U0001F635."
	given := input(" Kopfschmerzen ", instant(12, 30))
	given.Note, given.Severity = note, domain.SeverityModerate
	saved, err := store.AddEvent(given)
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	read, err := store.Event(saved.ID)
	if err != nil {
		t.Fatalf("Event: %v", err)
	}
	if read.Note != note || read.Symptom != " Kopfschmerzen " || read.Severity != domain.SeverityModerate {
		t.Errorf("read back %+v", read)
	}
	if !read.OccurredAt.Equal(instant(12, 30)) || !read.RecordedAt.Equal(instant(12, 30)) {
		t.Errorf("times read back as %v and %v", read.OccurredAt, read.RecordedAt)
	}
}

func caseVariant(t *testing.T, store application.Store) {
	first, _ := store.AddEvent(input("Tired", instant(9, 0)))
	second, err := store.AddEvent(input("  TIRED", instant(17, 12)))
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	if second.Definition != first.Definition || second.Symptom != "Tired" {
		t.Errorf("the variant made %+v, want the held Tired", second)
	}
	definitions, _ := store.Definitions()
	if len(definitions) != 1 || !definitions[0].LastUsed.Equal(instant(17, 12)) {
		t.Errorf("definitions = %+v, want one, last used 17:12", definitions)
	}
}

func updateKeepsRecordedAt(t *testing.T, store application.Store) {
	saved, _ := store.AddEvent(input("Tired", instant(9, 30)))
	change := input("Headache", instant(8, 0))
	change.RecordedAt = instant(23, 59)
	change.Note = "edited"
	updated, err := store.UpdateEvent(saved.ID, change)
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if !updated.RecordedAt.Equal(instant(9, 30)) {
		t.Errorf("recorded-at moved to %v", updated.RecordedAt)
	}
	if updated.Symptom != "Headache" || !updated.OccurredAt.Equal(instant(8, 0)) || updated.Note != "edited" {
		t.Errorf("update did not take: %+v", updated)
	}
	if _, err := store.UpdateEvent(999, change); !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("updating a missing event = %v, want ErrNoSuchEvent", err)
	}
}

func deleteAllOrNothing(t *testing.T, store application.Store) {
	one, _ := store.AddEvent(input("Tired", instant(9, 0)))
	two, _ := store.AddEvent(input("Tired", instant(10, 0)))
	err := store.DeleteEvents([]domain.EventID{one.ID, 999})
	if !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("deleting with a missing id = %v, want ErrNoSuchEvent", err)
	}
	if events, _ := store.Events(); len(events) != 2 {
		t.Fatalf("a refused deletion removed events: %d left", len(events))
	}
	if err := store.DeleteEvents([]domain.EventID{one.ID, two.ID}); err != nil {
		t.Fatalf("DeleteEvents: %v", err)
	}
	if events, _ := store.Events(); len(events) != 0 {
		t.Errorf("%d events left after deleting both", len(events))
	}
	if _, err := store.Event(one.ID); !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("a deleted event reads %v", err)
	}
}

func rename(t *testing.T, store application.Store) {
	tired, _ := store.AddEvent(input("Tired", instant(9, 0)))
	_, _ = store.AddEvent(input("Headache", instant(10, 0)))
	if err := store.RenameDefinition(tired.Definition, "Exhausted", domain.KeyOf("Exhausted")); err != nil {
		t.Fatalf("RenameDefinition: %v", err)
	}
	if read, _ := store.Event(tired.ID); read.Symptom != "Exhausted" {
		t.Errorf("the event reads %q after the rename", read.Symptom)
	}
	if err := store.RenameDefinition(tired.Definition, "HEADACHE", domain.KeyOf("HEADACHE")); err == nil {
		t.Error("renaming onto a held key was accepted")
	}
	if err := store.RenameDefinition(999, "Nausea", domain.KeyOf("Nausea")); !errors.Is(err, domain.ErrNoSuchDefinition) {
		t.Errorf("renaming a missing definition = %v", err)
	}
}

func importBoth(t *testing.T, store application.Store) {
	definitions := []application.DefinitionInput{
		{Label: "Nausea", Key: domain.KeyOf("Nausea"), LastUsed: instant(8, 0)},
	}
	events := []application.EventInput{input("Tired", instant(9, 0)), input("tired", instant(10, 0))}
	if err := store.Import(definitions, events); err != nil {
		t.Fatalf("Import: %v", err)
	}
	held, _ := store.Definitions()
	if len(held) != 2 {
		t.Errorf("definitions after import = %+v, want Nausea and Tired", held)
	}
	if read, _ := store.Events(); len(read) != len(events) {
		t.Errorf("%d events after import, want %d", len(read), len(events))
	}
}
