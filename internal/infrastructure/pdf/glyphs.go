package pdf

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"

	"github.com/oernster/symdiary/internal/domain"
)

// ErrUnprintable refuses a record holding a character the typeface carried in
// the binary has no glyph for.
//
// The Go fonts cover Latin, Greek and Cyrillic. A character outside them used to
// be drawn as the font's empty box, so a note written in Chinese, Arabic or with
// an emoji reached the reader's doctor as a row of boxes while the window showed
// it whole. A sheet that silently drops the user's words is the one thing the
// document must never be, so it is refused instead, naming what it cannot print.
var ErrUnprintable = errors.New("the record holds characters the document's typeface cannot print")

// mostNamed is how many unprintable characters a refusal names before it counts
// the rest: enough to recognise the script, short enough to read.
const mostNamed = 10

// Check refuses a record holding a character the typeface cannot draw, in the
// weight its line is set in. White space is never drawn as a glyph, so it is not
// asked about.
func (Sheet) Check(lines []domain.Line) error {
	return checkGlyphs(lines, goregular.TTF, gobold.TTF)
}

// checkGlyphs is Check against the typeface's two weights as bytes. A typeface
// that cannot be read is reported rather than taken to cover everything.
func checkGlyphs(lines []domain.Line, regularTTF, boldTTF []byte) error {
	regular, err := sfnt.Parse(regularTTF)
	if err != nil {
		return fmt.Errorf("reading the record's typeface: %w", err)
	}
	bold, err := sfnt.Parse(boldTTF)
	if err != nil {
		return fmt.Errorf("reading the record's typeface: %w", err)
	}
	var buffer sfnt.Buffer
	seen := map[rune]bool{}
	var missing []rune
	for _, line := range lines {
		face := regular
		if styles[line.Kind].bold {
			face = bold
		}
		for _, character := range line.Text {
			if unicode.IsSpace(character) || seen[character] {
				continue
			}
			seen[character] = true
			if index, err := face.GlyphIndex(&buffer, character); err == nil && index != 0 {
				continue
			}
			missing = append(missing, character)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnprintable, named(missing))
}

// named lists the first characters a refusal names, counting any beyond them.
func named(missing []rune) string {
	shown := missing
	if len(shown) > mostNamed {
		shown = shown[:mostNamed]
	}
	parts := make([]string, 0, len(shown))
	for _, character := range shown {
		parts = append(parts, string(character))
	}
	text := strings.Join(parts, " ")
	if rest := len(missing) - len(shown); rest > 0 {
		text += fmt.Sprintf(" (and %d more)", rest)
	}
	return text
}
