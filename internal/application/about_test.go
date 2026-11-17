package application

import (
	"slices"
	"strings"
	"testing"

	"github.com/oernster/symdiary/internal/product"
)

func TestAboutCarriesTheStatement(t *testing.T) {
	t.Parallel()
	about := NewAbout("1.2.3", DesktopCredits())
	if about.Name != product.Name || about.Version != "1.2.3" || about.Author != product.Author {
		t.Errorf("about = %+v", about)
	}
	if !strings.Contains(about.Statement, "no medical advice") {
		t.Errorf("the statement does not say it gives no medical advice: %q", about.Statement)
	}
	// The notice is written out here rather than compared against the constant,
	// which would assert nothing: the symbol and the year are the claim.
	if about.Copyright != "© Oliver Ernster 2026" {
		t.Errorf("the copyright reads %q", about.Copyright)
	}
	credits := DesktopCredits()
	about = NewAbout("1.2.3", credits)
	about.Credits[0].Work = "changed"
	if credits[0].Work == "changed" || DesktopCredits()[0].Work == "changed" {
		t.Error("a caller can change another's credit list")
	}
}

// works answers the names a credit list gives.
func works(credits []Credit) []string {
	names := make([]string, 0, len(credits))
	for _, credit := range credits {
		names = append(names, credit.Work)
	}
	return names
}

// The lists are written out here so a credit dropped or added by mistake fails:
// each name is a claim about what that build ships, measured against go.mod,
// the imports and package.json.
func TestEachBuildCreditsWhatItShips(t *testing.T) {
	t.Parallel()
	desktop := []string{
		"Go", "golang.org/x/sys", "Wails", "modernc.org/sqlite", "SQLite",
		"go-pdf/fpdf", "Go fonts (golang.org/x/image)", "React",
	}
	if got := works(DesktopCredits()); !slices.Equal(got, desktop) {
		t.Errorf("the desktop credits %q, want %q", got, desktop)
	}
	phone := []string{"Go", "React"}
	if got := works(PhoneCredits()); !slices.Equal(got, phone) {
		t.Errorf("the phone credits %q, want %q", got, phone)
	}
	for _, credit := range append(DesktopCredits(), PhoneCredits()...) {
		if credit.Licence == "" || credit.Holder == "" {
			t.Errorf("%s is credited without its licence or holder", credit.Work)
		}
	}
}
