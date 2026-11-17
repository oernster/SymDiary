package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
)

func TestDownloadsIsWhereTheDialogsOpen(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	if got := downloads(); got != "" {
		t.Errorf("with no Downloads folder = %q, want the empty string", got)
	}
	folder := filepath.Join(home, downloadsName)
	if err := os.Mkdir(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := downloads(); got != folder {
		t.Errorf("downloads = %q, want %q", got, folder)
	}
	file := filepath.Join(home, "elsewhere")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USERPROFILE", file)
	if got := downloads(); got != "" {
		t.Errorf("a Downloads that is not a folder = %q, want the empty string", got)
	}
}

func TestTheImportDialogListsTextAndJSON(t *testing.T) {
	t.Parallel()
	// One filter carrying both, so the phone's .txt sends are listed without
	// the user finding a second filter (FR-098).
	filters := importFilters()
	if len(filters) != 1 {
		t.Fatalf("import offers %d filters, want 1", len(filters))
	}
	got := strings.Split(filters[0].Pattern, ";")
	want := []string{"*." + exportKind.extension, "*." + application.ParcelExtension}
	if !slices.Equal(got, want) {
		t.Errorf("the import pattern lists %v, want %v", got, want)
	}
	if !strings.Contains(filters[0].DisplayName, filters[0].Pattern) {
		t.Errorf("the filter %q does not show the user what it lists", filters[0].DisplayName)
	}
}

func TestTheSystemClockReadsTheWallClock(t *testing.T) {
	t.Parallel()
	before := time.Now()
	read := systemClock{}.Now()
	if read.Before(before) || time.Since(read) > time.Minute {
		t.Errorf("the clock read %v, which is not now", read)
	}
}

func TestFocusIsBestEffortAndReturnsAtOnce(t *testing.T) {
	t.Parallel()
	// Focus does its work on a goroutine of its own, so nothing about opening
	// the window waits for the quarter second it sleeps first. What is
	// reachable without a real window is that the call returns and that it
	// asks for nothing of the facade; the Win32 half is covered by a person
	// opening the built window and pressing Tab.
	windowFocus{}.Focus()
}

// countingFocuser stands in for the window, counting the times it was asked
// for the keyboard rather than showing anything.
type countingFocuser struct{ asked int }

// Focus counts the request.
func (f *countingFocuser) Focus() { f.asked++ }

func TestReadyAsksTheWindowForTheKeyboard(t *testing.T) {
	t.Parallel()
	// WebView2 gives the page DOM focus without giving the webview the
	// keyboard, so every Tab reached nothing until the page was clicked. The
	// page cannot ask for this itself; the window has to.
	app, _, _ := window(t)
	focuser := &countingFocuser{}
	app.focuser = focuser

	app.ready(t.Context())
	if focuser.asked != 1 {
		t.Errorf("the window was asked for the keyboard %d times, want exactly 1", focuser.asked)
	}
}

func TestReadyWithNoFocuserStillOpens(t *testing.T) {
	t.Parallel()
	// A facade built without a focuser is a wiring mistake, not a state the
	// user can reach. It costs the keyboard, which is not worth ending a run
	// over: the record is still readable with a mouse.
	app, _, _ := window(t)
	app.focuser = nil
	app.ready(t.Context())
}

func TestStartupAndShutdown(t *testing.T) {
	t.Parallel()
	closed := false
	app, _, _ := window(t)
	app.close = func() error {
		closed = true
		return errors.New("the record would not close")
	}
	app.startup(t.Context())
	if app.ctx == nil {
		t.Error("startup did not keep the context")
	}
	app.shutdown(t.Context())
	if !closed {
		t.Error("shutdown did not close the record")
	}
}
