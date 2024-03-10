package storage_test

import (
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewDB_Success(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()

	db, err := storage.NewDB(cfg)

	assert.Nil(t, err)
	defer db.Close()
	_, err = db.Exec("SELECT 1;")
	assert.Nil(t, err)
}

func TestNewDB_UnsupportedDriver(t *testing.T) {
	t.Setenv("BOLT_DB_CONN_DRIVER", "abc123")
	cfg := bolttest.NewTestConnectionConfig()

	_, err := storage.NewDB(cfg)

	assert.ErrorIs(t, err, storage.ErrUnsupportedDriver)
}
