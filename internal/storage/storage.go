package storage

import (
	"database/sql"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	_ "github.com/lib/pq"
)

// DBConnect to the database using the driver and connection
// information.
func DBConnect(driver string, connectionInfo string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionInfo)
	if err != nil {
		return nil, err
	}

	// Note: `sql.Open` only validates the connection string we provided
	// it is sane. It doesn't actually open up a connection to the database.
	// For that, we ping the database to ensure the connection string is fully
	// valid.
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// DBConnectionString generates a properly formatted connection
// string to be passed to a database management system to initiate
// a connection. It uses the passed in configuration values to create
// the string. Note that sslmode is always disabled.
func DBConnectionString(cfg *configloader.ConnectionConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
}
