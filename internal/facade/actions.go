package facade

import (
	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/product"
)

// The bound methods both windows share, one per user-visible action. Each
// converts the page's shape, calls one application service and converts the
// answer back.

// Record records an event (FR-001).
func (c *Core) Record(form RecordFormDTO) (saved EventDTO, err error) {
	defer Guard(&err)
	occurred, err := c.occurrence(form.OccurredAt)
	if err != nil {
		return EventDTO{}, err
	}
	severity, err := domain.ParseSeverity(form.Severity)
	if err != nil {
		return EventDTO{}, err
	}
	event, err := c.services.Recorder.Record(application.RecordForm{
		Symptom: form.Symptom, OccurredAt: occurred, Severity: severity, Note: form.Note,
	})
	if err != nil {
		return EventDTO{}, err
	}
	return c.eventDTO(event), nil
}

// Suggest answers the symptoms matching what has been typed (FR-011).
func (c *Core) Suggest(typed string) (found []DefinitionDTO, err error) {
	defer Guard(&err)
	definitions, err := c.services.History.Suggest(typed)
	if err != nil {
		return nil, err
	}
	return definitionDTOs(definitions), nil
}

// Symptoms answers every symptom in label order.
func (c *Core) Symptoms() (found []DefinitionDTO, err error) {
	defer Guard(&err)
	definitions, err := c.services.History.Symptoms()
	if err != nil {
		return nil, err
	}
	return definitionDTOs(definitions), nil
}

// History answers the events passing a filter, newest first (FR-020, FR-021).
func (c *Core) History(filter FilterDTO) (events []EventDTO, err error) {
	defer Guard(&err)
	from, err := optionalDate(filter.From)
	if err != nil {
		return nil, err
	}
	to, err := optionalDate(filter.To)
	if err != nil {
		return nil, err
	}
	wanted := domain.Filter{From: from, To: to}
	for _, id := range filter.Definitions {
		wanted.Definitions = append(wanted.Definitions, domain.DefinitionID(id))
	}
	for _, name := range filter.Severities {
		severity, err := domain.ParseSeverity(name)
		if err != nil {
			return nil, err
		}
		wanted.Severities = append(wanted.Severities, severity)
	}
	listed, err := c.services.History.List(wanted)
	if err != nil {
		return nil, err
	}
	return c.eventDTOs(listed), nil
}

// Edit changes an event (FR-023, FR-024).
func (c *Core) Edit(id int64, form EditFormDTO) (saved EventDTO, err error) {
	defer Guard(&err)
	occurred, err := c.occurrence(form.OccurredAt)
	if err != nil {
		return EventDTO{}, err
	}
	severity, err := domain.ParseSeverity(form.Severity)
	if err != nil {
		return EventDTO{}, err
	}
	event, err := c.services.Editor.Edit(domain.EventID(id), application.EditForm{
		Symptom: form.Symptom, OccurredAt: occurred, Severity: severity, Note: form.Note,
	})
	if err != nil {
		return EventDTO{}, err
	}
	return c.eventDTO(event), nil
}

// DeletionPrompt answers the confirmation to show before a deletion (FR-025).
func (c *Core) DeletionPrompt(ids []int64) (prompt string, err error) {
	defer Guard(&err)
	return c.services.Editor.DeletionPrompt(eventIDs(ids))
}

// Delete removes events the user has confirmed.
func (c *Core) Delete(ids []int64) (err error) {
	defer Guard(&err)
	return c.services.Editor.Delete(eventIDs(ids))
}

// Rename renames a symptom (FR-028).
func (c *Core) Rename(id int64, label string) (err error) {
	defer Guard(&err)
	return c.services.History.Rename(domain.DefinitionID(id), label)
}

// PrintFraming is the fixed text printed around every record (FR-045). It is
// built here, where the product's own words meet the domain that cannot read
// them.
var PrintFraming = domain.Framing{
	Provenance: product.PrintProvenance,
	Statement:  product.PrintStatement,
}

// ReceiptFor reads the three dates the page sends and answers the receipt they
// describe. The window and the document both come through here, so the record
// on screen and the one saved cannot be asked for in two different ways. An
// empty appointment names none (FR-046).
//
// It is a function rather than a method so the window binds nothing extra.
func ReceiptFor(history application.History, from, to, appointment string) (domain.Receipt, error) {
	texts := [...]string{from, to, appointment}
	var dates [len(texts)]domain.Date
	for at, text := range texts {
		date, err := optionalDate(text)
		if err != nil {
			return domain.Receipt{}, err
		}
		dates[at] = date
	}
	return history.Receipt(dates[0], dates[1], dates[2])
}

// Receipt answers the receipt's lines for a range (FR-040, FR-041), naming the
// last appointment when the page sends one (FR-046).
func (c *Core) Receipt(from, to, appointment string) (lines []ReceiptLineDTO, err error) {
	defer Guard(&err)
	receipt, err := ReceiptFor(c.services.History, from, to, appointment)
	if err != nil {
		return nil, err
	}
	for _, line := range receipt.Lines(PrintFraming) {
		lines = append(lines, ReceiptLineDTO{Kind: string(line.Kind), Text: line.Text})
	}
	return lines, nil
}
