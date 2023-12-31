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

func NewMigration(message string) *Migration {
	now := time.Now()
	version := fmt.Sprintf(
		"%d%02d%02d%02d%02d%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
	)
	return &Migration{Version: version, Message: message, Applied: false}
}

func (m *Migration) Dirname() string {
	return fmt.Sprintf("%s_%s", m.Version, m.normalizedMessage())
}

func (m *Migration) normalizedMessage() string {
	return strings.ReplaceAll(
		strings.TrimSpace(strings.ToLower(m.Message)),
		" ",
		"_",
	)
}

func (m *Migration) String() string {
	checked := " "
	if m.Applied {
		checked = "x"
	}

	return fmt.Sprintf(
		"%s - %s - [%s]",
		m.Version,
		m.normalizedMessage(),
		checked,
	)
}
