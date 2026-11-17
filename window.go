package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/oernster/symdiary/internal/infrastructure/windowfocus"
	"github.com/oernster/symdiary/internal/product"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// filters limits a dialog to one kind of file.
func (k fileKind) filters() []runtime.FileFilter {
	return []runtime.FileFilter{{
		DisplayName: fmt.Sprintf("%s (*.%s)", k.describes, k.extension),
		Pattern:     "*." + k.extension,
	}}
}

// importFilters lists every file Import reads (importExtensions). One filter
// joins the patterns, so neither kind is hidden behind a choice the user has to
// find; Wails splits a pattern on ";" on all three desktops.
func importFilters() []runtime.FileFilter {
	globs := make([]string, len(importExtensions))
	for i, extension := range importExtensions {
		globs[i] = "*." + extension
	}
	patterns := strings.Join(globs, ";")
	return []runtime.FileFilter{{
		DisplayName: fmt.Sprintf("%s (%s)", exportKind.describes, patterns),
		Pattern:     patterns,
	}}
}

// downloadsName is the folder both file dialogs open in, where a person
// already looks for files they have saved. The user is free to go elsewhere;
// Windows then remembers where they went.
const downloadsName = "Downloads"

// downloads answers the user's Downloads folder, else the empty string where
// there is none, which leaves the dialog to choose for itself.
func downloads() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	folder := filepath.Join(home, downloadsName)
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return ""
	}
	return folder
}

// windowChooser asks through the window's own file dialogs.
type windowChooser struct{ app *App }

// SavePath opens the save dialog with a suggested name, for one kind of file.
func (c windowChooser) SavePath(suggested string, kind fileKind) (string, error) {
	return runtime.SaveFileDialog(c.app.ctx, runtime.SaveDialogOptions{
		Title: kind.title, DefaultFilename: suggested,
		DefaultDirectory: downloads(), Filters: kind.filters(),
	})
}

// OpenPath opens the open dialog.
func (c windowChooser) OpenPath() (string, error) {
	return runtime.OpenFileDialog(c.app.ctx, runtime.OpenDialogOptions{
		Title: "Import a record", DefaultDirectory: downloads(),
		Filters: importFilters(),
	})
}

// windowOpener asks the desktop to open an address, through the window's own
// runtime. Wails answers nothing, so a refusal by the desktop is invisible
// here; there is no outcome to report and none is invented.
type windowOpener struct{ app *App }

// Open hands the address to whatever opens links on this machine.
func (o windowOpener) Open(address string) {
	runtime.BrowserOpenURL(o.app.ctx, address)
}

// settleBeforeFocus lets the window finish showing before the keyboard is
// handed over, so the focus change lands on a settled window. PigeonPost's
// measured value, taken with the rest of this.
const settleBeforeFocus = 250 * time.Millisecond

// windowFocus hands the keyboard to the page.
type windowFocus struct{}

// Focus gives the WebView2 control the keyboard, a little after the window has
// opened.
//
// Ported from PigeonPost, where it was measured and where it works. Asking the
// Wails runtime to show the window is NOT enough, which is how this was got
// wrong the first time: that focuses the MAIN window, while WebView2 hosts the
// page in a child window of its own and the keys follow the child.
//
// The wait happens on a goroutine of its own, so nothing about opening the
// window waits for it. A panic there would end the process with the window
// already up, so it is recovered and reported to the log.
func (windowFocus) Focus() {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				fmt.Fprintf(os.Stderr, "handing over the keyboard: %v\n%s",
					recovered, debug.Stack())
			}
		}()
		time.Sleep(settleBeforeFocus)
		windowfocus.GiveTheKeyboardToThePage(product.Name)
	}()
}

// instanceID names SymDiary's single-instance lock, per user (FR-064).
const instanceID = "uk.codecrafter.symdiary"

// singleInstance brings the running window forward when SymDiary is started
// again; the second copy then ends.
func singleInstance(app *App) *options.SingleInstanceLock {
	return &options.SingleInstanceLock{
		UniqueId: instanceID,
		OnSecondInstanceLaunch: func(options.SecondInstanceData) {
			runtime.WindowUnminimise(app.ctx)
			runtime.Show(app.ctx)
		},
	}
}
