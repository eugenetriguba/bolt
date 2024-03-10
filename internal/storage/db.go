package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var postgresqlDriverName = "postgresql"
var mysqlDriverName = "mysql"

var supportedDrivers = map[string]string{
	postgresqlDriverName: "pgx",
	mysqlDriverName:      "mysql",
}

var driverAdapters = map[string]DBAdapter{
	postgresqlDriverName: PostgresqlAdapter{},
	mysqlDriverName:      MySQLAdapter{},
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
	driverName, exists := supportedDrivers[cfg.Driver]
	if !exists {
		return DB{}, ErrUnsupportedDriver
	}

	adapter, exists := driverAdapters[cfg.Driver]
	if !exists {
		return DB{}, ErrUnsupportedDriver
	}

	db, err := sql.Open(driverName, adapter.CreateDSN(cfg))
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

	return DB{executor: db, sqlDB: db, adapter: adapter}, nil
}

// Close closes the database connection. Any further
// queries will result in errors, and you should call
// NewDB again after if you'd like to run more.
func (db DB) Close() error {
	return db.sqlDB.Close()
}

func (db DB) Exec(query string, args ...any) (sql.Result, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Exec(newQuery, args...)
}

func (db DB) Query(query string, args ...any) (*sql.Rows, error) {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.Query(newQuery, args...)
}

func (db DB) QueryRow(query string, args ...any) *sql.Row {
	newQuery := db.adapter.ConvertGenericPlaceholders(query, len(args))
	return db.executor.QueryRow(newQuery, args...)
}

func (db DB) TableExists(tableName string) (bool, error) {
	return db.adapter.TableExists(db.executor, tableName)
}

type txFunc func(db DB) error

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
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
