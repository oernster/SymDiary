//go:build js && wasm

// Command recorder is the phone recorder's composition root (Amendment 21):
// the only file in the WebAssembly build permitted to wire concrete
// infrastructure to the application layer.
//
// It runs SymDiary's own rules (D-10) over the phone's snapshot record, then
// installs the facade where the page already looks for the desktop's:
// window.go.main.App. The page cannot tell the two windows apart, so the
// recording form the desktop uses runs here unchanged.
package main

import (
	"fmt"
	"os"
	"syscall/js"
	"time"

	// The phone's browser offers Go no time-zone files, and without them
	// time.Local is a fixed offset from start-up, an hour out across summer
	// time (measured 2026-11-17). The tables travel in the binary instead.
	_ "time/tzdata"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/facade"
	"github.com/oernster/symdiary/internal/infrastructure/browserstore"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/memstore"
)

// appVersion is set from VERSION by build-recorder.ps1 through -ldflags -X,
// which only reaches a var.
var appVersion = "0.0.0-dev"

// zoneGlobal is where the page leaves the phone's time-zone name before it
// starts Go, read from Intl in the browser.
const zoneGlobal = "symdiaryZone"

// readyGlobal is the function the page leaves for Go to call once the facade
// is installed, so the form's Record button can wake (FR-100).
const readyGlobal = "symdiaryReady"

// systemClock is the real clock.
type systemClock struct{}

// Now answers the current instant.
func (systemClock) Now() time.Time { return time.Now() }

// record is what the phone's store is: the record and what has been sent.
type record interface {
	application.Store
	application.Sent
}

func main() {
	zone := phoneZone()
	clock := systemClock{}

	var held record
	problem := ""
	snapshots, err := browserstore.Open()
	if err == nil {
		var opened *memstore.Store
		if opened, err = memstore.Open(snapshots); err == nil {
			held = opened
		}
	}
	if err != nil {
		// The page opens regardless and says what went wrong (house rule 1).
		fmt.Fprintf(os.Stderr, "opening the record: %v\n", err)
		problem = fmt.Sprintf(facade.ProblemLayout, err)
		held = application.Unavailable{Reason: err}
	}

	services := facade.Services{
		Recorder: application.NewRecorder(held, clock),
		Editor:   application.NewEditor(held, clock, zone).WithSent(held),
		History:  application.NewHistory(held, zone),
		Transfer: application.NewTransfer(held, export.File{}),
	}
	core := facade.New(services, clock, zone, appVersion, problem)
	outbox := facade.NewOutbox(application.NewOutbox(held, held, export.File{}, clock, zone), zone)
	install(core, outbox)

	if ready := js.Global().Get(readyGlobal); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {}
}

// phoneZone answers the zone the page named; the fixed start-up offset, said
// in the log, when it named none Go knows.
func phoneZone() *time.Location {
	name := js.Global().Get(zoneGlobal)
	if name.Type() == js.TypeString {
		if zone, err := time.LoadLocation(name.String()); err == nil {
			return zone
		}
	}
	fmt.Fprintln(os.Stderr, "the page named no time zone Go knows, so times use the offset at start-up")
	return time.Local
}
