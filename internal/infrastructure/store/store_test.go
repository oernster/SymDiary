package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/infrastructure/storetest"
)

// open answers a store over a fresh file in a folder whose name has a space in
// it, as %APPDATA% under a user name with a space would.
func open(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "App Data", "SymDiary", "symdiary.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, path
}

// instant answers a fixed instant with a one-hour offset, as British Summer
// Time writes it.
func instant(hour, minute int) time.Time {
	return time.Date(2026, time.September, 22, hour, minute, 0, 0, time.FixedZone("BST", 3600))
}

// input answers an event input for a label, occurring and recorded at a time.
func input(label string, when time.Time) application.EventInput {
	return application.EventInput{
		Label: label, Key: domain.KeyOf(label), OccurredAt: when, RecordedAt: when,
	}
}

func TestMissingFileGivesEmptyRecord(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the record file was not created: %v", err)
	}
	events, err := store.Events()
	if err != nil || len(events) != 0 {
		t.Errorf("a new record holds %v, %v", events, err)
	}
	definitions, err := store.Definitions()
	if err != nil || len(definitions) != 0 {
		t.Errorf("a new record holds definitions %v, %v", definitions, err)
	}
}

// TestTheStoreContract holds the SQLite record to the rules every store keeps;
// the tests below it cover what only this store does.
func TestTheStoreContract(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T) application.Store {
		store, _ := open(t)
		return store
	})
}

// refuseNote plants a trigger that makes any insert of an event with the given
// note fail, so a failure can be forced part way through a write.
func refuseNote(t *testing.T, store *Store, note string) {
	t.Helper()
	if _, err := store.db.Exec(`CREATE TRIGGER refuse BEFORE INSERT ON events
		WHEN NEW.note = '` + note + `' BEGIN SELECT RAISE(ABORT, 'refused'); END`); err != nil {
		t.Fatal(err)
	}
}

func TestFailedImportLeavesNothing(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	refuseNote(t, store, "boom")
	definitions := []application.DefinitionInput{
		{Label: "Nausea", Key: domain.KeyOf("Nausea"), LastUsed: instant(8, 0)},
	}
	good := input("Dizziness", instant(11, 0))
	bad := input("Dizziness", instant(12, 0))
	bad.Note = "boom"
	if err := store.Import(definitions, []application.EventInput{good, bad}); err == nil {
		t.Fatal("the planted refusal did not fail the import")
	}
	events, _ := store.Events()
	held, _ := store.Definitions()
	if len(events) != 0 || len(held) != 0 {
		t.Errorf("a failed import left %d events and %d definitions", len(events), len(held))
	}
}

func TestFailedAddLeavesNothing(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	refuseNote(t, store, "boom")
	bad := input("Tired", instant(9, 0))
	bad.Note = "boom"
	if _, err := store.AddEvent(bad); err == nil {
		t.Fatal("the planted refusal did not fail the add")
	}
	if held, _ := store.Definitions(); len(held) != 0 {
		t.Errorf("a failed add left the definition it created: %+v", held)
	}
}

func TestUnreadableValuesAreRefusedNotGuessed(t *testing.T) {
	t.Parallel()
	for _, column := range []string{"occurred_at", "recorded_at", "severity"} {
		store, _ := open(t)
		saved, _ := store.AddEvent(input("Tired", instant(9, 0)))
		if _, err := store.db.Exec(`UPDATE events SET ` + column + ` = 'garbage'`); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Events(); err == nil {
			t.Errorf("a garbage %s read back without complaint", column)
		}
		if _, err := store.Event(saved.ID); err == nil {
			t.Errorf("a garbage %s read back one event without complaint", column)
		}
	}
	store, _ := open(t)
	_, _ = store.AddEvent(input("Tired", instant(9, 0)))
	if _, err := store.db.Exec(`UPDATE definitions SET last_used = 'garbage'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Definitions(); err == nil {
		t.Error("a garbage last_used read back without complaint")
	}
	if _, err := store.AddEvent(input("Tired", instant(10, 0))); err == nil {
		t.Error("resolving onto a garbage last_used should fail")
	}
}

func TestClosedStoreRefusesEverything(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	_ = store.Close()
	if _, err := store.Definitions(); err == nil {
		t.Error("Definitions on a closed store")
	}
	if _, err := store.Events(); err == nil {
		t.Error("Events on a closed store")
	}
	if _, err := store.AddEvent(input("Tired", instant(9, 0))); err == nil {
		t.Error("AddEvent on a closed store")
	}
	if err := store.DeleteEvents([]domain.EventID{1}); err == nil {
		t.Error("DeleteEvents on a closed store")
	}
	if err := store.RenameDefinition(1, "X", "x"); err == nil {
		t.Error("RenameDefinition on a closed store")
	}
	if err := store.Import(nil, nil); err == nil {
		t.Error("Import on a closed store")
	}
}

func TestOpenRefusesAFolderItCannotMake(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(filepath.Join(blocker, "symdiary.db")); err == nil {
		t.Error("a record under a file opened")
	}
}

func TestReopenKeepsEverything(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	saved, _ := store.AddEvent(input("Tired", instant(9, 0)))
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	again, err := Open(path)
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	defer again.Close()
	if read, err := again.Event(saved.ID); err != nil || read.Symptom != "Tired" {
		t.Errorf("after reopening: %+v, %v", read, err)
	}
}

func TestCorruptFileIsNeverReplaced(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "symdiary.db")
	garbage := bytes.Repeat([]byte("not a database "), 512)
	if err := os.WriteFile(path, garbage, 0o600); err != nil {
		t.Fatal(err)
	}
	if store, err := Open(path); err == nil {
		_ = store.Close()
		t.Fatal("a garbage file opened as a record")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, garbage) {
		t.Errorf("the unreadable file was changed (%v)", err)
	}
}

func TestNewerSchemaRefused(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	if _, err := store.db.Exec(`PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	_ = store.Close()
	if _, err := Open(path); !errors.Is(err, ErrNewerSchema) {
		t.Errorf("a newer schema = %v, want ErrNewerSchema", err)
	}
}

func TestDurabilityPragmas(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	// synchronous FULL is 2; see https://sqlite.org/pragma.html#pragma_synchronous.
	const synchronousFull = 2
	var synchronous, foreignKeys, version int
	var journal string
	_ = store.db.QueryRow(`PRAGMA synchronous`).Scan(&synchronous)
	_ = store.db.QueryRow(`PRAGMA journal_mode`).Scan(&journal)
	_ = store.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys)
	_ = store.db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if synchronous != synchronousFull || journal != "wal" || foreignKeys != 1 {
		t.Errorf("synchronous=%d journal=%s foreign_keys=%d", synchronous, journal, foreignKeys)
	}
	if version != len(migrations) {
		t.Errorf("schema version %d, want %d", version, len(migrations))
	}
}
