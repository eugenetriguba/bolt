package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type FakeClock struct{}

func (c *FakeClock) Now() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestNewMigration(t *testing.T) {
	m := NewMigration(&FakeClock{}, "  test MESSAGE  ")

	assert.Equal(t, "20200101000000", m.Version)
	assert.Equal(t, "  test MESSAGE  ", m.Message)
	assert.Equal(t, false, m.Applied)
	assert.Equal(t, "20200101000000_test_message", m.Dirname())
	assert.Equal(t, "20200101000000 - test_message - [ ]", m.String())
}

func TestAppliedNewMigration(t *testing.T) {
	m := NewMigration(&FakeClock{}, "test message")
	m.Applied = true

	assert.Equal(t, "20200101000000 - test_message - [x]", m.String())
}

func TestEmptyMessageMigration(t *testing.T) {
	m := NewMigration(&FakeClock{}, "")

	assert.Equal(t, "20200101000000 - [ ]", m.String())
}
