package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var postgresqlDriverName = "postgresql"
var mysqlDriverName = "mysql"

var supportedDrivers = map[string]dbDriver{
	postgresqlDriverName: {name: "pgx", adapter: PostgresqlAdapter{}},
	mysqlDriverName:      {name: "mysql", adapter: MySQLAdapter{}},
}

type dbDriver struct {
	name    string
	adapter DBAdapter
}

var (
	ErrMalformedConnectionString = errors.New(
		"malformed database connection parameters provided",
	)
	ErrUnableToConnect   = errors.New("unable to open connection to database")
	ErrUnsupportedDriver = fmt.Errorf(
		"unsupported driver, supported drivers are %s",
		[]string{postgresqlDriverName, mysqlDriverName},
	)
)

type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type DB struct {
	executor sqlExecutor
	sqlDB    *sql.DB
	adapter  DBAdapter
}

// NewDB establishes a connection to the database using the
// connection configuration.
//
// The following errors may be returned:
//   - ErrMalformedConnectionString: The provided connection parameters are
//     not in a valid format.
//   - ErrUnableToConnect: Unable to make a connection to the database with
//     the provided connection parameters.
//   - ErrUnsupportedDriver: The provided driver is not supported.
func NewDB(cfg configloader.ConnectionConfig) (DB, error) {
	driver, exists := supportedDrivers[cfg.Driver]
	if !exists {
		return DB{}, ErrUnsupportedDriver
	}

	db, err := sql.Open(driver.name, driver.adapter.CreateDSN(cfg))
	if err != nil {
		return DB{}, fmt.Errorf("%w: %v", ErrMalformedConnectionString, err)
	}

	// Note: `sql.Open` only validates the connection string we provided is sane.
	// It doesn't open up a connection to the database. For that, we ping
	// the database to ensure the connection string is fully valid.
	err = db.Ping()
	if err != nil {
		return DB{}, fmt.Errorf("%w: %v", ErrUnableToConnect, err)
	}

	return DB{executor: db, sqlDB: db, adapter: driver.adapter}, nil
}

// Close closes the database connection. Any further
// queries will result in errors, and you should call
// NewDB again after if you'd like to run more.
func (db DB) Close() error {
	return db.sqlDB.Close()
}

// Exec is a wrapper around the sql.DB Exec.
func (db DB) Exec(query string, args ...any) (sql.Result, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Exec(newQuery, args...)
}

// Query is a wrapper around the sql.DB Query.
func (db DB) Query(query string, args ...any) (*sql.Rows, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Query(newQuery, args...)
}

// QueryRow is a wrapper around the sql.DB QueryRow.
func (db DB) QueryRow(query string, args ...any) *sql.Row {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.QueryRow(newQuery, args...)
}

// TableExists checks if the tableName exists within the
// database currently connected to.
func (db DB) TableExists(tableName string) (bool, error) {
	return db.adapter.TableExists(db.executor, tableName)
}

type txFunc func(db DB) error

// Tx executes fn within a transaction block. If
// fn returns an error, the transaction will be rolled
// back. Otherwise, it will be committed.
func (db *DB) Tx(fn txFunc) error {
	tx, err := db.sqlDB.Begin()
	if err != nil {
		return fmt.Errorf(
			"unable to start transaction: %w",
			err,
		)
	}
	defer tx.Rollback()

	// Create a shallow clone of the DB instance for
	// the transaction scope with a tx executor.
	txDB := *db
	txDB.executor = tx

	err = fn(txDB)
	if err != nil {
		return fmt.Errorf("unable to execute transaction: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to commit transaction: %w", err)
	}

	return nil
}
