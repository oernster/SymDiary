package facade

import (
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// WaitingDTO is what the phone shows about its unsent events (FR-082).
type WaitingDTO struct {
	Count int `json:"count"`
	// Oldest is when the oldest unsent event was recorded, as the history
	// shows a time; empty when nothing waits.
	Oldest   string `json:"oldest"`
	HasBatch bool   `json:"hasBatch"`
}

// ParcelDTO is one send, ready for the share sheet.
type ParcelDTO struct {
	Name   string  `json:"name"`
	Text   string  `json:"text"`
	Events []int64 `json:"events"`
}

// Outbox is the part of the facade only the phone recorder has: sending its
// events to the desktop (FR-089 to FR-097). The desktop window never sends,
// so it does not embed this.
type Outbox struct {
	outbox application.Outbox
	zone   *time.Location
}

// NewOutbox answers the phone's sending facade.
func NewOutbox(outbox application.Outbox, zone *time.Location) *Outbox {
	return &Outbox{outbox: outbox, zone: zone}
}

// Waiting answers how many events are unsent and since when.
func (o *Outbox) Waiting() (waiting WaitingDTO, err error) {
	defer Guard(&err)
	held, err := o.outbox.Waiting()
	if err != nil {
		return WaitingDTO{}, err
	}
	out := WaitingDTO{Count: held.Count, HasBatch: held.HasBatch}
	if held.Count > 0 {
		out.Oldest = domain.FormatDisplay(held.Oldest, o.zone)
	}
	return out, nil
}

// Prepare answers every unsent event as one parcel; nothing is marked sent.
func (o *Outbox) Prepare() (parcel ParcelDTO, err error) {
	defer Guard(&err)
	held, err := o.outbox.Prepare()
	if err != nil {
		return ParcelDTO{}, err
	}
	return parcelDTO(held), nil
}

// Again answers the last completed send's events once more (FR-095).
func (o *Outbox) Again() (parcel ParcelDTO, err error) {
	defer Guard(&err)
	held, err := o.outbox.Again()
	if err != nil {
		return ParcelDTO{}, err
	}
	return parcelDTO(held), nil
}

// Confirm records that the share of these events completed (FR-090).
func (o *Outbox) Confirm(ids []int64) (err error) {
	defer Guard(&err)
	return o.outbox.Confirm(eventIDs(ids))
}

// parcelDTO converts a parcel for the page.
func parcelDTO(parcel application.Parcel) ParcelDTO {
	events := make([]int64, 0, len(parcel.Events))
	for _, id := range parcel.Events {
		events = append(events, int64(id))
	}
	return ParcelDTO{Name: parcel.Name, Text: string(parcel.Data), Events: events}
}
