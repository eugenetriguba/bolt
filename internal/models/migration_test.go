package models_test

import (
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/checkmate"
)

func TestNewTimestampMigration(t *testing.T) {
	var timestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	m := models.NewTimestampMigration(timestamp, "test message")

	checkmate.AssertEqual(t, m.Version, "20200101000000")
	checkmate.AssertEqual(t, m.Message, "test message")
	checkmate.AssertEqual(t, m.Applied, false)
}

func TestNewSequentialMigration(t *testing.T) {
	m := models.NewSequentialMigration(1, "test message")

	checkmate.AssertEqual(t, m.Version, "001")
	checkmate.AssertEqual(t, m.Message, "test message")
	checkmate.AssertEqual(t, m.Applied, false)

	m = models.NewSequentialMigration(100, "test message")

	checkmate.AssertEqual(t, m.Version, "100")
	checkmate.AssertEqual(t, m.Message, "test message")
	checkmate.AssertEqual(t, m.Applied, false)
}
