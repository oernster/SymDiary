package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/symdiary/internal/domain"
)

// aTiredEvent puts one event in the record so there is something to write.
func aTiredEvent(t *testing.T, app *App) {
	t.Helper()
	record(t, app, RecordFormDTO{Symptom: "Tired", Note: "Only been awake for about 10 minutes."})
}

func TestSavePDFWritesTheRecordWhereTheReaderChose(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = filepath.Join(t.TempDir(), "record.pdf")

	path, err := app.SavePDF("2026-08-23", "2026-09-22", "")
	if err != nil {
		t.Fatalf("SavePDF: %v", err)
	}

	if path != chooser.save {
		t.Errorf("SavePDF answered %q, want the path the reader chose, %q", path, chooser.save)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading what was written: %v", err)
	}
	// A reader takes this to an appointment. It has to be a document a machine
	// they have never met will open, so the check is on the file itself.
	if !strings.HasPrefix(string(written), "%PDF-") {
		t.Errorf("what was written begins %q, which is not a PDF", string(written[:min(8, len(written))]))
	}
}

func TestSavePDFSaysSoWhenTheDocumentCannotBeWritten(t *testing.T) {
	t.Parallel()
	// A folder that is really a file: the reader chose somewhere nothing can be
	// written, so the save must be refused rather than reported as done.
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	blocker := filepath.Join(t.TempDir(), "not-a-folder")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatalf("making the blocking file: %v", err)
	}
	chooser.save = filepath.Join(blocker, "record.pdf")

	path, err := app.SavePDF("2026-08-23", "2026-09-22", "")
	if err == nil || path != "" {
		t.Errorf("SavePDF into a file = %q, %v; want no path and the reason", path, err)
	}
}

func TestTheSaveDialogAsksForTheKindOfFileItIsActuallyWriting(t *testing.T) {
	t.Parallel()
	// macOS applies the dialog's filter to the name it is given. A document
	// offered under the export filter was saved as `record.pdf.json`: a file
	// whose name said one thing, whose contents said another and which no
	// viewer would open. Found on macOS 2026-09-23.
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = filepath.Join(t.TempDir(), "record.pdf")

	if _, err := app.SavePDF("2026-08-23", "2026-09-22", ""); err != nil {
		t.Fatalf("SavePDF: %v", err)
	}
	if chooser.asked.extension != "pdf" {
		t.Errorf("the save dialog was opened for a .%s file while writing a PDF",
			chooser.asked.extension)
	}
	assertNameMatchesKind(t, chooser)

	chooser.save = filepath.Join(t.TempDir(), "record.json")
	if _, err := app.Export(); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if chooser.asked.extension != "json" {
		t.Errorf("the save dialog was opened for a .%s file while writing an export",
			chooser.asked.extension)
	}
	assertNameMatchesKind(t, chooser)
}

// assertNameMatchesKind states that the name a dialog suggests ends in the
// extension that dialog filters to. Where they disagree, macOS keeps both.
func assertNameMatchesKind(t *testing.T, chooser *fakeChooser) {
	t.Helper()
	if want := "." + chooser.asked.extension; !strings.HasSuffix(chooser.suggested, want) {
		t.Errorf("the dialog suggested %q under a %s filter, so the two would be joined",
			chooser.suggested, want)
	}
}

// TestEveryKindOfFileSaysWhatItIs holds the two kinds apart in every respect a
// reader sees, not only the extension: a dialog headed "Export your record"
// over a symptom record tells them they are doing something else.
func TestEveryKindOfFileSaysWhatItIs(t *testing.T) {
	t.Parallel()
	for _, kind := range []fileKind{exportKind, documentKind} {
		if kind.title == "" || kind.describes == "" || kind.extension == "" {
			t.Errorf("%+v leaves a dialog with nothing to say", kind)
		}
		if strings.HasPrefix(kind.extension, ".") {
			t.Errorf("%q carries its own dot, which the filter adds", kind.extension)
		}
	}
	if exportKind.extension == documentKind.extension {
		t.Error("both kinds filter to the same extension, so one of them is wrong")
	}
	if exportKind.title == documentKind.title {
		t.Error("both dialogs are headed the same, so one of them names the wrong act")
	}
}

func TestSavePDFWritesNothingWhenTheReaderCancels(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = ""

	path, err := app.SavePDF("2026-08-23", "2026-09-22", "")

	if err != nil {
		t.Errorf("a cancelled dialog answered %v, want no error: cancelling is not a fault", err)
	}
	if path != "" {
		t.Errorf("a cancelled dialog answered %q, want nothing written", path)
	}
}

func TestSavePDFRefusesAnEmptyRangeBeforeAskingWhereToSave(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	unreachable := filepath.Join(t.TempDir(), "should-not-exist.pdf")
	chooser.save = unreachable

	_, err := app.SavePDF("2026-01-01", "2026-01-31", "")

	if !errors.Is(err, domain.ErrEmptyRange) {
		t.Errorf("SavePDF over an empty range = %v, want ErrEmptyRange", err)
	}
	// Asking where to save and then saying there is nothing to save wastes the
	// reader's answer, so the refusal comes first.
	if _, statErr := os.Stat(unreachable); statErr == nil {
		t.Error("a file was written for a range holding no events")
	}
}

func TestSavePDFRefusesADateItCannotRead(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	aTiredEvent(t, app)

	if _, err := app.SavePDF("not-a-date", "2026-09-22", ""); err == nil {
		t.Error("SavePDF with an unreadable start date answered no error")
	}
	if _, err := app.SavePDF("2026-08-23", "not-a-date", ""); err == nil {
		t.Error("SavePDF with an unreadable end date answered no error")
	}
}

func TestSavePDFWithNoWriterIsAnInternalFault(t *testing.T) {
	t.Parallel()
	// A facade built without a sheet is a wiring mistake rather than a state the
	// reader can reach, so it answers the same internal fault every other bound
	// method answers rather than ending the window.
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = filepath.Join(t.TempDir(), "record.pdf")
	app.sheet = nil

	if _, err := app.SavePDF("2026-08-23", "2026-09-22", ""); err == nil {
		t.Error("SavePDF with no sheet answered no error, want the internal fault")
	}
}
