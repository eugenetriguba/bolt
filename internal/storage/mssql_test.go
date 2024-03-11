//go:build mssql

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
			expectedQuery: "SELECT id, name FROM some_table WHERE id = @p1;",
		},
		{
			query:         "SELECT id, name FROM some_table WHERE id = ? AND name = ?;",
			argCount:      2,
			expectedQuery: "SELECT id, name FROM some_table WHERE id = @p1 AND name = @p2;",
		},
	}
	for _, tc := range testCases {
		adapter := storage.MSSQLAdapter{}
		newQuery := adapter.ConvertGenericPlaceholders(tc.query, tc.argCount)
		assert.Equal(t, newQuery, tc.expectedQuery)
	}
}

func TestTableExists(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.MSSQLAdapter{}
	db, err := sql.Open("sqlserver", adapter.CreateDSN(cfg))
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
	adapter := storage.MSSQLAdapter{}
	db, err := sql.Open("sqlserver", adapter.CreateDSN(cfg))
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
		Driver:   "mssql",
		Host:     "db1",
		Port:     "5432",
		User:     "testuser",
		Password: "supersecretpassword",
		DBName:   "testdb",
	}
	adapter := storage.MSSQLAdapter{}

	cs := adapter.CreateDSN(cfg)

	expectedConnectionString := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&dial+timeout=0&disableretry=false",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	assert.Equal(t, cs, expectedConnectionString)
}

func TestCreateDSN_UnparsablePort(t *testing.T) {
	cfg := configloader.ConnectionConfig{
		Driver:   "mssql",
		Host:     "db1",
		Port:     "abc123",
		User:     "testuser",
		Password: "supersecretpassword",
		DBName:   "testdb",
	}
	adapter := storage.MSSQLAdapter{}

	cs := adapter.CreateDSN(cfg)

	expectedConnectionString := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&dial+timeout=0&disableretry=false",
		cfg.User, cfg.Password, cfg.Host, "1433", cfg.DBName,
	)
	assert.Equal(t, cs, expectedConnectionString)
}
