package application

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// History answers what has been recorded: the events, the suggestions and the
// symptom list.
type History struct {
	store Store
	zone  *time.Location
}

// NewHistory answers a history over a store in the user's zone.
func NewHistory(store Store, zone *time.Location) History {
	return History{store: store, zone: zone}
}

// List answers the events passing the filter, newest first (FR-020, FR-021).
func (h History) List(filter domain.Filter) ([]domain.Event, error) {
	if err := domain.CheckRange(filter.From, filter.To); err != nil {
		return nil, err
	}
	events, err := h.store.Events()
	if err != nil {
		return nil, because(ErrNotRead, err)
	}
	return domain.Select(events, filter, h.zone), nil
}

// Suggest answers the definitions matching what has been typed (FR-011).
func (h History) Suggest(typed string) ([]domain.Definition, error) {
	definitions, err := h.store.Definitions()
	if err != nil {
		return nil, because(ErrNotRead, err)
	}
	return domain.Suggest(definitions, typed), nil
}

// Symptoms answers every definition in label order, ignoring case: the list the
// symptoms screen shows and the history filter offers.
func (h History) Symptoms() ([]domain.Definition, error) {
	definitions, err := h.store.Definitions()
	if err != nil {
		return nil, because(ErrNotRead, err)
	}
	slices.SortFunc(definitions, func(a, b domain.Definition) int {
		if byLabel := strings.Compare(string(domain.KeyOf(a.Label)),
			string(domain.KeyOf(b.Label))); byLabel != 0 {
			return byLabel
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return definitions, nil
}

// Rename gives a definition a new label; every event using it follows (FR-028).
func (h History) Rename(id domain.DefinitionID, label string) error {
	definitions, err := h.store.Definitions()
	if err != nil {
		return because(ErrNotRenamed, err)
	}
	if err := domain.CheckRename(definitions, id, label); err != nil {
		return err
	}
	if err := h.store.RenameDefinition(id, label, domain.KeyOf(label)); err != nil {
		return because(ErrNotRenamed, err)
	}
	return nil
}

// Receipt answers the receipt for a range (FR-040 to FR-044), naming the last
// appointment when one is given; a zero appointment names none (FR-046).
func (h History) Receipt(from, to, appointment domain.Date) (domain.Receipt, error) {
	events, err := h.store.Events()
	if err != nil {
		return domain.Receipt{}, because(ErrNotRead, err)
	}
	receipt, err := domain.BuildReceipt(events, from, to, h.zone)
	if err != nil {
		return domain.Receipt{}, err
	}
	return receipt.WithAppointment(appointment)
}
