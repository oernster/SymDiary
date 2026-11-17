package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/product"
)

// DefaultPath answers where the record lives: %APPDATA%\SymDiary\symdiary.db
// (FR-060).
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding the folder for the record: %w", err)
	}
	return filepath.Join(base, product.Name, product.RecordFileName), nil
}

// Unavailable is the stand-in for a record that could not be opened. It moved
// to the application layer so the phone recorder can fall back to it without
// linking SQLite; the name stays here for the desktop's callers.
type Unavailable = application.Unavailable
