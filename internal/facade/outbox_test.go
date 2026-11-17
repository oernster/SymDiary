package facade

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/infrastructure/export"
	"github.com/oernster/symdiary/internal/infrastructure/memstore"
)

// memorySnapshots keeps the phone's record in memory for the test.
type memorySnapshots struct{ saved []byte }

func (m *memorySnapshots) Load() ([]byte, error)  { return m.saved, nil }
func (m *memorySnapshots) Save(data []byte) error { m.saved = data; return nil }

// phone answers the phone's shared facade and its outbox over one record.
func phone(t *testing.T) (*Core, *Outbox) {
	t.Helper()
	record, err := memstore.Open(&memorySnapshots{})
	if err != nil {
		t.Fatalf("opening the phone's record: %v", err)
	}
	core, clock := facadeOver(t, record)
	zone := london(t)
	outbox := application.NewOutbox(record, record, export.File{}, clock, zone)
	return core, NewOutbox(outbox, zone)
}

func TestThePhoneSendsWhatItRecorded(t *testing.T) {
	t.Parallel()
	core, outbox := phone(t)
	waiting, err := outbox.Waiting()
	if err != nil || waiting != (WaitingDTO{}) {
		t.Errorf("before recording = %+v, %v; want nothing waiting", waiting, err)
	}
	saved := record(t, core, RecordFormDTO{Symptom: "Tired"})

	waiting, err = outbox.Waiting()
	if err != nil || waiting.Count != 1 || waiting.Oldest != "22 Sep 2026 17:12" || waiting.HasBatch {
		t.Errorf("waiting = %+v, %v", waiting, err)
	}
	parcel, err := outbox.Prepare()
	if err != nil || parcel.Name != "SymDiary-phone-2026-09-22-1712.txt" ||
		len(parcel.Events) != 1 || parcel.Events[0] != saved.ID {
		t.Fatalf("parcel = %+v, %v", parcel, err)
	}
	var file struct {
		Format string `json:"format"`
		Events []struct {
			Symptom string `json:"symptom"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(parcel.Text), &file); err != nil ||
		file.Format != "symdiary-record" || len(file.Events) != 1 || file.Events[0].Symptom != "Tired" {
		t.Errorf("the parcel's text = %s, %v", parcel.Text, err)
	}

	if err := outbox.Confirm(parcel.Events); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if _, err := outbox.Prepare(); !errors.Is(err, application.ErrNothingToSend) {
		t.Errorf("Prepare once everything has gone = %v", err)
	}
	again, err := outbox.Again()
	if err != nil || again.Name != parcel.Name || again.Text != parcel.Text {
		t.Errorf("Again = %+v, %v; want the same parcel", again, err)
	}
}

func TestThePhoneSaysWhyWhenItCannotSend(t *testing.T) {
	t.Parallel()
	_, outbox := phone(t)
	if _, err := outbox.Again(); !errors.Is(err, application.ErrNoBatch) {
		t.Errorf("Again before any send = %v", err)
	}
	if err := outbox.Confirm(nil); !errors.Is(err, application.ErrNothingToSend) {
		t.Errorf("Confirm of nothing = %v", err)
	}
	reason := errors.New("the phone's storage could not be read")
	broken := application.Unavailable{Reason: reason}
	failing := NewOutbox(application.NewOutbox(broken, broken, export.File{},
		&stoppedClock{}, london(t)), london(t))
	if _, err := failing.Waiting(); !errors.Is(err, reason) {
		t.Errorf("Waiting with the record unreadable = %v", err)
	}
}
