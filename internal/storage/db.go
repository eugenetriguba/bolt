package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

var (
	ErrMalformedConnectionString = errors.New(
		"malformed database connection parameters provided",
	)
	ErrUnableToConnect = errors.New("unable to open connection to database")
)

type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type TxFunc func(db DB) error

type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Tx(fn TxFunc) error
	Close() error
	TableExists(tableName string) (bool, error)
}

type SqlDB struct {
	executor sqlExecutor
	conn     *sql.DB
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
func NewDB(cfg configloader.DatabaseConfig) (DB, error) {
	driver, err := NewDBDriverFromDSN(cfg.DSN)
	if err != nil {
		return SqlDB{}, fmt.Errorf("unable to determine driver from DSN %s: %w", cfg.DSN, err)
	}

	db, err := sql.Open(driver.name, cfg.DSN)
	if err != nil {
		return SqlDB{}, fmt.Errorf("%w: %v", ErrMalformedConnectionString, err)
	}

	// Note: `sql.Open` only validates the connection string we provided is sane.
	// It doesn't open up a connection to the database. For that, we ping
	// the database to ensure the connection string is fully valid.
	err = db.Ping()
	if err != nil {
		return SqlDB{}, fmt.Errorf("%w: %v", ErrUnableToConnect, err)
	}

	return SqlDB{executor: db, conn: db, adapter: driver.adapter}, nil
}

// Close closes the database connection. Any further
// queries will result in errors, and you should call
// NewDB again after if you'd like to run more.
func (db SqlDB) Close() error {
	return db.conn.Close()
}

// Exec is a wrapper around the sql.DB Exec.
func (db SqlDB) Exec(query string, args ...any) (sql.Result, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Exec(newQuery, args...)
}

// Query is a wrapper around the sql.DB Query.
func (db SqlDB) Query(query string, args ...any) (*sql.Rows, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Query(newQuery, args...)
}

// QueryRow is a wrapper around the sql.DB QueryRow.
func (db SqlDB) QueryRow(query string, args ...any) *sql.Row {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.QueryRow(newQuery, args...)
}

// TableExists checks if the tableName exists within the
// database currently connected to.
func (db SqlDB) TableExists(tableName string) (bool, error) {
	return db.adapter.TableExists(db.executor, tableName)
}

// Tx executes fn within a transaction block. If
// fn returns an error, the transaction will be rolled
// back. Otherwise, it will be committed.
func (db SqlDB) Tx(fn TxFunc) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf(
			"unable to start transaction: %w",
			err,
		)
	}
	defer tx.Rollback()

	// Create a shallow clone of the DB instance for
	// the transaction scope with a tx executor.
	txDB := db
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
