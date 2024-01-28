package models_test

import (
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewTimestampMigration(t *testing.T) {
	var timestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	m := models.NewTimestampMigration(timestamp, "test message")

	assert.Equal(t, m.Version, "20200101000000")
	assert.Equal(t, m.Message, "test message")
	assert.Equal(t, m.Applied, false)
}

func TestNewSequentialMigration(t *testing.T) {
	m := models.NewSequentialMigration(1, "test message")

	assert.Equal(t, m.Version, "001")
	assert.Equal(t, m.Message, "test message")
	assert.Equal(t, m.Applied, false)

	m = models.NewSequentialMigration(100, "test message")

	assert.Equal(t, m.Version, "100")
	assert.Equal(t, m.Message, "test message")
	assert.Equal(t, m.Applied, false)
}
