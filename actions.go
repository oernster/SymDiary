package main

import (
	"fmt"

	"github.com/oernster/symdiary/internal/domain"
	"github.com/oernster/symdiary/internal/facade"
	"github.com/oernster/symdiary/internal/product"
)

// The bound methods only a desktop window has: each asks the desktop for a
// file, a browser or a document. The methods every window shares are on
// facade.Core, which App embeds.

// exportNameLayout is the file name the save dialog suggests.
const exportNameLayout = "%s record %s.json"

// Export asks where to save, then writes the record there (FR-050). It answers
// the path written; empty when the user cancelled.
func (a *App) Export() (path string, err error) {
	defer facade.Guard(&err)
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
	defer facade.Guard(&err)
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
	defer facade.Guard(&err)
	receipt, err := facade.ReceiptFor(a.services.History, from, to, appointment)
	if err != nil {
		return "", err
	}
	lines := receipt.Lines(facade.PrintFraming)
	if err := a.sheet.Check(lines); err != nil {
		return "", err
	}
	today := domain.DateOf(a.clock.Now(), a.zone)
	path, err = a.chooser.SavePath(fmt.Sprintf(pdfNameLayout, product.Name, today), documentKind)
	if err != nil || path == "" {
		return "", err
	}
	if _, err := a.sheet.Write(path, lines); err != nil {
		return "", err
	}
	return path, nil
}

// Import asks for an export file, then adds what the record lacks (FR-053).
func (a *App) Import() (result facade.ImportResultDTO, err error) {
	defer facade.Guard(&err)
	path, err := a.chooser.OpenPath()
	if err != nil || path == "" {
		return facade.ImportResultDTO{}, err
	}
	done, err := a.services.Transfer.Import(path)
	if err != nil {
		return facade.ImportResultDTO{}, err
	}
	return facade.ImportResultDTO{Chosen: true, Added: done.Added, Skipped: done.Skipped}, nil
}
