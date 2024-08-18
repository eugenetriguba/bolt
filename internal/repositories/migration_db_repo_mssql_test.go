//go:build mssql

package repositories_test

import (
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewMigrationDBRepo_CreatesTableInSchema_MSSQL(t *testing.T) {
	testdb := bolttest.NewTestDB(t)
	schemaName := "custom_schema"
	tableName := schemaName + ".bolt_migrations"

	t.Cleanup(func() {
		_, err := testdb.Exec("DROP TABLE IF EXISTS " + tableName)
		assert.Nil(t, err)
		_, err = testdb.Exec("DROP SCHEMA IF EXISTS " + schemaName)
		assert.Nil(t, err)
	})

	_, err := testdb.Exec("CREATE SCHEMA " + schemaName)
	assert.Nil(t, err)

	exists, err := testdb.TableExists(tableName)
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = repositories.NewMigrationDBRepo(tableName, testdb)
	assert.Nil(t, err)

	exists, err = testdb.TableExists(tableName)
	assert.Nil(t, err)
	assert.True(t, exists)
}
