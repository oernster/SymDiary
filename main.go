// Command symdiary records symptoms as they are noticed and prints a factual
// record to take to a doctor.
//
// This file is the composition root: the only file permitted to wire concrete
// infrastructure to the application layer.
package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/facade"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/pdf"
	"github.com/oernster/symdiary/internal/infrastructure/runlog"
	"github.com/oernster/symdiary/internal/infrastructure/store"
	"github.com/oernster/symdiary/internal/product"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// mark is the application's own icon, drawn beside the line naming the program
// on the printed record (FR-045). It is embedded from the one copy the window
// already ships rather than from a second file kept in step by hand.
//
//go:embed frontend/src/assets/icons/application-icon.png
var mark []byte

// appVersion is set from VERSION by build.ps1 through -ldflags -X, which only
// reaches a var.
var appVersion = "0.0.0-dev"

// Window geometry: wide enough for the history's columns, tall enough for the
// recording form without scrolling.
const (
	windowWidth     = 1000
	windowHeight    = 760
	windowMinWidth  = 720
	windowMinHeight = 560
)

// problemLayout words a record that could not be opened (FR-062).
const problemLayout = "Your record could not be opened, so nothing can be recorded " +
	"until this is put right. Nothing in the file has been changed. %v"

// systemClock is the real clock.
type systemClock struct{}

// Now answers the current instant.
func (systemClock) Now() time.Time { return time.Now() }

func main() {
	keepLog(time.Now())
	zone := time.Local

	var record application.Store
	closeRecord := func() error { return nil }
	problem := ""
	path, err := store.DefaultPath()
	if err == nil {
		var opened *store.Store
		if opened, err = store.Open(path); err == nil {
			record, closeRecord = opened, opened.Close
		}
	}
	if err != nil {
		// The window opens regardless and says what went wrong (house rule 1).
		fmt.Fprintf(os.Stderr, "opening the record: %v\n", err)
		problem = fmt.Sprintf(problemLayout, err)
		record = store.Unavailable{Reason: err}
	}

	clock := systemClock{}
	services := facade.Services{
		Recorder: application.NewRecorder(record, clock),
		Editor:   application.NewEditor(record, clock, zone),
		History:  application.NewHistory(record, zone),
		Transfer: application.NewTransfer(record, export.File{}),
	}
	app := newApp(services, clock, zone, appVersion, problem, nil, closeRecord)
	app.chooser = windowChooser{app: app}
	app.opener = windowOpener{app: app}
	app.sheet = pdf.Sheet{Mark: mark}
	app.focuser = windowFocus{}

	err = wails.Run(&options.App{
		Title:              product.Name,
		Width:              windowWidth,
		Height:             windowHeight,
		MinWidth:           windowMinWidth,
		MinHeight:          windowMinHeight,
		AssetServer:        &assetserver.Options{Assets: assets},
		OnStartup:          app.startup,
		OnDomReady:         app.ready,
		OnShutdown:         app.shutdown,
		SingleInstanceLock: singleInstance(app),
		Bind:               []interface{}{app},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "running the window: %v\n", err)
		os.Exit(1)
	}
}

// keepLog opens the run log and points the error output at it before anything
// else can fail (FR-065). A log that cannot be kept does not stop the run.
func keepLog(started time.Time) {
	path, err := runlog.Path()
	if err != nil {
		return
	}
	log, err := runlog.Open(path, started)
	if err != nil {
		return
	}
	if err := runlog.Keep(log); err != nil {
		fmt.Fprintf(log, "keeping the log: %v\n", err)
	}
}
