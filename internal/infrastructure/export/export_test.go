package export

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// bst is British Summer Time as a fixed offset, so the sample does not depend on
// the machine's zone data.
var bst = time.FixedZone("", 3600)

// sampleRecord answers the record the committed sample file holds.
func sampleRecord() application.Record {
	when := time.Date(2026, time.September, 22, 17, 12, 0, 0, bst)
	return application.Record{
		Definitions: []domain.Definition{{ID: 1, Label: "Tired", LastUsed: when}},
		Events: []domain.Event{{
			ID: 1, Definition: 1, Symptom: "Tired", OccurredAt: when, RecordedAt: when,
			Severity: domain.SeverityMild, Note: "Only been awake for about 10 minutes.",
		}},
	}
}

func TestFormatIsStable(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.json")
	if err := (File{}).Write(path, sampleRecord()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	written, _ := os.ReadFile(path)
	want, err := os.ReadFile(filepath.Join("testdata", "sample.json"))
	if err != nil {
		t.Fatalf("reading the sample: %v", err)
	}
	if strings.ReplaceAll(string(want), "\r\n", "\n") != string(written) {
		t.Errorf("the format changed. Wrote:\n%s", written)
	}
}

func TestExportWritesEveryEventAndReadsBack(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "My Exports.json")
	if err := (File{}).Write(path, sampleRecord()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	read, err := (File{}).Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(read.Events) != 1 || !domain.SameEvent(read.Events[0], sampleRecord().Events[0]) {
		t.Errorf("read back %+v", read.Events)
	}
	if len(read.Definitions) != 1 || read.Definitions[0].Label != "Tired" {
		t.Errorf("definitions read back %+v", read.Definitions)
	}
}

func TestFailedExportLeavesNothing(t *testing.T) {
	t.Parallel()
	folder := t.TempDir()
	// The target is a folder, so the final rename fails after the temporary
	// file has been written.
	target := filepath.Join(folder, "taken")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "inside"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := (File{}).Write(target, sampleRecord()); err == nil {
		t.Fatal("writing over a folder succeeded")
	}
	entries, _ := os.ReadDir(folder)
	if len(entries) != 1 {
		t.Errorf("a failed export left files behind: %v", entries)
	}
	if err := (File{}).Write(filepath.Join(folder, "missing", "out.json"), sampleRecord()); err == nil {
		t.Error("writing into a missing folder succeeded")
	}
}

func TestReadRefusals(t *testing.T) {
	t.Parallel()
	folder := t.TempDir()
	valid := `"format":"symdiary-record","version":1`
	cases := map[string]struct {
		body string
		want error
	}{
		"not json":      {`{`, ErrNotAnExport},
		"other format":  {`{"format":"other","version":1}`, ErrNotAnExport},
		"newer":         {`{"format":"symdiary-record","version":2}`, ErrNewerFormat},
		"blank symptom": {`{` + valid + `,"events":[{"symptom":" "}]}`, domain.ErrBlankSymptom},
		// A definition is refused on the same rule as an event: a symptom kept with no events would
		// otherwise arrive as a blank entry in the symptoms list.
		"blank definition": {`{` + valid + `,"definitions":[{"label":"   ","lastUsed":"2026-09-22T17:12:00Z"}]}`,
			domain.ErrBlankSymptom},
		"bad severity": {`{` + valid + `,"events":[{"symptom":"Tired","occurredAt":"2026-09-22T17:12:00+01:00",` +
			`"recordedAt":"2026-09-22T17:12:00+01:00","severity":"awful"}]}`, domain.ErrUnknownSeverity},
	}
	for name, c := range cases {
		path := filepath.Join(folder, name+".json")
		if err := os.WriteFile(path, []byte(c.body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := (File{}).Read(path); !errors.Is(err, c.want) {
			t.Errorf("%s: Read = %v, want %v", name, err, c.want)
		}
	}
	for name, body := range map[string]string{
		"bad occurred": `{` + valid + `,"events":[{"symptom":"Tired","occurredAt":"noon"}]}`,
		"bad recorded": `{` + valid + `,"events":[{"symptom":"Tired","occurredAt":"2026-09-22T17:12:00Z","recordedAt":"x"}]}`,
		"bad last use": `{` + valid + `,"definitions":[{"label":"Tired","lastUsed":"x"}]}`,
	} {
		path := filepath.Join(folder, name+".json")
		_ = os.WriteFile(path, []byte(body), 0o600)
		if _, err := (File{}).Read(path); err == nil {
			t.Errorf("%s: an unreadable time was accepted", name)
		}
	}
	if _, err := (File{}).Read(filepath.Join(folder, "absent.json")); err == nil {
		t.Error("a missing file read without complaint")
	}
}

func TestOversizedFileIsRefusedBeforeReading(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "huge.json")
	handle, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// A sparse file: the size is set without writing the bytes.
	if err := handle.Truncate(maxFileBytes + 1); err != nil {
		t.Fatal(err)
	}
	_ = handle.Close()
	if _, err := (File{}).Read(path); !errors.Is(err, ErrTooLarge) {
		t.Errorf("an oversized file = %v, want ErrTooLarge", err)
	}
}
