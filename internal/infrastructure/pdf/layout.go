// Package pdf writes a symptom record out as a PDF.
//
// SymDiary used to hand the record to the browser's own print path. That path
// belongs to three different engines, so what came off the paper depended on
// three things: which desktop the reader was using; whether their print dialog
// had "Headers and footers" ticked; which of the CSS the sheet leaned on their
// engine had implemented. The record is the product, so the record is now drawn
// here, once, where the same bytes reach every reader (FR-040, Amendment 16).
//
// The layout in this file is a pure function of the record and the page
// geometry. Nothing here opens a file, measures a font or knows what a PDF is:
// the width of a string arrives through Measurer, so what lands on which page
// can be tested without drawing anything.
package pdf

import (
	"strings"

	"github.com/oernster/symdiary/internal/domain"
)

// Millimetres of paper. A4 is the size a record printed in the United Kingdom
// is read on; a PDF names its own page size, so a reader on another paper size
// gets a document their printer scales rather than one that is cut off.
const (
	PageWidth  = 210.0
	PageHeight = 297.0

	marginTop    = 16.0
	marginSide   = 16.0
	marginBottom = 18.0

	// footerBaseline is where the page number sits, measured up from the foot
	// of the paper. It is inside the bottom margin, which is why the margin is
	// the deeper of the two.
	footerBaseline = 10.0
)

// pointsToMM converts a font size to the unit the page is laid out in.
const pointsToMM = 25.4 / 72.0

// lineSpacing is the height of one line as a multiple of its font size. It is
// the figure the window uses, so a record reads at the same pace on paper.
const lineSpacing = 1.45

// style is how one kind of line is drawn: its size in points, whether it is
// bold, whether it is set in grey, how far it is indented and how much space
// is left above it.
type style struct {
	size       float64
	bold       bool
	grey       bool
	indent     float64
	spaceAbove float64
}

// height answers the height of one line in this style, in millimetres.
func (s style) height() float64 { return s.size * pointsToMM * lineSpacing }

// styles is how each kind of line the domain writes is drawn. A kind with no
// entry here would be drawn as an ordinary line, which is why there is a test
// asserting every kind the domain can write has one.
var styles = map[domain.LineKind]style{
	domain.LineProvenance: {size: 9, grey: true},
	domain.LineTitle:      {size: 14, bold: true, spaceAbove: 5},
	domain.LineRange:      {size: 10.5},
	// The appointment reads as part of the range it sits in, not as a heading.
	domain.LineAppointment: {size: 10.5},
	domain.LineStatement:   {size: 9, grey: true, spaceAbove: 2},
	domain.LineHeading:     {size: 11.5, bold: true, spaceAbove: 7},
	domain.LineWhen:        {size: 10.5, spaceAbove: 3},
	domain.LineSeverity:    {size: 10.5, indent: 6},
	domain.LineNote:        {size: 10.5, indent: 6},
}

// Measurer answers how wide a string would be, in millimetres, at a size and
// weight. The renderer asks the font; a test answers from a rule of its own, so
// what lands on which page can be settled without a font at all.
type Measurer interface {
	WidthOf(text string, size float64, bold bool) float64
}

// Row is one line of text as it will be drawn: the words after wrapping, the
// style to draw them in and how far down the page the line sits.
type Row struct {
	Kind   domain.LineKind
	Text   string
	Style  style
	Top    float64
	Indent float64
}

// block is a run of lines that belongs together on one page.
//
// A heading is its own block. An event is one block holding its time, its
// severity and its note, so a reader never meets a time at the foot of one
// sheet and what was observed at the head of the next.
type block struct {
	rows   []Row
	height float64
}

// Pages lays the record out, answering the rows of each page in order.
//
// A block that does not fit in what is left of a page starts the next one. A
// block taller than a whole page cannot be kept whole on any page, so it
// continues onto the next (see flow): losing a line of somebody's medical record
// to make the layout tidy is the one outcome worth avoiding above all others.
func Pages(lines []domain.Line, measure Measurer) [][]Row {
	textWidth := PageWidth - 2*marginSide
	limit := PageHeight - marginBottom

	var pages [][]Row
	var current []Row
	top := marginTop

	for _, item := range blocks(lines, textWidth, measure) {
		if len(current) > 0 && top+item.height > limit {
			pages = append(pages, current)
			current = nil
			top = marginTop
		}
		// A block at the head of a page has nothing above it to be separated
		// from, so the gap it would leave goes. Keeping it would give each
		// sheet a top margin that depended on what happened to land there.
		if len(current) == 0 {
			item = withoutSpaceAbove(item)
		}
		if top+item.height > limit {
			pages, current, top = flow(item, pages, current, top, limit)
			continue
		}
		for _, row := range item.rows {
			row.Top += top
			current = append(current, row)
		}
		top += item.height
	}
	if len(current) > 0 {
		pages = append(pages, current)
	}
	return pages
}

