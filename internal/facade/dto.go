package facade

import (
	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// The shapes that cross the window boundary. Each is restated by hand as a
// TypeScript interface in frontend/src/api.ts; the wire test compares the two
// field for field (NFR-MAINT-005).

// StateDTO is what the page needs to start.
type StateDTO struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Problem    string   `json:"problem"`
	Severities []string `json:"severities"`
}

// EventDTO is one event as the history shows it.
type EventDTO struct {
	ID         int64  `json:"id"`
	Definition int64  `json:"definition"`
	Symptom    string `json:"symptom"`
	OccurredAt string `json:"occurredAt"`
	When       string `json:"when"`
	Severity   string `json:"severity"`
	Note       string `json:"note"`
}

// DefinitionDTO is one symptom definition.
type DefinitionDTO struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

// RecordFormDTO is the recording form. An empty occurredAt means now.
type RecordFormDTO struct {
	Symptom    string `json:"symptom"`
	OccurredAt string `json:"occurredAt"`
	Severity   string `json:"severity"`
	Note       string `json:"note"`
}

// EditFormDTO is an edit. An empty occurredAt keeps the time held (FR-024).
type EditFormDTO struct {
	Symptom    string `json:"symptom"`
	OccurredAt string `json:"occurredAt"`
	Severity   string `json:"severity"`
	Note       string `json:"note"`
}

// FilterDTO is the history filter. Empty dates leave that end open.
type FilterDTO struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Definitions []int64  `json:"definitions"`
	Severities  []string `json:"severities"`
}

// ReceiptLineDTO is one line of the receipt.
type ReceiptLineDTO struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

// ImportResultDTO says what an import did; chosen is false when the user
// cancelled the file dialog.
type ImportResultDTO struct {
	Chosen  bool `json:"chosen"`
	Added   int  `json:"added"`
	Skipped int  `json:"skipped"`
}

// CreditDTO names one open-source work.
type CreditDTO struct {
	Work    string `json:"work"`
	Licence string `json:"licence"`
	Holder  string `json:"holder"`
}

// AboutDTO is the About dialog.
type AboutDTO struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Author    string      `json:"author"`
	Copyright string      `json:"copyright"`
	Statement string      `json:"statement"`
	Credits   []CreditDTO `json:"credits"`
}

// eventDTO converts an event for the page.
func (c *Core) eventDTO(event domain.Event) EventDTO {
	return EventDTO{
		ID:         int64(event.ID),
		Definition: int64(event.Definition),
		Symptom:    event.Symptom,
		OccurredAt: domain.FormatLocal(event.OccurredAt, c.zone),
		When:       domain.FormatDisplay(event.OccurredAt, c.zone),
		Severity:   event.Severity.String(),
		Note:       event.Note,
	}
}

// eventDTOs converts events for the page.
func (c *Core) eventDTOs(events []domain.Event) []EventDTO {
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, c.eventDTO(event))
	}
	return out
}

// definitionDTOs converts definitions for the page.
func definitionDTOs(definitions []domain.Definition) []DefinitionDTO {
	out := make([]DefinitionDTO, 0, len(definitions))
	for _, definition := range definitions {
		out = append(out, DefinitionDTO{ID: int64(definition.ID), Label: definition.Label})
	}
	return out
}

// aboutDTO converts the About content for the page.
func aboutDTO(about application.About) AboutDTO {
	credits := make([]CreditDTO, 0, len(about.Credits))
	for _, credit := range about.Credits {
		credits = append(credits, CreditDTO(credit))
	}
	return AboutDTO{
		Name: about.Name, Version: about.Version, Author: about.Author,
		Copyright: about.Copyright, Statement: about.Statement, Credits: credits,
	}
}
