package bolttest

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/storage"
	"gotest.tools/v3/assert"
)

func NewTestConnectionConfig(t *testing.T, driver string) *configloader.ConnectionConfig {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	assert.NilError(t, err)

	return &configloader.ConnectionConfig{
		Driver:   driver,
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}
}

func NewTestDB(t *testing.T, driver string) *sql.DB {
	connectionConfig := NewTestConnectionConfig(t, driver)
	connInfo := storage.DBConnectionString(connectionConfig)
	db, err := storage.DBConnect(driver, connInfo)
	assert.NilError(t, err)
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
		assert.NilError(t, err)
		err = db.Close()
		assert.NilError(t, err)
	})
	return db
}