// flow places a block taller than a page, starting at top on the current page
// and continuing onto as many more as it needs. It answers the pages, the page
// now open and the height reached on it.
//
// Each continuation page opens with the block's first row again where that row
// is an event's time. The time is a recorded field, so the sheet still carries
// no words of the program's own, while a reader meeting a page of note can see
// which observation it belongs to.
func flow(item block, pages [][]Row, current []Row, top, limit float64) ([][]Row, []Row, float64) {
	origin := top
	for i, row := range item.rows {
		placed := origin + row.Top
		if i > 0 && placed+row.Style.height() > limit {
			pages = append(pages, current)
			current = nil
			top = marginTop
			if lead := item.rows[0]; lead.Kind == domain.LineWhen {
				lead.Top = top
				current = append(current, lead)
				top += lead.Style.height()
			}
			origin = top - row.Top
			placed = top
		}
		row.Top = placed
		current = append(current, row)
	}
	return pages, current, origin + item.height
}

// withoutSpaceAbove drops the gap a block would have left above itself. At the
// head of a fresh page there is nothing above it to be separated from; the
// gap would otherwise read as a margin that varies from sheet to sheet.
func withoutSpaceAbove(item block) block {
	if len(item.rows) == 0 {
		return item
	}
	gap := item.rows[0].Top
	if gap == 0 {
		return item
	}
	shifted := make([]Row, len(item.rows))
	for i, row := range item.rows {
		row.Top -= gap
		shifted[i] = row
	}
	return block{rows: shifted, height: item.height - gap}
}

// blocks groups the record's lines into the runs that must not be split, each
// already wrapped to the width it will be drawn in.
func blocks(lines []domain.Line, textWidth float64, measure Measurer) []block {
	var out []block
	var open *block
	for _, line := range lines {
		drawn := styles[line.Kind]
		rows := rowsFor(line, textWidth, measure)
		if startsABlock(line.Kind) || open == nil {
			out = append(out, block{})
			open = &out[len(out)-1]
		}
		// Each line's rows are placed below whatever the block already holds,
		// and the block grows by the gap above the line plus the lines it
		// wrapped to. Growing it row by row would lose the gap.
		base := open.height
		for _, row := range rows {
			row.Top += base
			open.rows = append(open.rows, row)
		}
		open.height = base + drawn.spaceAbove + float64(len(rows))*drawn.height()
	}
	return out
}

// startsABlock says whether a kind begins a new run. Everything that follows a
// time belongs to that time until the next one.
func startsABlock(kind domain.LineKind) bool {
	switch kind {
	case domain.LineSeverity, domain.LineNote:
		return false
	default:
		return true
	}
}

// rowsFor answers the drawn lines one recorded line becomes, wrapping its words
// to the page and placing each below the last. Tops are relative to the block.
func rowsFor(line domain.Line, textWidth float64, measure Measurer) []Row {
	drawn := styles[line.Kind]
	width := textWidth - drawn.indent
	var out []Row
	top := drawn.spaceAbove
	for _, piece := range wrap(line.Text, drawn, width, measure) {
		out = append(out, Row{
			Kind: line.Kind, Text: piece, Style: drawn, Top: top, Indent: drawn.indent,
		})
		top += drawn.height()
	}
	return out
}

// wrap breaks a line into pieces that fit the width: at each line break the
// user typed, then on spaces within each paragraph. A paragraph left empty by
// two breaks in a row stays as an empty row, as it does in the window; breaks
// trailing the text add nothing.
func wrap(text string, drawn style, width float64, measure Measurer) []string {
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for len(paragraphs) > 1 && strings.TrimSpace(paragraphs[len(paragraphs)-1]) == "" {
		paragraphs = paragraphs[:len(paragraphs)-1]
	}
	var out []string
	for _, paragraph := range paragraphs {
		out = append(out, wrapParagraph(paragraph, drawn, width, measure)...)
	}
	return out
}

// wrapParagraph breaks one paragraph into pieces that fit the width, on spaces.
//
// A single word wider than the line is broken at the line rather than left
// whole. Left whole, it was measured running off the paper rather than into the
// margin, so the end of what the user typed was not on the sheet at all.
func wrapParagraph(text string, drawn style, width float64, measure Measurer) []string {
	fits := func(piece string) bool { return measure.WidthOf(piece, drawn.size, drawn.bold) <= width }
	var out []string
	line := ""
	for _, word := range strings.Fields(text) {
		switch {
		case line != "" && fits(line+" "+word):
			line += " " + word
		case fits(word):
			if line != "" {
				out = append(out, line)
			}
			line = word
		default:
			if line != "" {
				out = append(out, line)
			}
			pieces := breakWord(word, fits)
			out = append(out, pieces[:len(pieces)-1]...)
			line = pieces[len(pieces)-1]
		}
	}
	return append(out, line)
}

// breakWord cuts a word too wide for the line into pieces that each fit, never
// fewer than one character to a piece so the cutting always ends.
func breakWord(word string, fits func(string) bool) []string {
	var out []string
	piece := ""
	for _, character := range word {
		if piece != "" && !fits(piece+string(character)) {
			out = append(out, piece)
			piece = ""
		}
		piece += string(character)
	}
	return append(out, piece)
}
