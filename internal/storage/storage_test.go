package storage_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/dbtest"
	"github.com/eugenetriguba/bolt/internal/storage"
)

func TestDBConnectWithPostgres(t *testing.T) {
	cfg, err := dbtest.NewTestConnectionConfig("postgres")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	if err != nil {
		t.Fatalf("could not connect to postgres db using %s: %s",
			storage.DBConnectionString(cfg),
			err,
		)
	}
	defer conn.Close()

	_, err = conn.Exec("SELECT 1;")
	if err != nil {
		t.Fatalf("could not execute query against postgres db: %s", err)
	}
}

func TestDBConnectMalformedConnectionString(t *testing.T) {
	_, err := storage.DBConnect("postgres", "pizza=123")
	if err == nil || (err != nil && !strings.Contains(err.Error(), "connection refused")) {
		t.Fatalf("expected connection error, got %s", err)
	}
}

func TestDBConnectUnsupportedDriver(t *testing.T) {
	cfg, err := dbtest.NewTestConnectionConfig("redis")
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.DBConnect(cfg.Driver, storage.DBConnectionString(cfg))
	if err == nil || (err != nil && !strings.Contains(err.Error(), `unknown driver "redis"`)) {
		t.Fatalf("expected bad driver, got %s", err)
	}
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
	expectedConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	cs := storage.DBConnectionString(&cfg)

	if cs != expectedConnectionString {
		t.Fatalf("Expected %s, got %s", expectedConnectionString, cs)
	}
}
