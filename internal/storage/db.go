package storage

import (
	"database/sql"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	_ "github.com/lib/pq"
)

// DBConnect establishes a connection to the database using the driver
// and connection information.
//
// Note that only "postgres" is supported as the driver right now.
func DBConnect(driver string, connectionParams string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionParams)
	if err != nil {
		return nil, fmt.Errorf("invalid database connection parameters: %w", err)
	}

	// Note: `sql.Open` only validates the connection string we provided is sane.
	// It doesn't actually open up a connection to the database. For that, we ping
	// the database to ensure the connection string is fully valid.
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return db, nil
}

// DBConnectionString generates a properly formatted connection
// string that can be used to establish a database connection.
// Note that it always creates a connection string with sslmode disabled.
func DBConnectionString(cfg *configloader.ConnectionConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
}
