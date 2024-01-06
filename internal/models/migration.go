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

func NewMigration(timestamp time.Time, message string) *Migration {
	version := fmt.Sprintf(
		"%d%02d%02d%02d%02d%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second(),
	)
	return &Migration{Version: version, Message: message, Applied: false}
}

func (m *Migration) NormalizedMessage() string {
	lowercaseMessage := strings.ToLower(m.Message)
	trimmedMessage := strings.TrimSpace(lowercaseMessage)
	return strings.ReplaceAll(trimmedMessage, " ", "_")
}

func (m *Migration) String() string {
	checked := " "
	if m.Applied {
		checked = "x"
	}

	message := m.NormalizedMessage()
	if len(message) > 0 {
		message = fmt.Sprintf("- %s ", message)
	}

	return fmt.Sprintf(
		"%s %s- [%s]",
		m.Version,
		message,
		checked,
	)
}
