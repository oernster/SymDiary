package application

import (
	"slices"
	"time"

	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/product"
)

// ParcelExtension is the extension a send from the phone carries: text,
// because Chrome on Android refuses to share a .json file (M-4, D-9). The
// contents are export format version 1 all the same, which is why the
// desktop's Import lists it (FR-098).
const ParcelExtension = "txt"

// parcelNameLayout names a send in the phone's local time, so a reader can
// tell batches apart in a folder (FR-094).
const parcelNameLayout = "-phone-2006-01-02-1504." + ParcelExtension

// Waiting is what the phone shows about its unsent events (FR-082).
type Waiting struct {
	// Count is how many events have not been sent.
	Count int
	// Oldest is when the oldest of them was recorded; zero when Count is zero.
	Oldest time.Time
	// HasBatch says whether a completed send exists to send again (FR-097).
	HasBatch bool
}

// Parcel is one send, ready for the share sheet.
type Parcel struct {
	// Name is the file name the share sheet is given.
	Name string
	// Data is the file's contents: an export of Events.
	Data []byte
	// Events are the events the file carries, to mark sent once the share
	// completes (FR-090).
	Events []domain.EventID
}

// Outbox hands the phone's events to the desktop (FR-089 to FR-097). It never
// marks anything sent itself: only a share that completed does that, through
// Confirm.
type Outbox struct {
	store   Store
	sent    Sent
	encoder RecordEncoder
	clock   Clock
	zone    *time.Location
}

// NewOutbox answers an outbox over the phone's store, its record of what was
// sent, the export encoding, a clock and the phone's zone.
func NewOutbox(store Store, sent Sent, encoder RecordEncoder, clock Clock, zone *time.Location) Outbox {
	return Outbox{store: store, sent: sent, encoder: encoder, clock: clock, zone: zone}
}

// Waiting answers how many events are unsent and since when.
func (o Outbox) Waiting() (Waiting, error) {
	unsent, err := o.unsent()
	if err != nil {
		return Waiting{}, err
	}
	batch, err := o.sent.LastBatch()
	if err != nil {
		return Waiting{}, because(ErrNotRead, err)
	}
	waiting := Waiting{Count: len(unsent), HasBatch: len(batch) > 0}
	for _, event := range unsent {
		if waiting.Oldest.IsZero() || event.RecordedAt.Before(waiting.Oldest) {
			waiting.Oldest = event.RecordedAt
		}
	}
	return waiting, nil
}

// Prepare answers every unsent event as one parcel (FR-089, D-6).
func (o Outbox) Prepare() (Parcel, error) {
	unsent, err := o.unsent()
	if err != nil {
		return Parcel{}, err
	}
	if len(unsent) == 0 {
		return Parcel{}, ErrNothingToSend
	}
	return o.parcel(unsent)
}

// Again answers the last completed send's events as a parcel once more
// (FR-095). The events are read as they stand now; one deleted since is left
// out rather than refused.
func (o Outbox) Again() (Parcel, error) {
	batch, err := o.sent.LastBatch()
	if err != nil {
		return Parcel{}, because(ErrNotRead, err)
	}
	if len(batch) == 0 {
		return Parcel{}, ErrNoBatch
	}
	events, err := o.events(func(id domain.EventID) bool { return slices.Contains(batch, id) })
	if err != nil {
		return Parcel{}, err
	}
	if len(events) == 0 {
		return Parcel{}, ErrNoBatch
	}
	return o.parcel(events)
}

// Confirm records that a share of these events completed (FR-090). A share
// that was cancelled or failed never reaches here, so nothing is marked
// (FR-091).
func (o Outbox) Confirm(ids []domain.EventID) error {
	if len(ids) == 0 {
		return ErrNothingToSend
	}
	if err := o.sent.MarkSent(ids); err != nil {
		return because(ErrNotSent, err)
	}
	return nil
}

// unsent answers every event not yet sent.
func (o Outbox) unsent() ([]domain.Event, error) {
	sent, err := o.sent.SentIDs()
	if err != nil {
		return nil, because(ErrNotRead, err)
	}
	return o.events(func(id domain.EventID) bool { return !slices.Contains(sent, id) })
}

// events answers the events whose ids pass keep.
func (o Outbox) events(keep func(domain.EventID) bool) ([]domain.Event, error) {
	held, err := o.store.Events()
	if err != nil {
		return nil, because(ErrNotRead, err)
	}
	var out []domain.Event
	for _, event := range held {
		if keep(event.ID) {
			out = append(out, event)
		}
	}
	return out, nil
}

// parcel encodes events with the definitions they use.
func (o Outbox) parcel(events []domain.Event) (Parcel, error) {
	definitions, err := o.store.Definitions()
	if err != nil {
		return Parcel{}, because(ErrNotRead, err)
	}
	used := make([]domain.Definition, 0, len(definitions))
	ids := make([]domain.EventID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	for _, definition := range definitions {
		if slices.ContainsFunc(events, func(e domain.Event) bool { return e.Definition == definition.ID }) {
			used = append(used, definition)
		}
	}
	data, err := o.encoder.Encode(Record{Definitions: used, Events: events})
	if err != nil {
		return Parcel{}, because(ErrNotExported, err)
	}
	name := product.Name + o.clock.Now().In(o.zone).Format(parcelNameLayout)
	return Parcel{Name: name, Data: data, Events: ids}, nil
}
