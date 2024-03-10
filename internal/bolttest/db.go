package bolttest

import (
	"fmt"
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func NewTestDB(t *testing.T) storage.DB {
	connectionConfig := NewTestConnectionConfig()
	testdb, err := storage.NewDB(connectionConfig)
	assert.Nil(t, err)
	t.Cleanup(func() {
		DropTable(t, testdb, "bolt_migrations")
		assert.Nil(t, testdb.Close())
	})
	return testdb
}

func NewTestConnectionConfig() configloader.ConnectionConfig {
	return configloader.ConnectionConfig{
		Driver:   os.Getenv("BOLT_DB_CONN_DRIVER"),
		DBName:   os.Getenv("BOLT_DB_CONN_DBNAME"),
		Host:     os.Getenv("BOLT_DB_CONN_HOST"),
		Port:     os.Getenv("BOLT_DB_CONN_PORT"),
		User:     os.Getenv("BOLT_DB_CONN_USER"),
		Password: os.Getenv("BOLT_DB_CONN_PASSWORD"),
	}
}

func DropTable(t *testing.T, testdb storage.DB, tableName string) {
	_, err := testdb.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	assert.Nil(t, err)
}
