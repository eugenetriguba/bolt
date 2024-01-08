package storage_test

import (
	"fmt"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"gotest.tools/v3/assert"
)

func TestDBConnectWithPostgres(t *testing.T) {
	cfg, err := bolttest.NewTestConnectionConfig("postgres")
	assert.NilError(t, err)

	conn, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	assert.NilError(t, err)
	defer conn.Close()

	_, err = conn.Exec("SELECT 1;")
	assert.NilError(t, err)
}

func TestDBConnectMalformedConnectionString(t *testing.T) {
	_, err := storage.DBConnect("postgres", "pizza=123")
	assert.ErrorContains(t, err, "connection refused")
}

func TestDBConnectUnsupportedDriver(t *testing.T) {
	cfg, err := bolttest.NewTestConnectionConfig("redis")
	assert.NilError(t, err)

	_, err = storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	assert.ErrorContains(t, err, `unknown driver "redis"`)
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
