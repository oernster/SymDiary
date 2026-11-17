package facade

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/memstore"
)

// stoppedClock is a clock the test moves by hand.
type stoppedClock struct{ now time.Time }

func (c *stoppedClock) Now() time.Time { return c.now }

// london is the zone the facade works in.
func london(t *testing.T) *time.Location {
	t.Helper()
	zone, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("loading Europe/London: %v", err)
	}
	return zone
}

// facade answers a Core over a real record held in memory. It is the phone's
// store, which keeps the same contract as the desktop's (storetest) and builds
// for every platform, so these tests vet for WebAssembly too.
func facade(t *testing.T) (*Core, *stoppedClock) {
	t.Helper()
	opened, err := memstore.Open(&memorySnapshots{})
	if err != nil {
		t.Fatalf("opening the record: %v", err)
	}
	return facadeOver(t, opened)
}

// facadeOver answers a Core over any store.
func facadeOver(t *testing.T, record application.Store) (*Core, *stoppedClock) {
	t.Helper()
	zone := london(t)
	clock := &stoppedClock{now: time.Date(2026, time.September, 22, 17, 12, 0, 0, zone)}
	services := Services{
		Recorder: application.NewRecorder(record, clock),
		Editor:   application.NewEditor(record, clock, zone),
		History:  application.NewHistory(record, zone),
		Transfer: application.NewTransfer(record, export.File{}),
	}
	return New(services, clock, zone, "1.2.3", "", application.PhoneCredits()), clock
}

// record records one symptom, failing the test where it is refused.
func record(t *testing.T, core *Core, form RecordFormDTO) EventDTO {
	t.Helper()
	saved, err := core.Record(form)
	if err != nil {
		t.Fatalf("Record(%+v): %v", form, err)
	}
	return saved
}

func TestStateAndAbout(t *testing.T) {
	t.Parallel()
	app, _ := facade(t)
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
	if err != nil || about.Version != "1.2.3" || len(about.Credits) != len(application.PhoneCredits()) {
		t.Errorf("about = %+v, %v; want the credits it was built with", about, err)
	}
	now, err := app.Now()
	if err != nil || now != "2026-09-22T17:12" {
		t.Errorf("Now = %q, %v", now, err)
	}
}

func TestRecordThenHistory(t *testing.T) {
	t.Parallel()
	app, clock := facade(t)
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

func TestHistoryFiltersBySymptom(t *testing.T) {
	t.Parallel()
	app, _ := facade(t)
	record(t, app, RecordFormDTO{Symptom: "Tired"})
	headache := record(t, app, RecordFormDTO{Symptom: "Headache"})

	events, err := app.History(FilterDTO{Definitions: []int64{headache.Definition}})
	if err != nil || len(events) != 1 || events[0].Symptom != "Headache" {
		t.Errorf("filtered by Headache = %+v, %v; want the one Headache event", events, err)
	}
}

func TestRecordRefusals(t *testing.T) {
	t.Parallel()
	app, _ := facade(t)
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
	app, clock := facade(t)
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
	app, _ := facade(t)
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
	app, _ := facade(t)
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

func TestEveryActionSaysWhyWhenTheRecordCannotBeRead(t *testing.T) {
	t.Parallel()
	reason := errors.New("the record could not be read")
	app, _ := facadeOver(t, application.Unavailable{Reason: reason})
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
		defer Guard(&err)
		panic("a planted panic")
	}()
	if !errors.Is(err, ErrInternal) {
		t.Errorf("a panic answered %v, want ErrInternal", err)
	}
}
