package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/facade"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/pdf"
	"github.com/oernster/symdiary/internal/infrastructure/store"
)

// The window's own methods: the ones that ask the desktop for a file. The
// methods every window shares are tested with internal/facade.

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

// window answers an App over a real record in a temporary folder.
func window(t *testing.T) (*App, *stoppedClock, *fakeChooser) {
	t.Helper()
	folder := t.TempDir()
	opened, err := store.Open(filepath.Join(folder, "symdiary.db"))
	if err != nil {
		t.Fatalf("opening the record: %v", err)
	}
	t.Cleanup(func() { _ = opened.Close() })
	return windowOver(t, opened, folder)
}

// windowOver answers an App over any store.
func windowOver(t *testing.T, record application.Store, folder string) (*App, *stoppedClock, *fakeChooser) {
	t.Helper()
	zone := london(t)
	clock := &stoppedClock{now: time.Date(2026, time.September, 22, 17, 12, 0, 0, zone)}
	chooser := &fakeChooser{save: filepath.Join(folder, "export.json")}
	services := facadeServices(record, clock, zone)
	app := newApp(services, clock, zone, "1.2.3", "", chooser, func() error { return nil })
	// The real renderer, writing into the test's own folder. A double would
	// assert that the double was called; this asserts that a file arrives.
	app.sheet = pdf.Sheet{}
	return app, clock, chooser
}

// facadeServices builds the services the composition root builds.
func facadeServices(record application.Store, clock application.Clock, zone *time.Location) facade.Services {
	return facade.Services{
		Recorder: application.NewRecorder(record, clock),
		Editor:   application.NewEditor(record, clock, zone),
		History:  application.NewHistory(record, zone),
		Transfer: application.NewTransfer(record, export.File{}),
	}
}

// record records one symptom, failing the test where it is refused.
func record(t *testing.T, app *App, form facade.RecordFormDTO) facade.EventDTO {
	t.Helper()
	saved, err := app.Record(form)
	if err != nil {
		t.Fatalf("Record(%+v): %v", form, err)
	}
	return saved
}

func TestExportAndImport(t *testing.T) {
	t.Parallel()
	app, _, chooser := window(t)
	record(t, app, facade.RecordFormDTO{Symptom: "Tired", Severity: "Mild"})

	path, err := app.Export()
	if err != nil || path != chooser.save {
		t.Fatalf("Export = %q, %v", path, err)
	}

	empty, _, receiver := window(t)
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
	app, _, chooser := window(t)
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

func TestExportSaysWhyWhenTheRecordCannotBeRead(t *testing.T) {
	t.Parallel()
	reason := errors.New("the record could not be read")
	app, _, _ := windowOver(t, store.Unavailable{Reason: reason}, t.TempDir())
	if _, err := app.Export(); !errors.Is(err, reason) {
		t.Errorf("export = %v, want the reason", err)
	}
}
