package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// logWriters names every file allowed to write to the run log, with the number
// of places in it that do, then says what each one writes.
//
// FR-065 promises that the log never holds a symptom, a note or any other part
// of the record. Nothing about a write to standard error announces whether its
// arguments came from the record, so the promise cannot be kept by inspecting
// the arguments; it is kept by there being few enough writers to read, all of
// them known. A new one fails here and has to be declared, which is the moment
// to ask what it puts in the file.
//
// The runlog package is not listed: it owns the log and writes the run's own
// lines, never the record's.
var logWriters = map[string]int{
	// The record could not be closed; the reason is the store's, not the user's.
	// A window with no focuser: a fixed sentence.
	"app.go": 2,
	// A bound method panicked: the recovered value and the stack. Every window
	// shares this guard (Amendment 21), so it lives with the facade.
	"internal/facade/facade.go": 1,
	// The record could not be opened; the window could not be run. A path
	// and an error each.
	//
	// keepLog writes one more line ("keeping the log: ...") straight to the
	// file it has just opened rather than through standard error, so this scan
	// cannot see it: the handle is a local and an AST knows nothing of types.
	// It is a constant sentence plus an error, written before the record is
	// opened, so there is nothing of the record in the program yet to leak.
	"main.go": 2,
	// Handing the keyboard to the page panicked: the recovered value and the
	// stack. It runs on a goroutine of its own, where a panic would otherwise
	// end the process with the window already open and nothing said.
	"window.go": 1,
}

// TestOnlyTheKnownPlacesWriteToTheLog keeps FR-065's promise checkable.
func TestOnlyTheKnownPlacesWriteToTheLog(t *testing.T) {
	root := repoRoot(t)
	found := map[string]int{}
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatalf("relative path of %s: %v", path, err)
		}
		relative = filepath.ToSlash(relative)
		if strings.HasPrefix(relative, "internal/infrastructure/runlog/") {
			continue
		}
		found[relative] += logWritesIn(t, path)
	}

	for name, count := range found {
		if count == 0 {
			continue
		}
		allowed, ok := logWriters[name]
		if !ok {
			t.Errorf("%s writes to the log in %d place(s) and is not declared in "+
				"logWriters: FR-065 promises the log holds no part of the record, "+
				"which is only checkable while every writer is known", name, count)
			continue
		}
		if count != allowed {
			t.Errorf("%s writes to the log in %d place(s), declared %d: a new write "+
				"is a decision about what reaches the file, so it is taken twice",
				name, count, allowed)
		}
	}

	var absent []string
	for name := range logWriters {
		if found[name] == 0 {
			absent = append(absent, name)
		}
	}
	sort.Strings(absent)
	for _, name := range absent {
		t.Errorf("logWriters declares %s, which no longer writes to the log: a "+
			"declaration nothing matches is a rule a reader will look for and not find", name)
	}
}

// logWritesIn counts the calls in one file that write to standard error, which
// is where the run log is pointed before anything else can fail.
func logWritesIn(t *testing.T, path string) int {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	count := 0
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, argument := range call.Args {
			if isStandardError(argument) {
				count++
				return true
			}
		}
		return true
	})
	return count
}

// isStandardError answers whether an expression is os.Stderr.
func isStandardError(node ast.Expr) bool {
	selector, ok := node.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Stderr" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "os"
}
