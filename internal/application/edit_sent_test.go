package application

import (
	"testing"

	"github.com/oernster/symdiary/internal/domain"
)

// The phone recorder's guard (FR-088): once an event has been sent, the
// desktop owns its corrections, so the phone refuses to edit or delete it.

func TestASentEventIsNeitherEditedNorDeleted(t *testing.T) {
	t.Parallel()
	_, store, sent, ids := outboxOver(t)
	sent.sent = ids[:1]
	editor := NewEditor(store, fixedClock{at(t, 17, 12)}, london(t)).WithSent(sent)

	_, err := editor.Edit(ids[0], EditForm{Symptom: "Tired", Note: "changed"})
	wantIs(t, "editing a sent event", err, ErrAlreadySent)
	_, err = editor.DeletionPrompt(ids[:2])
	wantIs(t, "asking to delete a sent event among others", err, ErrAlreadySent)
	wantIs(t, "deleting a sent event", editor.Delete(ids[:1]), ErrAlreadySent)
	if held, _ := store.Event(ids[0]); held.Note != "" {
		t.Errorf("the refused edit reached the store: note %q", held.Note)
	}
}

func TestAnUnsentEventIsStillEditedAndDeleted(t *testing.T) {
	t.Parallel()
	_, store, sent, ids := outboxOver(t)
	sent.sent = ids[:1]
	editor := NewEditor(store, fixedClock{at(t, 17, 12)}, london(t)).WithSent(sent)

	edited, err := editor.Edit(ids[1], EditForm{Symptom: "Tired", Note: "changed"})
	if err != nil || edited.Note != "changed" {
		t.Errorf("editing an unsent event = %+v, %v", edited, err)
	}
	if _, err := editor.DeletionPrompt(ids[1:2]); err != nil {
		t.Errorf("asking to delete an unsent event: %v", err)
	}
	if err := editor.Delete(ids[1:2]); err != nil {
		t.Errorf("deleting an unsent event: %v", err)
	}
}

func TestTheGuardSaysWhyWhenItCannotReadWhatWasSent(t *testing.T) {
	t.Parallel()
	_, store, sent, ids := outboxOver(t)
	sent.failRead = true
	editor := NewEditor(store, fixedClock{at(t, 17, 12)}, london(t)).WithSent(sent)

	_, err := editor.Edit(ids[0], EditForm{Symptom: "Tired"})
	wantIs(t, "editing with sends unreadable", err, ErrNotRead)
	wantIs(t, "deleting with sends unreadable", editor.Delete(ids[:1]), ErrNotRead)
	_, err = editor.DeletionPrompt([]domain.EventID{ids[0]})
	wantIs(t, "asking with sends unreadable", err, ErrNotRead)
}
