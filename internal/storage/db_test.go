package storage_test

import (
	"fmt"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate"
)

func TestDBConnectWithPostgres(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig(t, "postgres")

	conn, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	checkmate.AssertNil(t, err)
	defer conn.Close()

	_, err = conn.Exec("SELECT 1;")
	checkmate.AssertNil(t, err)
}

func TestDBConnectMalformedConnectionString(t *testing.T) {
	_, err := storage.DBConnect("postgres", "pizza=123")
	checkmate.AssertErrorContains(t, err, "connection refused")
}

func TestDBConnectUnsupportedDriver(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig(t, "redis")

	_, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	checkmate.AssertErrorContains(t, err, `unknown driver "redis"`)
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
	checkmate.AssertEqual(t, cs, expectedConnectionString)
}
