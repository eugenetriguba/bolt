//go:build mysql

package storage_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestConvertGenericPlaceholders(t *testing.T) {
	type test struct {
		query         string
		argCount      int
		expectedQuery string
	}

	testCases := []test{
		{
			query:         "SELECT id, name FROM some_table;",
			argCount:      0,
			expectedQuery: "SELECT id, name FROM some_table;",
		},
		{
			query:         "SELECT id, name FROM some_table WHERE id = ?;",
			argCount:      1,
			expectedQuery: "SELECT id, name FROM some_table WHERE id = ?;",
		},
		{
			query:         "SELECT id, name FROM some_table WHERE id = ? AND name = ?;",
			argCount:      2,
			expectedQuery: "SELECT id, name FROM some_table WHERE id = ? AND name = ?;",
		},
	}
	for _, tc := range testCases {
		adapter := storage.MySQLAdapter{}
		newQuery := adapter.ConvertGenericPlaceholders(tc.query, tc.argCount)
		assert.Equal(t, newQuery, tc.expectedQuery)
	}
}

func TestTableExists(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.MySQLAdapter{}
	db, err := sql.Open("mysql", adapter.CreateDSN(cfg))
	assert.Nil(t, err)
	t.Cleanup(func() {
		_, err = db.Exec("DROP TABLE IF EXISTS tmp;")
		assert.Nil(t, err)
		assert.Nil(t, db.Close())
	})

	exists, err := adapter.TableExists(db, "tmp")
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = db.Exec("CREATE TABLE tmp(id INT PRIMARY KEY);")
	assert.Nil(t, err)

	exists, err = adapter.TableExists(db, "tmp")
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestDatabaseName(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.MySQLAdapter{}
	db, err := sql.Open("mysql", adapter.CreateDSN(cfg))
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, db.Close())
	})

	name, err := adapter.DatabaseName(db)
	assert.Nil(t, err)

	assert.Equal(t, name, cfg.DBName)
}

func TestCreateDSN(t *testing.T) {
	cfg := configloader.ConnectionConfig{
		Driver:   "mysql",
		Host:     "db1",
		Port:     "5432",
		User:     "testuser",
		Password: "supersecretpassword",
		DBName:   "testdb",
	}
	adapter := storage.PostgresqlAdapter{}

	cs := adapter.CreateDSN(cfg)

	expectedConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
	assert.Equal(t, cs, expectedConnectionString)
}
