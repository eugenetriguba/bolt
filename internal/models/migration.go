package models

import (
	"fmt"
	"time"
)

type Migration struct {
	Version string
	Message string
	Applied bool
}

func NewMigration(timestamp time.Time, message string) *Migration {
	version := fmt.Sprintf(
		"%d%02d%02d%02d%02d%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second(),
	)
	return &Migration{Version: version, Message: message, Applied: false}
}

func (m *Migration) String() string {
	checkmark := " "
	if m.Applied {
		checkmark = "x"
	}

	message := m.Message
	if len(message) > 0 {
		message = fmt.Sprintf("- %s ", message)
	}

	return fmt.Sprintf(
		"%s %s- [%s]",
		m.Version,
		message,
		checkmark,
	)
}
