package db

import (
	"database/sql"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/config"
	_ "github.com/lib/pq"
)

var db *sql.DB

// Connect to the database using the driver and connection
// parameter that are specified in the bolt configuration file.
//
// Note that subsequent calls to Connect return the same database
// handle.
func Connect() (*sql.DB, error) {
	if db == nil {
		config, err := config.NewConfig()
		if err != nil {
			return nil, err
		}

		connInfo := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Connection.Host, config.Connection.Port, config.Connection.User,
			config.Connection.Password, config.Connection.DBName,
		)
		newDb, err := sql.Open(config.Connection.Driver, connInfo)
		if err != nil {
			return nil, err
		}

		// Note: `sql.Open` only validates the connection string we provided
		// it is sane. It doesn't actually open up a connection to the database.
		// For that, we ping the database to ensure the connection string is fully
		// valid.
		err = newDb.Ping()
		if err != nil {
			return nil, err
		}

		db = newDb
	}

	return db, nil
}
