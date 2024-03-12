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
	db, err := storage.NewDB(connectionConfig)
	assert.Nil(t, err)

	DropTable(t, db, "bolt_migrations")
	DropTable(t, db, "tmp")
	t.Cleanup(func() {
		DropTable(t, db, "bolt_migrations")
		DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})

	return db
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

func DropTable(t *testing.T, db storage.DB, tableName string) {
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	assert.Nil(t, err)
}
