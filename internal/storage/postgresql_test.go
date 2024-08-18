//go:build postgresql

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

func TestPostgresql_ConvertGenericPlaceholders(t *testing.T) {
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
			expectedQuery: "SELECT id, name FROM some_table WHERE id = $1;",
		},
		{
			query:         "SELECT id, name FROM some_table WHERE id = ? AND name = ?;",
			argCount:      2,
			expectedQuery: "SELECT id, name FROM some_table WHERE id = $1 AND name = $2;",
		},
	}
	for _, tc := range testCases {
		adapter := storage.PostgresqlAdapter{}
		newQuery := adapter.ConvertGenericPlaceholders(tc.query, tc.argCount)
		assert.Equal(t, newQuery, tc.expectedQuery)
	}
}

func TestPostgresql_TableExistsDefaultSchema(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.PostgresqlAdapter{}
	db, err := sql.Open("pgx", adapter.CreateDSN(cfg))
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

func TestPostgresql_TableExistsCustomSchema(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.PostgresqlAdapter{}
	db, err := sql.Open("pgx", adapter.CreateDSN(cfg))
	assert.Nil(t, err)
	t.Cleanup(func() {
		_, err = db.Exec("DROP SCHEMA IF EXISTS custom_schema CASCADE;")
		assert.Nil(t, err)
		assert.Nil(t, db.Close())
	})

	exists, err := adapter.TableExists(db, "custom_table.tmp")
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = db.Exec("CREATE TABLE custom_schema.tmp(id INT PRIMARY KEY);")
	assert.Nil(t, err)

	exists, err = adapter.TableExists(db, "custom_schema.tmp")
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestPostgresql_DatabaseName(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	adapter := storage.PostgresqlAdapter{}
	db, err := sql.Open("pgx", adapter.CreateDSN(cfg))
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, db.Close())
	})

	name, err := adapter.DatabaseName(db)
	assert.Nil(t, err)

	assert.Equal(t, name, cfg.DBName)
}

func TestPostgresql_CreateDSN(t *testing.T) {
	cfg := configloader.ConnectionConfig{
		Driver:   "postgres",
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
