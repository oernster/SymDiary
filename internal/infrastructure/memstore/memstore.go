// Package memstore keeps the phone recorder's record in memory and saves the
// whole of it as one snapshot after every change (Amendment 21).
//
// The phone's browser storage is reached through Snapshots, so everything here
// is ordinary Go that the suite exercises on any machine; only the adapter that
// talks to the browser is WebAssembly-only. One snapshot per write makes every
// write whole or not at all, which is the Store contract: a change is applied
// to a copy, the copy is saved, and only a save that succeeded replaces what
// is held (FR-101).
package memstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

// schemaVersion is the shape of the snapshot this code writes. A snapshot
// naming a later one was written by a newer recorder and is refused rather
// than guessed at (FR-102).
const schemaVersion = 1

// The refusals opening a snapshot can answer.
var (
	ErrNewerSchema = errors.New("this phone's record was written by a newer SymDiary, so it is not read")
	ErrNoSchema    = errors.New("this phone's record names no version SymDiary wrote, so it is not read")
	ErrKeyHeld     = errors.New("another symptom already has that name")
	ErrDamaged     = errors.New("this phone's record is damaged, so it is not read")
)

// Snapshots keeps the record's bytes somewhere that outlives the page.
type Snapshots interface {
	// Load answers the saved bytes; nil when nothing has been saved yet.
	Load() ([]byte, error)
	// Save replaces what is saved, answering only once it is kept.
	Save(data []byte) error
}

// definition is a definition as the snapshot holds it.
type definition struct {
	ID       int64     `json:"id"`
	Label    string    `json:"label"`
	LastUsed time.Time `json:"lastUsed"`
}

// event is an event as the snapshot holds it, its symptom held by reference.
type event struct {
	ID         int64     `json:"id"`
	Definition int64     `json:"definition"`
	OccurredAt time.Time `json:"occurredAt"`
	RecordedAt time.Time `json:"recordedAt"`
	Severity   string    `json:"severity"`
	Note       string    `json:"note"`
}

// state is the whole snapshot.
type state struct {
	Schema      int          `json:"schema"`
	NextID      int64        `json:"nextId"`
	Definitions []definition `json:"definitions"`
	Events      []event      `json:"events"`
	// Sent and Batch belong to the Outbox: every event in a completed send,
	// then the events of the last one (FR-090, FR-095).
	Sent  []int64 `json:"sent"`
	Batch []int64 `json:"batch"`
}

// clone answers a copy that shares no slice with s, so a change to the copy
// leaves s as it was.
func (s state) clone() state {
	s.Definitions = slices.Clone(s.Definitions)
	s.Events = slices.Clone(s.Events)
	s.Sent = slices.Clone(s.Sent)
	s.Batch = slices.Clone(s.Batch)
	return s
}

// Store is the phone's record. It implements application.Store and
// application.Sent. Calls arrive on goroutines of their own, so one lock
// serialises them.
type Store struct {
	mu        sync.Mutex
	snapshots Snapshots
	held      state
}

// Open reads the saved record; an empty one when nothing has been saved.
func Open(snapshots Snapshots) (*Store, error) {
	data, err := snapshots.Load()
	if err != nil {
		return nil, err
	}
	held := state{Schema: schemaVersion, NextID: 1}
	if data != nil {
		var envelope struct {
			Schema int `json:"schema"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			return nil, fmt.Errorf("reading this phone's record: %w", err)
		}
		switch {
		case envelope.Schema > schemaVersion:
			return nil, ErrNewerSchema
		case envelope.Schema < schemaVersion:
			return nil, ErrNoSchema
		}
		if err := json.Unmarshal(data, &held); err != nil {
			return nil, fmt.Errorf("reading this phone's record: %w", err)
		}
		if err := held.check(); err != nil {
			return nil, err
		}
	}
	return &Store{snapshots: snapshots, held: held}, nil
}

// change applies act to a copy of the record and saves the copy; only a save
// that succeeded replaces what is held.
func (s *Store) change(act func(next *state) error) error {
	next := s.held.clone()
	if err := act(&next); err != nil {
		return err
	}
	data, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if err := s.snapshots.Save(data); err != nil {
		return err
	}
	s.held = next
	return nil
}

// take answers the next id and moves the counter on.
func (s *state) take() int64 {
	id := s.NextID
	s.NextID++
	return id
}
