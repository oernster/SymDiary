package application

import (
	"errors"
	"fmt"
)

// The refusals the application answers when the store or a file fails. Each is
// wrapped round the underlying reason, so the message names what was not done
// and why (FR-009, FR-027, FR-052).
var (
	ErrNotSaved      = errors.New("the event was not saved")
	ErrNotChanged    = errors.New("the change was not saved")
	ErrNotDeleted    = errors.New("nothing was deleted")
	ErrNotRenamed    = errors.New("the symptom was not renamed")
	ErrNotRead       = errors.New("the record could not be read")
	ErrNotExported   = errors.New("the export was not written")
	ErrNotImported   = errors.New("nothing was imported")
	ErrNothingChosen = errors.New("no events were chosen")
)

// The phone recorder's refusals (Amendment 21).
var (
	// ErrAlreadySent refuses a change to an event the desktop already holds
	// (FR-088): the desktop owns its corrections from then on (D-3). It is
	// never wrapped, so it is shown whole and ends as a sentence.
	//
	//lint:ignore ST1005 shown to the user verbatim as a sentence
	ErrAlreadySent   = errors.New("this event has been sent to the desktop, so it is corrected there.")
	ErrNothingToSend = errors.New("there is nothing waiting to be sent")
	ErrNoBatch       = errors.New("nothing has been sent yet, so there is nothing to send again")
	ErrNotSent       = errors.New("the send was not recorded")
)

// because wraps a failure in the refusal it caused, keeping both inspectable
// with errors.Is.
func because(refusal, reason error) error {
	return fmt.Errorf("%w: %w", refusal, reason)
}
