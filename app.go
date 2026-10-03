package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/product"
)

// errInternal is what the page is told when a bound method panics. The stack
// goes to the log; the page gets a sentence it can show.
var errInternal = errors.New(product.Name + " hit an internal fault; the details are in its log")

// fileKind names what a file dialog is for: the words in its title, how the
// kind of file is described to the user and the extension it filters to.
//
// The extension is not decoration. macOS applies the dialog's filter to the
// name it is given, so a PDF offered under an export filter is saved as
// `record.pdf.json`: a file whose name says one thing, whose contents say
// another and which no viewer will open. Found on macOS 2026-09-23, when both
// dialogs still shared the one export filter.
type fileKind struct {
	title     string
	describes string
	extension string
}

// The two kinds of file SymDiary writes.
var (
	exportKind = fileKind{
		title:     "Export your record",
		describes: product.Name + " record",
		extension: "json",
	}
	documentKind = fileKind{
		title:     "Save your symptom record",
		describes: product.Name + " symptom record",
		extension: "pdf",
	}
)

// fileChooser asks the user where to write or read a file. An empty path with
// no error means the user cancelled.
type fileChooser interface {
	SavePath(suggested string, kind fileKind) (string, error)
	OpenPath() (string, error)
}

// windowFocuser hands the window's keyboard to the page.
//
// WebView2 keeps DOM focus and keyboard focus apart: the page can hold the
// first while the webview holds none of the second; a key pressed then reaches
// no listener at all. Measured 2026-09-22 in the built window, where
// no Tab ever stepped the ring until the page had been clicked once. Showing
// the main window is not enough either: WebView2 hosts the page in a child
// window of its own and the keys follow the child, so the child is what has
// to be focused.
type windowFocuser interface {
	Focus()
}

// browserOpener hands an address to whatever the desktop opens links with.
// Wails' own opener answers nothing, so neither does this: a desktop that
// declines to open a browser tells the application nothing it could report.
type browserOpener interface {
	Open(address string)
}

// recordSheet writes a record out as a document the reader keeps, answering how
// many pages it came to.
//
// SymDiary used to hand the record to the browser's own print path instead. That
// path belongs to three different engines: what came off the paper depended on
// which desktop the reader was using, on whether their print dialog had
// "Headers and footers" ticked and on which of the CSS the sheet leaned on
// their engine had implemented. Measured on Windows, Linux and macOS: three
// different answers. The record is the product, so it is drawn once, here
// (FR-040, Amendment 16).
type recordSheet interface {
	// Check refuses a record the document cannot print faithfully, so the reader
	// is told before being asked where to save it.
	Check(lines []domain.Line) error
	Write(path string, lines []domain.Line) (int, error)
}

// Services is what the facade drives: one application service per concern.
type Services struct {
	Recorder application.Recorder
	Editor   application.Editor
	History  application.History
	Transfer application.Transfer
}

// App is the facade the window calls. It converts between the page's shapes
// and the application's; it owns nothing else, since every rule lives below it.
type App struct {
	ctx      context.Context
	zone     *time.Location
	clock    application.Clock
	version  string
	problem  string
	services Services
	chooser  fileChooser
	opener   browserOpener
	sheet    recordSheet
	focuser  windowFocuser
	close    func() error
}

// newApp answers the facade. problem is the reason the record could not be
// opened, empty when it opened; the page shows it (FR-062).
func newApp(services Services, clock application.Clock, zone *time.Location,
	version, problem string, chooser fileChooser, close func() error) *App {
	return &App{
		zone: zone, clock: clock, version: version, problem: problem,
		services: services, chooser: chooser, close: close,
	}
}

// startup keeps the window's context for the dialogs.
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// ready is called once the page exists. It asks the window for the keyboard,
// which the page cannot ask for itself: a DOM focus call sets which element
// would receive a key without making the webview the thing keys are sent to
// (NFR-USE-002).
//
// A window with no focuser keeps running and says so in the log. The keyboard
// is not worth ending a run over; the record is still readable with a mouse.
func (a *App) ready(context.Context) {
	if a.focuser == nil {
		fmt.Fprintln(os.Stderr, "the window has no focuser, so the page starts without the keyboard")
		return
	}
	a.focuser.Focus()
}

// shutdown closes the record.
func (a *App) shutdown(context.Context) {
	if err := a.close(); err != nil {
		fmt.Fprintf(os.Stderr, "closing the record: %v\n", err)
	}
}

// guard turns a panic in a bound method into errInternal, writing the stack to
// the log. A bound method runs on a goroutine Wails owns, so the recover has to
// sit here, in the method itself.
func guard(err *error) {
	if recovered := recover(); recovered != nil {
		fmt.Fprintf(os.Stderr, "panic in a bound method: %v\n%s", recovered, debug.Stack())
		*err = errInternal
	}
}

// State answers what the page needs to start.
func (a *App) State() (state StateDTO, err error) {
	defer guard(&err)
	severities := []string{}
	for _, severity := range domain.Severities() {
		severities = append(severities, severity.String())
	}
	return StateDTO{
		Name: product.Name, Version: a.version, Problem: a.problem, Severities: severities,
	}, nil
}

// Now answers the current local time in the form the time field takes.
func (a *App) Now() (now string, err error) {
	defer guard(&err)
	return domain.FormatLocal(a.clock.Now(), a.zone), nil
}

// About answers the About dialog.
func (a *App) About() (about AboutDTO, err error) {
	defer guard(&err)
	return aboutDTO(application.NewAbout(a.version)), nil
}

// occurrence reads an occurrence time from the page; nil when none was given.
func (a *App) occurrence(text string) (*time.Time, error) {
	if text == "" {
		return nil, nil
	}
	parsed, err := domain.ParseLocal(text, a.zone)
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

// eventDTOs converts events for the page.
func (a *App) eventDTOs(events []domain.Event) []EventDTO {
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, a.eventDTO(event))
	}
	return out
}
