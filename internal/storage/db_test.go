package storage_test

import (
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestDBConnect_Success(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	conn, err := storage.DBConnect(cfg)
	assert.Nil(t, err)
	defer conn.Session.Close()

	_, err = conn.Session.SQL().Exec("SELECT 1;")
	assert.Nil(t, err)
}

func TestDBConnect_UnsupportedDriver(t *testing.T) {
	t.Setenv("BOLT_DB_CONN_DRIVER", "abc123")
	cfg := bolttest.NewTestConnectionConfig()

	_, err := storage.DBConnect(cfg)
	assert.ErrorIs(t, err, storage.ErrUnsupportedDriver)
}

func TestDBConnect_BadConnectionString(t *testing.T) {
	t.Setenv("BOLT_DB_CONN_DRIVER", "mysql")
	t.Setenv("BOLT_DB_CONN_HOST", "abc123")
	cfg := bolttest.NewTestConnectionConfig()

	_, err := storage.DBConnect(cfg)

	assert.ErrorIs(t, err, storage.ErrUnableToConnect)
	assert.ErrorContains(t, err, "unable to open connection to database")
}
