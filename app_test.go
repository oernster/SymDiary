package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/pdf"
	"github.com/oernster/symdiary/internal/infrastructure/store"
)

// stoppedClock is a clock the test moves by hand.
type stoppedClock struct{ now time.Time }

func (c *stoppedClock) Now() time.Time { return c.now }

// fakeChooser answers the paths a test has set, standing in for the window's
// file dialogs. An empty path is a cancelled dialog.
type fakeChooser struct {
	save, open string
	err        error
	// asked is the kind the last save dialog was opened for. macOS applies the
	// filter to the name, so a dialog opened for the wrong kind writes a file
	// nothing will open.
	asked fileKind
	// suggested is the name that dialog was opened with, kept because the name
	// and the kind have to agree: it is their disagreement that reaches the user.
	suggested string
}

func (c *fakeChooser) SavePath(suggested string, kind fileKind) (string, error) {
	c.asked, c.suggested = kind, suggested
	return c.save, c.err
}
func (c *fakeChooser) OpenPath() (string, error) { return c.open, c.err }

// london is the zone the facade works in.
func london(t *testing.T) *time.Location {
	t.Helper()
	zone, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("loading Europe/London: %v", err)
	}
	return zone
}

// facade answers an App over a real record in a temporary folder.
func facade(t *testing.T) (*App, *stoppedClock, *fakeChooser) {
	t.Helper()
	folder := t.TempDir()
	opened, err := store.Open(filepath.Join(folder, "symdiary.db"))
	if err != nil {
		t.Fatalf("opening the record: %v", err)
	}
	t.Cleanup(func() { _ = opened.Close() })
	return facadeOver(t, opened, folder)
}

// facadeOver answers an App over any store.
func facadeOver(t *testing.T, record application.Store, folder string) (*App, *stoppedClock, *fakeChooser) {
	t.Helper()
	zone := london(t)
	clock := &stoppedClock{now: time.Date(2026, time.September, 22, 17, 12, 0, 0, zone)}
	chooser := &fakeChooser{save: filepath.Join(folder, "export.json")}
	services := Services{
		Recorder: application.NewRecorder(record, clock),
		Editor:   application.NewEditor(record, clock, zone),
		History:  application.NewHistory(record, zone),
		Transfer: application.NewTransfer(record, export.File{}),
	}
	app := newApp(services, clock, zone, "1.2.3", "", chooser, func() error { return nil })
	// The real renderer, writing into the test's own folder. A double would
	// assert that the double was called; this asserts that a file arrives.
	app.sheet = pdf.Sheet{}
	return app, clock, chooser
}

// record records one symptom, failing the test where it is refused.
func record(t *testing.T, app *App, form RecordFormDTO) EventDTO {
	t.Helper()
	saved, err := app.Record(form)
	if err != nil {
		t.Fatalf("Record(%+v): %v", form, err)
	}
	return saved
}

func TestStateAndAbout(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	state, err := app.State()
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if state.Name != "SymDiary" || state.Version != "1.2.3" || state.Problem != "" {
		t.Errorf("state = %+v", state)
	}
	if strings.Join(state.Severities, ",") != "Mild,Moderate,Severe" {
		t.Errorf("severities = %v", state.Severities)
	}
	about, err := app.About()
	if err != nil || about.Version != "1.2.3" || len(about.Credits) == 0 {
		t.Errorf("about = %+v, %v", about, err)
	}
	now, err := app.Now()
	if err != nil || now != "2026-09-22T17:12" {
		t.Errorf("Now = %q, %v", now, err)
	}
}

