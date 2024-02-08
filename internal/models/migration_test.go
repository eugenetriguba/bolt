package models_test

import (
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/checkmate/check"
)

func TestNewTimestampMigration(t *testing.T) {
	var timestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	m := models.NewTimestampMigration(timestamp, "test message")

	check.Equal(t, m.Version, "20200101000000")
	check.Equal(t, m.Message, "test message")
	check.Equal(t, m.Applied, false)
}

func TestNewSequentialMigration(t *testing.T) {
	m := models.NewSequentialMigration(1, "test message")

	check.Equal(t, m.Version, "001")
	check.Equal(t, m.Message, "test message")
	check.Equal(t, m.Applied, false)

	m = models.NewSequentialMigration(100, "test message")

	check.Equal(t, m.Version, "100")
	check.Equal(t, m.Message, "test message")
	check.Equal(t, m.Applied, false)
}
