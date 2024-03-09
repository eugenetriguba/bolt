package storage

import (
	"errors"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var postgresqlDriverName = "postgresql"
var mysqlDriverName = "mysql"
var supportedDrivers = []string{postgresqlDriverName, mysqlDriverName}

var (
	ErrMalformedConnectionString = errors.New(
		"malformed database connection parameters provided",
	)
	ErrUnableToConnect   = errors.New("unable to open connection to database")
	ErrUnsupportedDriver = fmt.Errorf("unsupported driver, supported drivers are %s", supportedDrivers)
)

type DB struct {
	Session db.Session
}

// DBConnect establishes a connection to the database using the
// connection configuration.
//
// The following errors may be returned:
//   - ErrUnableToConnect: Unable to make a connection to the database with
//     the provided connection parameters.
//   - ErrUnsupportedDriver: The provided driver is not supported.
func DBConnect(cfg configloader.ConnectionConfig) (DB, error) {
	var db db.Session
	var err error = nil

	switch cfg.Driver {
	case postgresqlDriverName:
		db, err = postgresql.Open(postgresql.ConnectionURL{
			User:     cfg.User,
			Password: cfg.Password,
			Host:     cfg.Host,
			Database: cfg.DBName,
		})
	case mysqlDriverName:
		db, err = mysql.Open(mysql.ConnectionURL{
			User:     cfg.User,
			Password: cfg.Password,
			Host:     cfg.Host,
			Database: cfg.DBName,
		})
	default:
		return DB{}, ErrUnsupportedDriver
	}

	if err != nil {
		return DB{}, fmt.Errorf("%w: %v", ErrUnableToConnect, err)
	}

	return DB{Session: db}, nil
}
