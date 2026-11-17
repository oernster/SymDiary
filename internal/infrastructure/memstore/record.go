package memstore

import (
	"fmt"
	"slices"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// Definitions answers every definition held.
func (s *Store) Definitions() ([]domain.Definition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Definition, 0, len(s.held.Definitions))
	for _, d := range s.held.Definitions {
		out = append(out, domain.Definition{
			ID: domain.DefinitionID(d.ID), Label: d.Label, LastUsed: d.LastUsed,
		})
	}
	return out, nil
}

// Events answers every event held, each carrying its definition's label.
func (s *Store) Events() ([]domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Event, 0, len(s.held.Events))
	for _, e := range s.held.Events {
		out = append(out, s.held.toDomain(e))
	}
	return out, nil
}

// Event answers one event; domain.ErrNoSuchEvent when it is not held.
func (s *Store) Event(id domain.EventID) (domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.held.event(id)
}

// AddEvent writes a new event and answers it as stored.
func (s *Store) AddEvent(input application.EventInput) (domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var id int64
	err := s.change(func(next *state) error {
		id = next.insert(input)
		return nil
	})
	if err != nil {
		return domain.Event{}, err
	}
	return s.held.event(domain.EventID(id))
}

// UpdateEvent replaces an event's symptom, occurrence, severity and note. Its
// recorded-at is never touched, whatever the input carries (FR-007).
func (s *Store) UpdateEvent(id domain.EventID, input application.EventInput) (domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.change(func(next *state) error {
		at := next.indexOf(id)
		if at < 0 {
			return domain.ErrNoSuchEvent
		}
		held := &next.Events[at]
		held.Definition = next.resolve(input.Label, input.Key, held.RecordedAt)
		held.OccurredAt = input.OccurredAt
		held.Severity = input.Severity.String()
		held.Note = input.Note
		return nil
	})
	if err != nil {
		return domain.Event{}, err
	}
	return s.held.event(id)
}

// DeleteEvents removes every listed event, else none: an id that is not held
// leaves the whole record as it was.
func (s *Store) DeleteEvents(ids []domain.EventID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.change(func(next *state) error {
		for _, id := range ids {
			at := next.indexOf(id)
			if at < 0 {
				return domain.ErrNoSuchEvent
			}
			next.Events = slices.Delete(next.Events, at, at+1)
		}
		return nil
	})
}

// RenameDefinition gives a definition a new label and key.
func (s *Store) RenameDefinition(id domain.DefinitionID, label string, key domain.SymptomKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.change(func(next *state) error {
		at := slices.IndexFunc(next.Definitions, func(d definition) bool { return d.ID == int64(id) })
		if at < 0 {
			return domain.ErrNoSuchDefinition
		}
		if other := next.definitionOf(key); other >= 0 && other != at {
			return ErrKeyHeld
		}
		next.Definitions[at].Label = label
		return nil
	})
}

// Import adds the definitions not yet held and then the events, in one write.
func (s *Store) Import(definitions []application.DefinitionInput, events []application.EventInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.change(func(next *state) error {
		for _, d := range definitions {
			next.resolve(d.Label, d.Key, d.LastUsed)
		}
		for _, input := range events {
			next.insert(input)
		}
		return nil
	})
}

// insert adds one event and answers its id.
func (s *state) insert(input application.EventInput) int64 {
	definition := s.resolve(input.Label, input.Key, input.RecordedAt)
	id := s.take()
	s.Events = append(s.Events, event{
		ID: id, Definition: definition, OccurredAt: input.OccurredAt,
		RecordedAt: input.RecordedAt, Severity: input.Severity.String(), Note: input.Note,
	})
	return id
}

// resolve answers the definition holding key, creating it with label when there
// is none; it moves its last use forward to used when that is later.
func (s *state) resolve(label string, key domain.SymptomKey, used time.Time) int64 {
	if at := s.definitionOf(key); at >= 0 {
		if used.After(s.Definitions[at].LastUsed) {
			s.Definitions[at].LastUsed = used
		}
		return s.Definitions[at].ID
	}
	id := s.take()
	s.Definitions = append(s.Definitions, definition{ID: id, Label: label, LastUsed: used})
	return id
}

// definitionOf answers the index of the definition holding key; -1 when none.
func (s *state) definitionOf(key domain.SymptomKey) int {
	return slices.IndexFunc(s.Definitions, func(d definition) bool { return domain.KeyOf(d.Label) == key })
}

// indexOf answers the index of an event; -1 when it is not held.
func (s *state) indexOf(id domain.EventID) int {
	return slices.IndexFunc(s.Events, func(e event) bool { return e.ID == int64(id) })
}

// event answers one event as the domain knows it.
func (s *state) event(id domain.EventID) (domain.Event, error) {
	at := s.indexOf(id)
	if at < 0 {
		return domain.Event{}, domain.ErrNoSuchEvent
	}
	return s.toDomain(s.Events[at]), nil
}

// toDomain converts a held event, carrying its definition's label. Every held
// event names a definition held and a severity that parses: Open refuses a
// snapshot where one does not (check), and every write stores a severity's own
// name, so neither lookup can miss here.
func (s *state) toDomain(e event) domain.Event {
	at := slices.IndexFunc(s.Definitions, func(d definition) bool { return d.ID == e.Definition })
	severity, _ := domain.ParseSeverity(e.Severity)
	return domain.Event{
		ID: domain.EventID(e.ID), Definition: domain.DefinitionID(e.Definition),
		Symptom: s.Definitions[at].Label, OccurredAt: e.OccurredAt, RecordedAt: e.RecordedAt,
		Severity: severity, Note: e.Note,
	}
}

// check refuses a snapshot whose events name a definition that is not held or
// a severity that does not parse, naming the event.
func (s *state) check() error {
	for _, e := range s.Events {
		if !slices.ContainsFunc(s.Definitions, func(d definition) bool { return d.ID == e.Definition }) {
			return fmt.Errorf("%w: event %d names no symptom held", ErrDamaged, e.ID)
		}
		if _, err := domain.ParseSeverity(e.Severity); err != nil {
			return fmt.Errorf("%w: event %d: %w", ErrDamaged, e.ID, err)
		}
	}
	return nil
}
