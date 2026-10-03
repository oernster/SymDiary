// Package export writes the whole record to a JSON file and reads one back
// (FR-050 to FR-053). The file is the user's copy of their record: plain,
// versioned and readable by anything that reads JSON.
package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// formatName marks a file as a SymDiary export.
const formatName = "symdiary-record"

// formatVersion is the version this code writes. It reads this one and every
// older one, which versions.go states as a table and a test enforces: raising
// this constant without adding a reader and a sample fails the suite.
const formatVersion = 1

// maxFileBytes caps what a read will take in. An export of the 20,000 events
// the specification plans for (A-1) is a few megabytes; a file sixteen times
// larger than any plausible record is refused before it is read, never
// allocated (house robustness rule 5).
const maxFileBytes = 64 << 20

// timeLayout is how instants are written: RFC 3339 with the offset (FR-051).
const timeLayout = time.RFC3339Nano

// tempPattern names the temporary file a write goes through.
const tempPattern = ".symdiary-export-*.tmp"

// The refusals a read can answer.
var (
	ErrNotAnExport = errors.New("not a SymDiary export")
	ErrNewerFormat = errors.New("written by a newer SymDiary")
	ErrTooLarge    = errors.New("larger than any SymDiary export")
	// ErrNoVersion refuses a file that claims to be an export but names no
	// version SymDiary ever wrote. Reading one as though it were the current
	// format is how a hand-edited or third-party file gets to decide what a
	// medical record says.
	ErrNoVersion = errors.New("names no SymDiary export version")
)

// fileShape is the file's JSON form.
type fileShape struct {
	Format      string            `json:"format"`
	Version     int               `json:"version"`
	Definitions []definitionShape `json:"definitions"`
	Events      []eventShape      `json:"events"`
}

type definitionShape struct {
	Label    string `json:"label"`
	LastUsed string `json:"lastUsed"`
}

type eventShape struct {
	Symptom    string `json:"symptom"`
	OccurredAt string `json:"occurredAt"`
	RecordedAt string `json:"recordedAt"`
	Severity   string `json:"severity"`
	Note       string `json:"note"`
}

// File is the JSON export format.
type File struct{}

// Write writes the record to path through a temporary file in the same folder,
// replaced in one rename, so a failure leaves no partial file (FR-052).
func (File) Write(path string, record application.Record) error {
	shape := fileShape{Format: formatName, Version: formatVersion}
	for _, definition := range record.Definitions {
		shape.Definitions = append(shape.Definitions, definitionShape{
			Label: definition.Label, LastUsed: definition.LastUsed.Format(timeLayout),
		})
	}
	for _, event := range record.Events {
		shape.Events = append(shape.Events, eventShape{
			Symptom:    event.Symptom,
			OccurredAt: event.OccurredAt.Format(timeLayout),
			RecordedAt: event.RecordedAt.Format(timeLayout),
			Severity:   event.Severity.String(),
			Note:       event.Note,
		})
	}
	encoded, err := json.MarshalIndent(shape, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomically(path, encoded)
}

// writeAtomically writes data to path by way of a temporary file.
func writeAtomically(path string, data []byte) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), tempPattern)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = temp.Close()
			_ = os.Remove(temp.Name())
		}
	}()
	if _, err = temp.Write(data); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

// Read reads a record from path, refusing anything that is not a SymDiary export
// this version can read.
func (File) Read(path string) (application.Record, error) {
	opened, err := os.Open(path)
	if err != nil {
		return application.Record{}, err
	}
	defer opened.Close()
	raw, err := io.ReadAll(io.LimitReader(opened, maxFileBytes+1))
	if err != nil {
		return application.Record{}, err
	}
	if len(raw) > maxFileBytes {
		return application.Record{}, ErrTooLarge
	}
	return readVersioned(raw)
}

// readVersioned reads the envelope, then hands the file to the reader for the
// version it declares. Every refusal names the version it read, so a user told
// their export cannot be opened can see why.
func readVersioned(raw []byte) (application.Record, error) {
	var declared envelope
	if err := json.Unmarshal(raw, &declared); err != nil || declared.Format != formatName {
		return application.Record{}, ErrNotAnExport
	}
	if declared.Version < firstVersion {
		return application.Record{}, fmt.Errorf("%w (format %d)", ErrNoVersion, declared.Version)
	}
	if declared.Version > formatVersion {
		return application.Record{}, fmt.Errorf("%w (format %d)", ErrNewerFormat, declared.Version)
	}
	read, known := readers[declared.Version]
	if !known {
		// Unreachable while the guard in versions_test.go holds. It is answered
		// rather than assumed: a gap in the table must refuse the file, never
		// read it as some other version.
		return application.Record{}, fmt.Errorf("%w (format %d)", ErrNoVersion, declared.Version)
	}
	return read(raw)
}

// record converts the file's shape to a record, refusing any unreadable field and any blank symptom,
// whether it arrives on an event or on a definition kept with none.
func (shape fileShape) record() (application.Record, error) {
	var record application.Record
	for i, definition := range shape.Definitions {
		if err := domain.CheckLabel(definition.Label); err != nil {
			return application.Record{}, fmt.Errorf("definition %d: %w", i+1, err)
		}
		used, err := time.Parse(timeLayout, definition.LastUsed)
		if err != nil {
			return application.Record{}, fmt.Errorf("definition %d: %w", i+1, err)
		}
		record.Definitions = append(record.Definitions, domain.Definition{
			Label: definition.Label, LastUsed: used,
		})
	}
	for i, event := range shape.Events {
		converted, err := event.event()
		if err != nil {
			return application.Record{}, fmt.Errorf("event %d: %w", i+1, err)
		}
		record.Events = append(record.Events, converted)
	}
	return record, nil
}

// event converts one event's shape, refusing a blank symptom, an unreadable
// time or an unknown severity.
func (shape eventShape) event() (domain.Event, error) {
	if err := domain.CheckLabel(shape.Symptom); err != nil {
		return domain.Event{}, err
	}
	occurred, err := time.Parse(timeLayout, shape.OccurredAt)
	if err != nil {
		return domain.Event{}, err
	}
	recorded, err := time.Parse(timeLayout, shape.RecordedAt)
	if err != nil {
		return domain.Event{}, err
	}
	severity, err := domain.ParseSeverity(shape.Severity)
	if err != nil {
		return domain.Event{}, err
	}
	return domain.Event{
		Symptom: shape.Symptom, OccurredAt: occurred, RecordedAt: recorded,
		Severity: severity, Note: shape.Note,
	}, nil
}
