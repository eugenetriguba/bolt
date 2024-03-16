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

func TestName(t *testing.T) {
	type test struct {
		migration    *models.Migration
		expectedName string
	}

	testCases := []test{
		// Underscores are used to separate version and message.
		{
			migration:    models.NewSequentialMigration(1, "test_message"),
			expectedName: "001_test_message",
		},
		{
			migration:    models.NewTimestampMigration(time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC), "test_message"),
			expectedName: "12340102030405_test_message",
		},
		// Leading/Trailing spaces trimmed, lowercased, and spaces converted to underscores.
		{
			migration:    models.NewSequentialMigration(1, "  teST mesSAGE  "),
			expectedName: "001_test_message",
		},
		{
			migration:    models.NewTimestampMigration(time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC), "  teST mesSAGE  "),
			expectedName: "12340102030405_test_message",
		},
	}

	for _, tc := range testCases {
		check.Equal(t, tc.migration.Name(), tc.expectedName)
	}
}
