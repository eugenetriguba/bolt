package models_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/eugenetriguba/bolt/internal/models"
)

var timestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNewMigrationInit(t *testing.T) {
	m := models.NewMigration(timestamp, "test message")

	assert.Equal(t, m.Version, "20200101000000")
	assert.Equal(t, m.Message, "test message")
	assert.Equal(t, m.Applied, false)
}

func TestAppliedNewMigration(t *testing.T) {
	m := models.NewMigration(timestamp, "test_message")

	assert.Equal(t, m.String(), "20200101000000 - test_message - [ ]")
	m.Applied = true
	assert.Equal(t, m.String(), "20200101000000 - test_message - [x]")
}

func TestNewMigrationMessageIsNormalized(t *testing.T) {
	m := models.NewMigration(timestamp, "  test MESSAGE  ")
	assert.Equal(t, m.NormalizedMessage(), "test_message")
}

func TestEmptyMessageMigration(t *testing.T) {
	m := models.NewMigration(timestamp, "")

	assert.Equal(t, "", m.NormalizedMessage())
	assert.Equal(t, "20200101000000 - [ ]", m.String())
}
