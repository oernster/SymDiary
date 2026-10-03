package main

import (
	"fmt"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/product"
)

// The bound methods, one per user-visible action. Each converts the page's
// shape, calls one application service and converts the answer back.

// Record records an event (FR-001).
func (a *App) Record(form RecordFormDTO) (saved EventDTO, err error) {
	defer guard(&err)
	occurred, err := a.occurrence(form.OccurredAt)
	if err != nil {
		return EventDTO{}, err
	}
	severity, err := domain.ParseSeverity(form.Severity)
	if err != nil {
		return EventDTO{}, err
	}
	event, err := a.services.Recorder.Record(application.RecordForm{
		Symptom: form.Symptom, OccurredAt: occurred, Severity: severity, Note: form.Note,
	})
	if err != nil {
		return EventDTO{}, err
	}
	return a.eventDTO(event), nil
}

// Suggest answers the symptoms matching what has been typed (FR-011).
func (a *App) Suggest(typed string) (found []DefinitionDTO, err error) {
	defer guard(&err)
	definitions, err := a.services.History.Suggest(typed)
	if err != nil {
		return nil, err
	}
	return definitionDTOs(definitions), nil
}

// Symptoms answers every symptom in label order.
func (a *App) Symptoms() (found []DefinitionDTO, err error) {
	defer guard(&err)
	definitions, err := a.services.History.Symptoms()
	if err != nil {
		return nil, err
	}
	return definitionDTOs(definitions), nil
}

// History answers the events passing a filter, newest first (FR-020, FR-021).
func (a *App) History(filter FilterDTO) (events []EventDTO, err error) {
	defer guard(&err)
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
	listed, err := a.services.History.List(wanted)
	if err != nil {
		return nil, err
	}
	return a.eventDTOs(listed), nil
}

// Edit changes an event (FR-023, FR-024).
func (a *App) Edit(id int64, form EditFormDTO) (saved EventDTO, err error) {
	defer guard(&err)
	occurred, err := a.occurrence(form.OccurredAt)
	if err != nil {
		return EventDTO{}, err
	}
	severity, err := domain.ParseSeverity(form.Severity)
	if err != nil {
		return EventDTO{}, err
	}
	event, err := a.services.Editor.Edit(domain.EventID(id), application.EditForm{
		Symptom: form.Symptom, OccurredAt: occurred, Severity: severity, Note: form.Note,
	})
	if err != nil {
		return EventDTO{}, err
	}
	return a.eventDTO(event), nil
}

// DeletionPrompt answers the confirmation to show before a deletion (FR-025).
func (a *App) DeletionPrompt(ids []int64) (prompt string, err error) {
	defer guard(&err)
	return a.services.Editor.DeletionPrompt(eventIDs(ids))
}

// Delete removes events the user has confirmed.
func (a *App) Delete(ids []int64) (err error) {
	defer guard(&err)
	return a.services.Editor.Delete(eventIDs(ids))
}

// Rename renames a symptom (FR-028).
func (a *App) Rename(id int64, label string) (err error) {
	defer guard(&err)
	return a.services.History.Rename(domain.DefinitionID(id), label)
}

// printFraming is the fixed text printed around every record (FR-045). It is
// built here, where the product's own words meet the domain that cannot read
// them.
var printFraming = domain.Framing{
	Provenance: product.PrintProvenance,
	Statement:  product.PrintStatement,
}

// receiptFor reads the three dates the page sends and answers the receipt they
// describe. The window and the document both come through here, so the record
// on screen and the one saved cannot be asked for in two different ways. An
// empty appointment names none (FR-046).
func (a *App) receiptFor(from, to, appointment string) (domain.Receipt, error) {
	texts := [...]string{from, to, appointment}
	var dates [len(texts)]domain.Date
	for at, text := range texts {
		date, err := optionalDate(text)
		if err != nil {
			return domain.Receipt{}, err
		}
		dates[at] = date
	}
	return a.services.History.Receipt(dates[0], dates[1], dates[2])
}

// Receipt answers the receipt's lines for a range (FR-040, FR-041), naming the
// last appointment when the page sends one (FR-046).
func (a *App) Receipt(from, to, appointment string) (lines []ReceiptLineDTO, err error) {
	defer guard(&err)
	receipt, err := a.receiptFor(from, to, appointment)
	if err != nil {
		return nil, err
	}
	for _, line := range receipt.Lines(printFraming) {
		lines = append(lines, ReceiptLineDTO{Kind: string(line.Kind), Text: line.Text})
	}
	return lines, nil
}

// exportNameLayout is the file name the save dialog suggests.
const exportNameLayout = "%s record %s.json"

// Export asks where to save, then writes the record there (FR-050). It answers
// the path written; empty when the user cancelled.
func (a *App) Export() (path string, err error) {
	defer guard(&err)
	today := domain.DateOf(a.clock.Now(), a.zone)
	path, err = a.chooser.SavePath(fmt.Sprintf(exportNameLayout, product.Name, today), exportKind)
	if err != nil || path == "" {
		return "", err
	}
	if err := a.services.Transfer.Export(path); err != nil {
		return "", err
	}
	return path, nil
}

// Donate opens the donation page in the user's browser (FR-069).
//
// SymDiary does not fetch that page: it hands the address to the desktop and the
// browser does the asking, so the button leaves the no-network guarantee
// untouched. The address does not cross the wire either. The page asks for the
// donation page rather than naming one, so there is nothing arriving from
// outside that would have to be checked before it was opened.
func (a *App) Donate() (err error) {
	defer guard(&err)
	a.opener.Open(product.DonateURL)
	return nil
}

// pdfNameLayout is the file name the save dialog suggests for a record.
const pdfNameLayout = "%s symptom record %s.pdf"

// SavePDF writes the record for a range to a file the reader chooses,
// answering the path written; empty when they cancelled (FR-040).
//
// The page shows the same record on screen and the two are built from one
// source: both ask the history for the receipt and both draw the framing of
// FR-045 around it. What the page cannot do is decide where a page ends, which
// is why the document is drawn rather than printed.
func (a *App) SavePDF(from, to, appointment string) (path string, err error) {
	defer guard(&err)
	receipt, err := a.receiptFor(from, to, appointment)
	if err != nil {
		return "", err
	}
	today := domain.DateOf(a.clock.Now(), a.zone)
	path, err = a.chooser.SavePath(fmt.Sprintf(pdfNameLayout, product.Name, today), documentKind)
	if err != nil || path == "" {
		return "", err
	}
	if _, err := a.sheet.Write(path, receipt.Lines(printFraming)); err != nil {
		return "", err
	}
	return path, nil
}

// Import asks for an export file, then adds what the record lacks (FR-053).
func (a *App) Import() (result ImportResultDTO, err error) {
	defer guard(&err)
	path, err := a.chooser.OpenPath()
	if err != nil || path == "" {
		return ImportResultDTO{}, err
	}
	done, err := a.services.Transfer.Import(path)
	if err != nil {
		return ImportResultDTO{}, err
	}
	return ImportResultDTO{Chosen: true, Added: done.Added, Skipped: done.Skipped}, nil
}
