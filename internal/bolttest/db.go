package bolttest

import (
	"database/sql"
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

	DropTable(t, db, connectionConfig.MigrationsTable)
	DropTable(t, db, "tmp")
	t.Cleanup(func() {
		DropTable(t, db, connectionConfig.MigrationsTable)
		DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})

	return db
}

func NewTestConnectionConfig() configloader.ConnectionConfig {
	return configloader.ConnectionConfig{
		Driver:          os.Getenv("BOLT_DB_DRIVER"),
		DBName:          os.Getenv("BOLT_DB_NAME"),
		Host:            os.Getenv("BOLT_DB_HOST"),
		Port:            os.Getenv("BOLT_DB_PORT"),
		User:            os.Getenv("BOLT_DB_USER"),
		Password:        os.Getenv("BOLT_DB_PASSWORD"),
		MigrationsTable: os.Getenv("BOLT_DB_MIGRATIONS_TABLE"),
	}
}

func DropTable(t *testing.T, db storage.DB, tableName string) {
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	assert.Nil(t, err)
}

type MockDB struct {
	ExecFunc        func(query string, args ...interface{}) (sql.Result, error)
	QueryFunc       func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFunc    func(query string, args ...interface{}) *sql.Row
	TxFunc          func(fn storage.TxFunc) error
	CloseFunc       func() error
	TableExistsFunc func(tableName string) (bool, error)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args...)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.QueryFunc(query, args...)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.QueryRowFunc(query, args...)
}

func (m *MockDB) Tx(fn storage.TxFunc) error {
	return m.TxFunc(fn)
}

func (m *MockDB) Close() error {
	return m.CloseFunc()
}

func (m *MockDB) TableExists(tableName string) (bool, error) {
	return m.TableExistsFunc(tableName)
}
