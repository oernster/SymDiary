package structural

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// siteFolder is what GitHub Pages publishes as symdiary.com, the phone
// recorder's pages under app/ included.
const siteFolder = "docs"

// foreignScript finds a script tag whose source names another site: a full
// address or one starting "//". A path on the site itself names no host.
var foreignScript = regexp.MustCompile(`(?is)<script\b[^>]*\bsrc\s*=\s*["']?\s*(?:https?:)?//[^"'\s>]*`)

func TestNoSitePageLoadsAScriptFromAnotherSite(t *testing.T) {
	t.Parallel()
	// The phone recorder keeps a person's symptoms in the browser's storage for
	// symdiary.com, which every script running on any page of the site can
	// read. So no page may run code fetched from anywhere else (NFR-PRIV-005).
	root := filepath.Join(repoRoot(t), siteFolder)
	pages := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".html") {
			return err
		}
		pages++
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, found := range foreignScript.FindAllString(string(text), -1) {
			relative, _ := filepath.Rel(root, path)
			t.Errorf("%s loads a script from another site: %s", filepath.ToSlash(relative), found)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("reading %s: %v", siteFolder, err)
	}
	if pages == 0 {
		t.Fatalf("no pages found under %s, so nothing was checked", siteFolder)
	}
}