func TestRecordThenHistory(t *testing.T) {
	t.Parallel()
	app, clock, _ := facade(t)
	saved := record(t, app, RecordFormDTO{
		Symptom: "Tired", Severity: "Mild", Note: "Only been awake for about 10 minutes.",
	})
	if saved.When != "22 Sep 2026 17:12" || saved.OccurredAt != "2026-09-22T17:12" {
		t.Errorf("saved = %+v", saved)
	}

	clock.now = clock.now.Add(time.Hour)
	record(t, app, RecordFormDTO{Symptom: "Headache", OccurredAt: "2026-09-22T12:30"})

	events, err := app.History(FilterDTO{})
	if err != nil || len(events) != 2 || events[0].Symptom != "Tired" {
		t.Fatalf("history = %+v, %v; want Tired first (newest)", events, err)
	}
	narrowed, err := app.History(FilterDTO{
		From: "2026-09-22", To: "2026-09-22", Severities: []string{"Mild"},
	})
	if err != nil || len(narrowed) != 1 || narrowed[0].Severity != "Mild" {
		t.Errorf("narrowed = %+v, %v", narrowed, err)
	}
	suggestions, err := app.Suggest("he")
	if err != nil || len(suggestions) != 1 || suggestions[0].Label != "Headache" {
		t.Errorf("Suggest = %+v, %v", suggestions, err)
	}
	symptoms, err := app.Symptoms()
	if err != nil || len(symptoms) != 2 {
		t.Errorf("Symptoms = %+v, %v", symptoms, err)
	}
}

func TestRecordRefusals(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	if _, err := app.Record(RecordFormDTO{Symptom: " "}); !errors.Is(err, domain.ErrBlankSymptom) {
		t.Errorf("blank symptom = %v", err)
	}
	if _, err := app.Record(RecordFormDTO{Symptom: "Tired", OccurredAt: "noon"}); !errors.Is(err, domain.ErrBadDate) {
		t.Errorf("unreadable time = %v", err)
	}
	if _, err := app.Record(RecordFormDTO{Symptom: "Tired", Severity: "awful"}); !errors.Is(err, domain.ErrUnknownSeverity) {
		t.Errorf("unknown severity = %v", err)
	}
	if _, err := app.Record(RecordFormDTO{
		Symptom: "Tired", OccurredAt: "2026-09-22T17:13",
	}); !errors.Is(err, domain.ErrFutureOccurrence) {
		t.Errorf("future occurrence = %v", err)
	}
}

