package models

import (
	"fmt"
	"strings"
	"time"
)

type Migration struct {
	Version string
	Message string
	Applied bool
}

func NewTimestampMigration(version time.Time, message string) *Migration {
	return &Migration{
		Version: fmt.Sprintf(
			"%d%02d%02d%02d%02d%02d",
			version.Year(), version.Month(), version.Day(),
			version.Hour(), version.Minute(), version.Second(),
		),
		Message: message,
		Applied: false,
	}
}

func NewSequentialMigration(version uint64, message string) *Migration {
	return &Migration{
		Version: fmt.Sprintf("%03d", version),
		Message: message,
		Applied: false,
	}
}

// NormalizedMessage gives back the Message in lowercase,
// leading and trailing whitespace removed, and spaces
// replaced with underscores.
func (m Migration) NormalizedMessage() string {
	message := strings.ToLower(m.Message)
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, " ", "_")
	return message
}

func (m *Migration) Name() string {
	return fmt.Sprintf("%s_%s", m.Version, m.NormalizedMessage())
}
