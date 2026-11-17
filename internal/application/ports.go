// Package application holds SymDiary's use cases, one per user-visible action.
//
// It depends on the domain and on the ports declared in this file. The
// infrastructure implements the ports; the composition root wires them in.
package application

import (
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// Clock answers the current instant. The real one reads the system clock; tests
// hand in a fixed one.
type Clock interface {
	Now() time.Time
}

// EventInput is everything the store needs to write an event. The store finds
// the definition whose key is Key, creating it with Label when there is none, so
// a case variant of a held symptom reuses it (FR-014).
type EventInput struct {
	Label      string
	Key        domain.SymptomKey
	OccurredAt time.Time
	RecordedAt time.Time
	Severity   domain.Severity
	Note       string
}

// DefinitionInput is a definition carried in from an export file.
type DefinitionInput struct {
	Label    string
	Key      domain.SymptomKey
	LastUsed time.Time
}

// Store keeps the record. Every write is one transaction: it happens whole or
// not at all.
type Store interface {
	// Definitions answers every definition held.
	Definitions() ([]domain.Definition, error)
	// Events answers every event held, each carrying its definition's label.
	Events() ([]domain.Event, error)
	// Event answers one event; domain.ErrNoSuchEvent when it is not held.
	Event(id domain.EventID) (domain.Event, error)
	// AddEvent writes a new event and answers it as stored.
	AddEvent(input EventInput) (domain.Event, error)
	// UpdateEvent replaces an event's symptom, occurrence, severity and note.
	// Its recorded-at is never touched, whatever the input carries (FR-007).
	UpdateEvent(id domain.EventID, input EventInput) (domain.Event, error)
	// DeleteEvents removes every listed event, else none of them.
	DeleteEvents(ids []domain.EventID) error
	// RenameDefinition gives a definition a new label and key.
	RenameDefinition(id domain.DefinitionID, label string, key domain.SymptomKey) error
	// Import adds definitions not yet held and then the events, in one
	// transaction.
	Import(definitions []DefinitionInput, events []EventInput) error
}

// Record is the whole of what SymDiary keeps, as an export carries it.
type Record struct {
	Definitions []domain.Definition
	Events      []domain.Event
}

// Sent remembers which events the phone recorder has handed to the desktop
// (Amendment 21). The desktop hands nothing over, so it answers NothingSent.
type Sent interface {
	// SentIDs answers every event that has been in a completed send.
	SentIDs() ([]domain.EventID, error)
	// LastBatch answers the events the last completed send carried; none
	// before the first send.
	LastBatch() ([]domain.EventID, error)
	// MarkSent records a completed send of these events, which become the last
	// batch, in one write.
	MarkSent(ids []domain.EventID) error
}

// NothingSent is the desktop's Sent: no event of its record is ever handed
// over, so no rule that turns on sending applies to it.
type NothingSent struct{}

// SentIDs answers none.
func (NothingSent) SentIDs() ([]domain.EventID, error) { return nil, nil }

// LastBatch answers none.
func (NothingSent) LastBatch() ([]domain.EventID, error) { return nil, nil }

// MarkSent refuses: the desktop never sends.
func (NothingSent) MarkSent([]domain.EventID) error { return ErrNothingToSend }

// RecordEncoder turns a record into the bytes of an export file, for a window
// that hands the file over rather than writing it to a path.
type RecordEncoder interface {
	Encode(record Record) ([]byte, error)
}

// RecordFile writes and reads an export file.
type RecordFile interface {
	// Write writes the record to path, leaving no partial file on failure.
	Write(path string, record Record) error
	// Read reads a record from path.
	Read(path string) (Record, error)
}