func TestEditAndDelete(t *testing.T) {
	t.Parallel()
	app, clock, _ := facade(t)
	saved := record(t, app, RecordFormDTO{Symptom: "Tired"})
	clock.now = clock.now.Add(3 * time.Hour)

	edited, err := app.Edit(saved.ID, EditFormDTO{Symptom: "Tired", Note: "Went away."})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if edited.OccurredAt != saved.OccurredAt {
		t.Errorf("the occurrence moved from %s to %s", saved.OccurredAt, edited.OccurredAt)
	}
	moved, err := app.Edit(saved.ID, EditFormDTO{Symptom: "Tired", OccurredAt: "2026-09-22T12:30"})
	if err != nil || moved.When != "22 Sep 2026 12:30" {
		t.Errorf("moved = %+v, %v", moved, err)
	}

	prompt, err := app.DeletionPrompt([]int64{saved.ID})
	if err != nil || !strings.Contains(prompt, "Tired") || !strings.Contains(prompt, "12:30") {
		t.Errorf("prompt = %q, %v", prompt, err)
	}
	if err := app.Delete([]int64{saved.ID}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if events, _ := app.History(FilterDTO{}); len(events) != 0 {
		t.Errorf("%d events left after deleting", len(events))
	}
}

func TestEditAndFilterRefusals(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	saved := record(t, app, RecordFormDTO{Symptom: "Tired"})
	for name, call := range map[string]func() error{
		"unreadable time": func() error {
			_, err := app.Edit(saved.ID, EditFormDTO{Symptom: "Tired", OccurredAt: "noon"})
			return err
		},
		"unknown severity": func() error {
			_, err := app.Edit(saved.ID, EditFormDTO{Symptom: "Tired", Severity: "awful"})
			return err
		},
		"missing event": func() error {
			_, err := app.Edit(999, EditFormDTO{Symptom: "Tired"})
			return err
		},
		"bad from": func() error {
			_, err := app.History(FilterDTO{From: "yesterday"})
			return err
		},
		"bad to": func() error {
			_, err := app.History(FilterDTO{To: "tomorrow"})
			return err
		},
		"bad filter severity": func() error {
			_, err := app.History(FilterDTO{Severities: []string{"awful"}})
			return err
		},
		"bad receipt from": func() error {
			_, err := app.Receipt("first", "2026-09-22", "")
			return err
		},
		"bad receipt to": func() error {
			_, err := app.Receipt("2026-09-01", "last", "")
			return err
		},
		"delete nothing": func() error { return app.Delete(nil) },
		"prompt for nothing": func() error {
			_, err := app.DeletionPrompt(nil)
			return err
		},
	} {
		if err := call(); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestRename(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	record(t, app, RecordFormDTO{Symptom: "Tired"})

	symptoms, _ := app.Symptoms()
	if err := app.Rename(symptoms[0].ID, "Exhausted"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	events, _ := app.History(FilterDTO{})
	if events[0].Symptom != "Exhausted" {
		t.Errorf("after the rename the event reads %q", events[0].Symptom)
	}
	if err := app.Rename(999, "Nausea"); !errors.Is(err, domain.ErrNoSuchDefinition) {
		t.Errorf("renaming a missing symptom = %v", err)
	}
}

func TestExportAndImport(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	record(t, app, RecordFormDTO{Symptom: "Tired", Severity: "Mild"})

	path, err := app.Export()
	if err != nil || path != chooser.save {
		t.Fatalf("Export = %q, %v", path, err)
	}

	empty, _, receiver := facade(t)
	receiver.open = path
	result, err := empty.Import()
	if err != nil || !result.Chosen || result.Added != 1 || result.Skipped != 0 {
		t.Fatalf("Import = %+v, %v", result, err)
	}
	again, err := empty.Import()
	if err != nil || again.Added != 0 || again.Skipped != 1 {
		t.Errorf("a second import = %+v, %v", again, err)
	}
}

func TestCancelledDialogsChangeNothing(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	chooser.save, chooser.open = "", ""
	if path, err := app.Export(); path != "" || err != nil {
		t.Errorf("a cancelled save = %q, %v", path, err)
	}
	result, err := app.Import()
	if err != nil || result.Chosen {
		t.Errorf("a cancelled open = %+v, %v", result, err)
	}
	chooser.err = errors.New("the dialog failed")
	if _, err := app.Export(); err == nil {
		t.Error("a failed save dialog was not reported")
	}
	if _, err := app.Import(); err == nil {
		t.Error("a failed open dialog was not reported")
	}
}

func TestEveryActionSaysWhyWhenTheRecordCannotBeRead(t *testing.T) {
	t.Parallel()
	reason := errors.New("the record could not be read")
	app, _, _ := facadeOver(t, store.Unavailable{Reason: reason}, t.TempDir())
	calls := map[string]func() error{
		"record": func() error { _, err := app.Record(RecordFormDTO{Symptom: "Tired"}); return err },
		"history": func() error {
			_, err := app.History(FilterDTO{})
			return err
		},
		"suggest": func() error { _, err := app.Suggest("t"); return err },
		"symptoms": func() error {
			_, err := app.Symptoms()
			return err
		},
		"edit":   func() error { _, err := app.Edit(1, EditFormDTO{Symptom: "Tired"}); return err },
		"prompt": func() error { _, err := app.DeletionPrompt([]int64{1}); return err },
		"delete": func() error { return app.Delete([]int64{1}) },
		"rename": func() error { return app.Rename(1, "Tired") },
		"receipt": func() error {
			_, err := app.Receipt("2026-09-01", "2026-09-30", "")
			return err
		},
		"export": func() error { _, err := app.Export(); return err },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, reason) {
			t.Errorf("%s = %v, want the reason", name, err)
		}
	}
}

func TestAPanicInABoundMethodBecomesAnError(t *testing.T) {
	t.Parallel()
	var err error
	func() {
		defer guard(&err)
		panic("a planted panic")
	}()
	if !errors.Is(err, errInternal) {
		t.Errorf("a panic answered %v, want errInternal", err)
	}
}
