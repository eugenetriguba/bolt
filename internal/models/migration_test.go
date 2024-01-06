package models_test

import (
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/models"
)

var ts = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func TestMigration(t *testing.T) {
	type test struct {
		inputTs         time.Time
		inputMessage    string
		expectedVersion string
		expectedDirname string
		expectedString  string
	}
	tests := []test{
		// Ensure message is trimmed of spaces, lowercased, and
		// using underscores for spaces in normalized version.
		{
			inputTs:         ts,
			inputMessage:    "  test MESSAGE  ",
			expectedVersion: "20200101000000",
			expectedDirname: "20200101000000_test_message",
			expectedString:  "20200101000000 - test_message - [ ]",
		},
		// Ensure an empty message means there is no message shown for
		// dirname and string.
		{
			inputTs:         ts,
			inputMessage:    "",
			expectedVersion: "20200101000000",
			expectedDirname: "20200101000000_",
			expectedString:  "20200101000000 - [ ]",
		}
	}

	for _, tc := range tests {
		m := models.NewMigration(tc.inputTs, tc.inputMessage)
		if m.Version != tc.expectedVersion {
			t.Fatalf("got: %v, expected: %v", m.Version, tc.expectedVersion)
		}
		if m.Message != tc.inputMessage {
			t.Fatalf("got: %v, expected: %v", m.Message, tc.inputMessage)
		}
		if m.Applied != false {
			t.Fatalf("got: %v, expected: %v", m.Applied, false)
		}
		if m.Dirname() != tc.expectedDirname {
			t.Fatalf("got: %v, expected: %v", m.Dirname(), tc.expectedDirname)
		}
		if m.String() != tc.expectedString {
			t.Fatalf("got: %v, expected: %v", m.String(), tc.expectedString)
		}
	}
}

func TestAppliedNewMigration(t *testing.T) {
	m := models.NewMigration(ts, "test message")
	m.Applied = true

	expectedString := "20200101000000 - test_message - [x]"
	if m.String() != expectedString {
		t.Fatalf("got: %v, expected: %v", m.String(), expectedString)
	}
}
