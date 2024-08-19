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
	databaseConfig := NewDatabaseConfig()
	db, err := storage.NewDB(databaseConfig)
	assert.Nil(t, err)

	DropTable(t, db, databaseConfig.MigrationsTable)
	DropTable(t, db, "tmp")
	t.Cleanup(func() {
		DropTable(t, db, databaseConfig.MigrationsTable)
		DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})

	return db
}

func NewDatabaseConfig() configloader.DatabaseConfig {
	return configloader.DatabaseConfig{
		DSN:             os.Getenv("BOLT_DB_DSN"),
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
