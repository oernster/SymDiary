package application

import "github.com/oernster/symdiary/internal/domain"

// Unavailable stands in for a record that could not be opened. Every operation
// refuses with the reason, so the window still opens and every action says why
// it cannot act, while the record itself is never touched (FR-062). It does no
// I/O, so it lives here where both windows reach it: the desktop's SQLite
// record and the phone recorder's snapshot record each fall back to it.
type Unavailable struct {
	Reason error
}

// Definitions refuses.
func (u Unavailable) Definitions() ([]domain.Definition, error) { return nil, u.Reason }

// Events refuses.
func (u Unavailable) Events() ([]domain.Event, error) { return nil, u.Reason }

// Event refuses.
func (u Unavailable) Event(domain.EventID) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// AddEvent refuses.
func (u Unavailable) AddEvent(EventInput) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// UpdateEvent refuses.
func (u Unavailable) UpdateEvent(domain.EventID, EventInput) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// DeleteEvents refuses.
func (u Unavailable) DeleteEvents([]domain.EventID) error { return u.Reason }

// RenameDefinition refuses.
func (u Unavailable) RenameDefinition(domain.DefinitionID, string, domain.SymptomKey) error {
	return u.Reason
}

// Import refuses.
func (u Unavailable) Import([]DefinitionInput, []EventInput) error { return u.Reason }

// SentIDs refuses.
func (u Unavailable) SentIDs() ([]domain.EventID, error) { return nil, u.Reason }

// LastBatch refuses.
func (u Unavailable) LastBatch() ([]domain.EventID, error) { return nil, u.Reason }

// MarkSent refuses.
func (u Unavailable) MarkSent([]domain.EventID) error { return u.Reason }
