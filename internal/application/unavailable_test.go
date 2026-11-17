package application

import (
	"errors"
	"testing"
)

func TestUnavailableRefusesEverythingWithTheReason(t *testing.T) {
	t.Parallel()
	reason := errors.New("the record could not be read")
	unavailable := Unavailable{Reason: reason}
	var store Store = unavailable
	var sent Sent = unavailable
	_, e1 := store.Definitions()
	_, e2 := store.Events()
	_, e3 := store.Event(1)
	_, e4 := store.AddEvent(EventInput{})
	_, e5 := store.UpdateEvent(1, EventInput{})
	e6 := store.DeleteEvents(nil)
	e7 := store.RenameDefinition(1, "", "")
	e8 := store.Import(nil, nil)
	_, e9 := sent.SentIDs()
	_, e10 := sent.LastBatch()
	e11 := sent.MarkSent(nil)
	for i, err := range []error{e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11} {
		if !errors.Is(err, reason) {
			t.Errorf("operation %d answered %v, want the reason", i+1, err)
		}
	}
}
