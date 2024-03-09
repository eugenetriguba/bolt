package storage_test

import (
	"fmt"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestDBConnect_Success(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig(t, "postgres")

	conn, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	assert.Nil(t, err)
	defer conn.Close()

	_, err = conn.Exec("SELECT 1;")
	assert.Nil(t, err)
}

func TestDBConnect_BadConnectionString(t *testing.T) {
	_, err := storage.DBConnect("postgres", "pizza=123")
	assert.ErrorIs(t, err, storage.ErrUnableToConnect)
	assert.ErrorContains(t, err, "unable to open connection to database")
}

func TestDBConnect_UnsupportedDriver(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig(t, "redis")

	_, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	assert.ErrorIs(t, err, storage.ErrUnsupportedDriver)
}

func TestDBConnectionString(t *testing.T) {
	cfg := configloader.ConnectionConfig{
		Driver:   "postgres",
		Host:     "db1",
		Port:     5432,
		User:     "testuser",
		Password: "supersecretpassword",
		DBName:   "testdb",
	}

	cs := storage.DBConnectionString(&cfg)

	expectedConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
	assert.Equal(t, cs, expectedConnectionString)
}
