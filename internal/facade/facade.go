// Package facade converts between the page's shapes and the application's.
//
// Both windows call it: the desktop's, through Wails, and the phone recorder's,
// through WebAssembly (Amendment 21, D-10). It owns nothing but the conversion;
// every rule lives below it. What only one window can do (file dialogs, the
// browser, the share sheet) stays with that window's own composition root.
package facade

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/product"
)

// ErrInternal is what the page is told when a bound method panics. The stack
// goes to the log; the page gets a sentence it can show.
var ErrInternal = errors.New(product.Name + " hit an internal fault; the details are in its log")

// Guard turns a panic in a bound method into ErrInternal, writing the stack to
// the log. A bound method runs on a goroutine the window owns, so the recover
// has to sit in the method itself: every bound method defers it first.
func Guard(err *error) {
	if recovered := recover(); recovered != nil {
		fmt.Fprintf(os.Stderr, "panic in a bound method: %v\n%s", recovered, debug.Stack())
		*err = ErrInternal
	}
}

// Services is what the facade drives: one application service per concern.
type Services struct {
	Recorder application.Recorder
	Editor   application.Editor
	History  application.History
	Transfer application.Transfer
}

// Core is the part of the facade every window shares.
type Core struct {
	zone     *time.Location
	clock    application.Clock
	version  string
	problem  string
	services Services
}

// New answers the shared facade. problem is the reason the record could not be
// opened, empty when it opened; the page shows it (FR-062).
func New(services Services, clock application.Clock, zone *time.Location,
	version, problem string) *Core {
	return &Core{
		zone: zone, clock: clock, version: version, problem: problem, services: services,
	}
}

// State answers what the page needs to start.
func (c *Core) State() (state StateDTO, err error) {
	defer Guard(&err)
	severities := []string{}
	for _, severity := range domain.Severities() {
		severities = append(severities, severity.String())
	}
	return StateDTO{
		Name: product.Name, Version: c.version, Problem: c.problem, Severities: severities,
	}, nil
}

// Now answers the current local time in the form the time field takes.
func (c *Core) Now() (now string, err error) {
	defer Guard(&err)
	return domain.FormatLocal(c.clock.Now(), c.zone), nil
}

// About answers the About dialog.
func (c *Core) About() (about AboutDTO, err error) {
	defer Guard(&err)
	return aboutDTO(application.NewAbout(c.version)), nil
}

// occurrence reads an occurrence time from the page; nil when none was given.
func (c *Core) occurrence(text string) (*time.Time, error) {
	if text == "" {
		return nil, nil
	}
	parsed, err := domain.ParseLocal(text, c.zone)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// optionalDate reads a date from the page; the zero date when none was given.
func optionalDate(text string) (domain.Date, error) {
	if text == "" {
		return domain.Date{}, nil
	}
	return domain.ParseDate(text)
}

// eventIDs converts ids from the page.
func eventIDs(ids []int64) []domain.EventID {
	out := make([]domain.EventID, 0, len(ids))
	for _, id := range ids {
		out = append(out, domain.EventID(id))
	}
	return out
}
