package application

import "github.com/oernster/symdiary/internal/product"

// Credit names one open-source work SymDiary is built with, with its licence as
// its own LICENSE file states it.
type Credit struct {
	Work    string
	Licence string
	Holder  string
}

// About is what the About dialog shows (FR-066, FR-067).
type About struct {
	Name      string
	Version   string
	Author    string
	Copyright string
	Statement string
	Credits   []Credit
}

// Each build credits exactly the works it ships (FR-066, FR-106). Build tools
// that ship nothing (Vite, TypeScript, the test runners) are not credited. Each
// licence was read from the work's own LICENSE file; the PDF library and the Go
// fonts on 2026-11-17, the rest on 2026-09-22.
var (
	// goItself is Go's runtime and standard library, in every build; the phone
	// also ships Go's wasm_exec.js, under the same licence.
	goItself = Credit{Work: "Go", Licence: "BSD 3-Clause", Holder: "The Go Authors"}
	// react draws the page in every build.
	react = Credit{Work: "React", Licence: "MIT", Holder: "Facebook, Inc. and its affiliates"}

	desktopOnly = []Credit{
		{Work: "golang.org/x/sys", Licence: "BSD 3-Clause", Holder: "The Go Authors"},
		{Work: "Wails", Licence: "MIT", Holder: "Lea Anthony"},
		{Work: "modernc.org/sqlite", Licence: "BSD 3-Clause", Holder: "The Sqlite Authors"},
		{Work: "SQLite", Licence: "Public domain", Holder: "D. Richard Hipp and contributors"},
		{Work: "go-pdf/fpdf", Licence: "MIT", Holder: "The Go-PDF authors and David Barnes"},
		{Work: "Go fonts (golang.org/x/image)", Licence: "BSD 3-Clause", Holder: "Bigelow & Holmes Inc. and The Go Authors"},
	}
)

// DesktopCredits answers the works the desktop application ships.
func DesktopCredits() []Credit {
	return append(append([]Credit{goItself}, desktopOnly...), react)
}

// PhoneCredits answers the works the phone recorder ships: Go compiled to
// WebAssembly and the page, with no window toolkit, database or PDF.
func PhoneCredits() []Credit {
	return []Credit{goItself, react}
}

// NewAbout answers the About content for a version and the works that build
// ships. The credits are copied, so no caller can change another's.
func NewAbout(version string, credits []Credit) About {
	return About{
		Name:      product.Name,
		Version:   version,
		Author:    product.Author,
		Copyright: product.Copyright,
		Statement: product.Statement,
		Credits:   append([]Credit(nil), credits...),
	}
}
