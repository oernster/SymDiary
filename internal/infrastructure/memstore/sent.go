package memstore

import (
	"slices"

	"github.com/oernster/symdiary/internal/domain"
)

// SentIDs answers every event that has been in a completed send.
func (s *Store) SentIDs() ([]domain.EventID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ids(s.held.Sent), nil
}

// LastBatch answers the events the last completed send carried.
func (s *Store) LastBatch() ([]domain.EventID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ids(s.held.Batch), nil
}

// MarkSent records a completed send: the events join those sent and become the
// last batch, in one write.
func (s *Store) MarkSent(sent []domain.EventID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.change(func(next *state) error {
		next.Batch = next.Batch[:0]
		for _, id := range sent {
			if !slices.Contains(next.Sent, int64(id)) {
				next.Sent = append(next.Sent, int64(id))
			}
			next.Batch = append(next.Batch, int64(id))
		}
		return nil
	})
}

// ids converts held ids for the application.
func ids(held []int64) []domain.EventID {
	out := make([]domain.EventID, 0, len(held))
	for _, id := range held {
		out = append(out, domain.EventID(id))
	}
	return out
}
