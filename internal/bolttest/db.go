package bolttest

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/eugenetriguba/checkmate/assert"
)

func NewTestConnectionConfig(t *testing.T, driver string) *configloader.ConnectionConfig {
	port, err := strconv.Atoi(os.Getenv("BOLT_DB_CONN_PORT"))
	assert.Nil(t, err)

	return &configloader.ConnectionConfig{
		Driver:   driver,
		DBName:   os.Getenv("BOLT_DB_CONN_DBNAME"),
		Host:     os.Getenv("BOLT_DB_CONN_HOST"),
		Port:     port,
		User:     os.Getenv("BOLT_DB_CONN_USER"),
		Password: os.Getenv("BOLT_DB_CONN_PASSWORD"),
	}
}

func NewTestDB(t *testing.T, driver string) *sql.DB {
	connectionConfig := NewTestConnectionConfig(t, driver)
	connInfo := storage.DBConnectionString(connectionConfig)
	db, err := storage.DBConnect(driver, connInfo)
	assert.Nil(t, err)
	t.Cleanup(func() {
		_, err = db.Exec(`
			DO $$ DECLARE rec RECORD;
			BEGIN FOR rec IN (
				SELECT table_name 
				FROM information_schema.tables 
				WHERE table_schema = 'public'
			) LOOP EXECUTE 'DROP TABLE IF EXISTS ' || rec.table_name || ' CASCADE';
			END LOOP;
			END $$;
		`)
		assert.Nil(t, err)
		err = db.Close()
		assert.Nil(t, err)
	})
	return db
}
