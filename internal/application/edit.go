package application

import (
	"fmt"
	"slices"
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// EditForm is what the user changed on an event.
type EditForm struct {
	Symptom string
	// OccurredAt is the new occurrence time; nil keeps the one held, so an edit
	// can never move it by accident (FR-024).
	OccurredAt *time.Time
	Severity   domain.Severity
	Note       string
}

// Editor changes and deletes events.
type Editor struct {
	store Store
	clock Clock
	zone  *time.Location
	sent  Sent
}

// NewEditor answers an editor over a store, a clock and the user's zone. It
// treats nothing as sent, which is the desktop's case; the phone recorder
// hands its own Sent in with WithSent.
func NewEditor(store Store, clock Clock, zone *time.Location) Editor {
	return Editor{store: store, clock: clock, zone: zone, sent: NothingSent{}}
}

// WithSent answers a copy of the editor that refuses to change an event once
// it has been sent (FR-088).
func (e Editor) WithSent(sent Sent) Editor {
	e.sent = sent
	return e
}

// refuseSent answers ErrAlreadySent when any of ids has been sent, so a sent
// event is never edited or deleted here (FR-088, D-3).
func (e Editor) refuseSent(ids ...domain.EventID) error {
	sent, err := e.sent.SentIDs()
	if err != nil {
		return because(ErrNotRead, err)
	}
	for _, id := range ids {
		if slices.Contains(sent, id) {
			return ErrAlreadySent
		}
	}
	return nil
}

// Edit replaces an event's symptom, severity and note, plus its occurrence time
// where the form carries one. Recorded-at is left alone (FR-007).
func (e Editor) Edit(id domain.EventID, form EditForm) (domain.Event, error) {
	if err := domain.CheckLabel(form.Symptom); err != nil {
		return domain.Event{}, err
	}
	if err := e.refuseSent(id); err != nil {
		return domain.Event{}, err
	}
	held, err := e.store.Event(id)
	if err != nil {
		return domain.Event{}, because(ErrNotChanged, err)
	}
	occurred := held.OccurredAt
	if form.OccurredAt != nil {
		occurred = *form.OccurredAt
		if err := domain.CheckOccurrence(occurred, e.clock.Now()); err != nil {
			return domain.Event{}, err
		}
	}
	saved, err := e.store.UpdateEvent(id, EventInput{
		Label:      form.Symptom,
		Key:        domain.KeyOf(form.Symptom),
		OccurredAt: occurred,
		RecordedAt: held.RecordedAt,
		Severity:   form.Severity,
		Note:       form.Note,
	})
	if err != nil {
		return domain.Event{}, because(ErrNotChanged, err)
	}
	return saved, nil
}

// The words of the delete confirmation (FR-025, FR-026).
const (
	deleteOneLayout  = "Delete the %s event of %s? This cannot be undone."
	deleteManyLayout = "Delete %d events? This cannot be undone."
)

// DeletionPrompt answers the confirmation to show before deleting: one event is
// named by its symptom and time; several are counted.
func (e Editor) DeletionPrompt(ids []domain.EventID) (string, error) {
	if len(ids) == 0 {
		return "", ErrNothingChosen
	}
	if err := e.refuseSent(ids...); err != nil {
		return "", err
	}
	switch len(ids) {
	case 1:
		held, err := e.store.Event(ids[0])
		if err != nil {
			return "", because(ErrNotRead, err)
		}
		return fmt.Sprintf(deleteOneLayout,
			held.Symptom, domain.FormatDisplay(held.OccurredAt, e.zone)), nil
	default:
		return fmt.Sprintf(deleteManyLayout, len(ids)), nil
	}
}

// Delete removes the events. The window asks first, showing DeletionPrompt.
func (e Editor) Delete(ids []domain.EventID) error {
	if len(ids) == 0 {
		return ErrNothingChosen
	}
	if err := e.refuseSent(ids...); err != nil {
		return err
	}
	if err := e.store.DeleteEvents(ids); err != nil {
		return because(ErrNotDeleted, err)
	}
	return nil
}
