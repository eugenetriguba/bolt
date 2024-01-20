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
