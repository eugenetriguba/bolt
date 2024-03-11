package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewDB_Success(t *testing.T) {
	db, err := storage.NewDB(bolttest.NewTestConnectionConfig())
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, db.Close())
	})
	_, err = db.Exec("SELECT 1;")
	assert.Nil(t, err)
}

func TestNewDB_UnsupportedDriver(t *testing.T) {
	t.Setenv("BOLT_DB_CONN_DRIVER", "abc123")
	_, err := storage.NewDB(bolttest.NewTestConnectionConfig())
	assert.ErrorIs(t, err, storage.ErrUnsupportedDriver)
}

func TestNewDB_UnableToConnect(t *testing.T) {
	t.Setenv("BOLT_DB_CONN_HOST", "")
	t.Setenv("BOLT_DB_CONN_PORT", "")
	_, err := storage.NewDB(bolttest.NewTestConnectionConfig())
	assert.ErrorIs(t, err, storage.ErrUnableToConnect)
}

func TestClose_IsClosed(t *testing.T) {
	db, err := storage.NewDB(bolttest.NewTestConnectionConfig())
	assert.Nil(t, err)

	err = db.Close()
	assert.Nil(t, err)

	_, err = db.Exec("SELECT 1;")
	assert.ErrorContains(t, err, "sql: database is closed")
}

func TestTableExists_DoesExist(t *testing.T) {
	db, err := storage.NewDB(bolttest.NewTestConnectionConfig())
	assert.Nil(t, err)
	t.Cleanup(func() {
		bolttest.DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})
	bolttest.DropTable(t, db, "tmp")
	_, err = db.Exec("CREATE TABLE tmp(id int primary key);")
	assert.Nil(t, err)

	exists, err := db.TableExists("tmp")

	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestTableExists_DoesNotExist(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	db, err := storage.NewDB(cfg)
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, db.Close())
	})

	exists, err := db.TableExists("tmp")

	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestQueryPlaceholders(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	db, err := storage.NewDB(cfg)
	assert.Nil(t, err)
	t.Cleanup(func() {
		bolttest.DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})
	bolttest.DropTable(t, db, "tmp")
	_, err = db.Exec(`CREATE TABLE tmp(id int primary key);`)
	assert.Nil(t, err)
	_, err = db.Exec(`INSERT INTO tmp(id) VALUES(1);`)
	assert.Nil(t, err)
	_, err = db.Exec(`INSERT INTO tmp(id) VALUES(2);`)
	assert.Nil(t, err)

	var queryRowId int
	err = db.QueryRow("SELECT id FROM tmp WHERE id = ?", 1).Scan(&queryRowId)
	assert.Nil(t, err)
	assert.Equal(t, queryRowId, 1)

	rows, err := db.Query("SELECT id FROM tmp WHERE id = ?", 1)
	assert.Nil(t, err)
	assert.True(t, rows.Next())
	var queryId int
	err = rows.Scan(&queryId)
	assert.Nil(t, err)
	assert.Equal(t, queryId, 1)
	assert.False(t, rows.Next())
}

func TestTx_Commit(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	db, err := storage.NewDB(cfg)
	assert.Nil(t, err)
	t.Cleanup(func() {
		bolttest.DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})
	bolttest.DropTable(t, db, "tmp")

	err = db.Tx(func(db storage.DB) error {
		_, err = db.Exec(`CREATE TABLE tmp(id int primary key);`)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	exists, err := db.TableExists("tmp")
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestTx_Rollback(t *testing.T) {
	cfg := bolttest.NewTestConnectionConfig()
	db, err := storage.NewDB(cfg)
	assert.Nil(t, err)
	t.Cleanup(func() {
		bolttest.DropTable(t, db, "tmp")
		assert.Nil(t, db.Close())
	})
	bolttest.DropTable(t, db, "tmp")
	_, err = db.Exec(`CREATE TABLE tmp(id INT PRIMARY KEY);`)
	assert.Nil(t, err)
	expectedErr := errors.New("error!")

	err = db.Tx(func(db storage.DB) error {
		_, err = db.Exec(`INSERT INTO tmp(id) VALUES(1)`)
		assert.Nil(t, err)
		return expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)

	var id int
	err = db.QueryRow("SELECT id FROM tmp WHERE id = 1;").Scan(&id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
