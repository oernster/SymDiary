package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/oernster/symdiary/internal/domain"
)

// family is the one typeface the record is set in.
//
// The Go fonts are used rather than the screen's own because they carry a
// licence that lets them travel inside the program, plus because they are a Go
// module rather than a file to remember to ship. A PDF's fourteen built-in
// fonts were the alternative and were refused: they can say nothing outside
// Latin-1, so a note with a curly quote or an accented name in it would have
// reached the reader's doctor with the user's own words silently mangled.
const family = "Record"

// grey is what the framing is set in: present, plainly not part of the record.
const grey = 95

// markSize is the side of the letterhead mark in millimetres. It is large
// enough to be recognised across a desk and small enough to read as a
// letterhead rather than as the subject of the page (FR-045).
const markSize = 11.0

// markGap is the space between the mark and the address beside it.
const markGap = 4.0

// Sheet writes a record out as a PDF. Mark is the application's own icon, as
// PNG bytes, drawn beside the line naming the program; a Sheet with no mark
// writes the record without one rather than refusing to write it.
type Sheet struct {
	Mark []byte
}

// Write draws the record at path, answering how many pages it came to. A record
// holding a character the typeface cannot draw is refused before anything is
// written (see Check).
func (s Sheet) Write(path string, lines []domain.Line) (int, error) {
	if err := s.Check(lines); err != nil {
		return 0, err
	}
	doc := newDoc()
	pages := Pages(lines, fontWidths{doc: doc})
	total := len(pages)
	if total == 0 {
		return 0, fmt.Errorf("write %s: the record holds no lines", path)
	}
	s.register(doc)
	for number, rows := range pages {
		doc.AddPage()
		s.draw(doc, rows)
		footer(doc, number+1, total)
	}
	if err := doc.OutputFileAndClose(path); err != nil {
		return 0, fmt.Errorf("write %s: %w", path, err)
	}
	return total, nil
}

// newDoc answers a document set up the way the record is laid out: A4, the
// margins the layout works in, the one typeface in both weights and no padding
// inside a cell.
//
// It is one function rather than two statements of the same setup because the
// widths the layout measures come from this document's own font: a document
// measured one way and drawn another would wrap in places it does not break.
func newDoc() *fpdf.Fpdf {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(marginSide, marginTop, marginSide)
	doc.SetAutoPageBreak(false, marginBottom)
	// Every cell is placed at a measured position, so the padding fpdf would
	// otherwise add inside each one is a millimetre the layout does not know
	// about: text would start a millimetre in from where it was measured and
	// wrap a millimetre past the margin.
	doc.SetCellMargin(0)
	doc.AddUTF8FontFromBytes(family, "", goregular.TTF)
	doc.AddUTF8FontFromBytes(family, "B", gobold.TTF)
	return doc
}

// register hands the mark to the document once, so a record of forty pages
// carries one copy of it rather than forty.
func (s Sheet) register(doc *fpdf.Fpdf) {
	if len(s.Mark) == 0 {
		return
	}
	doc.RegisterImageOptionsReader(
		"mark", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(s.Mark),
	)
}

// draw puts one page's rows on the page.
func (s Sheet) draw(doc *fpdf.Fpdf, rows []Row) {
	for _, row := range rows {
		weight := ""
		if row.Style.bold {
			weight = "B"
		}
		doc.SetFont(family, weight, row.Style.size)
		if row.Style.grey {
			doc.SetTextColor(grey, grey, grey)
		} else {
			doc.SetTextColor(0, 0, 0)
		}

		left := marginSide + row.Indent
		if row.Kind == domain.LineProvenance && len(s.Mark) > 0 {
			doc.ImageOptions("mark", left, row.Top-markSize/2+row.Style.height()/2,
				markSize, markSize, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
			left += markSize + markGap
		}
		doc.SetXY(left, row.Top)
		doc.CellFormat(PageWidth-marginSide-left, row.Style.height(),
			row.Text, "", 0, "L", false, 0, "")
	}
}

// footer writes which page this is and how many there are.
//
// Every page says it, including the only page of a one page record: a reader
// holding one sheet then knows there is not a second one they were never given.
func footer(doc *fpdf.Fpdf, number, total int) {
	doc.SetFont(family, "", 9)
	doc.SetTextColor(grey, grey, grey)
	doc.SetXY(marginSide, PageHeight-footerBaseline)
	doc.CellFormat(PageWidth-2*marginSide, 6,
		fmt.Sprintf("Page %d of %d", number, total), "", 0, "C", false, 0, "")
}

// fontWidths measures a string by asking the document's own font, which is the
// only thing that knows how wide the words will actually be drawn.
type fontWidths struct{ doc *fpdf.Fpdf }

// WidthOf answers the width of text in millimetres at a size and weight.
func (f fontWidths) WidthOf(text string, size float64, bold bool) float64 {
	weight := ""
	if bold {
		weight = "B"
	}
	f.doc.SetFont(family, weight, size)
	return f.doc.GetStringWidth(text)
}
