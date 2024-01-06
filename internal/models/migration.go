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

func NewMigration(ts time.Time, message string) *Migration {
	version := fmt.Sprintf(
		"%d%02d%02d%02d%02d%02d", ts.Year(), ts.Month(), ts.Day(),
		ts.Hour(), ts.Minute(), ts.Second(),
	)
	return &Migration{Version: version, Message: message, Applied: false}
}

func (m *Migration) Dirname() string {
	return fmt.Sprintf("%s_%s", m.Version, m.normalizedMessage())
}

func (m *Migration) normalizedMessage() string {
	lowercaseMessage := strings.ToLower(m.Message)
	trimmedMessage := strings.TrimSpace(lowercaseMessage)
	return strings.ReplaceAll(trimmedMessage, " ", "_")
}

func (m *Migration) String() string {
	checked := " "
	if m.Applied {
		checked = "x"
	}

	message := m.normalizedMessage()
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
