package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/symdiary/internal/product"
)

func TestOpensAtTheGivenPath(t *testing.T) {
	base := t.TempDir()
	t.Setenv("APPDATA", base)
	path, err := DefaultPath()
	if want := filepath.Join(base, "SymDiary", "symdiary.db"); err != nil || path != want {
		t.Errorf("DefaultPath = %q, %v; want %q", path, err, want)
	}
	t.Setenv("APPDATA", "")
	if _, err := DefaultPath(); err == nil {
		t.Error("with no APPDATA there is no folder for the record")
	}
}

// TestTheRecordSitsInTheFolderThePlatformNames pins the one fact every
// packaged claim about the record's whereabouts rests on: the record is the
// product's own folder inside whatever the platform calls the place for a
// user's configuration; nothing else decides it.
//
// That is what makes the Flatpak's promise true without a second code path.
// Flatpak redirects XDG_CONFIG_HOME into the sandbox, os.UserConfigDir reads
// it on Linux, so the record lands at
// ~/.var/app/uk.codecrafter.SymDiary/config/SymDiary/symdiary.db, which is the
// path build_flatpak.sh prints when it finishes. A change here that reached
// for the home directory instead would move the record on Linux and macOS
// while leaving Windows looking correct, so it is asserted rather than
// assumed.
func TestTheRecordSitsInTheFolderThePlatformNames(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("this machine names no configuration folder: %v", err)
	}
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if want := filepath.Join(base, product.Name, product.RecordFileName); path != want {
		t.Errorf("DefaultPath = %q, want %q: the record belongs in the folder "+
			"the platform names, which is what the Flatpak's redirect relies on", path, want)
	}
}
