package storage

import (
	"fmt"
	"net/url"
)

type DBAdapter interface {
	// ConvertGenericPlaceholders replaces any generic `?` placeholders
	// in `query` for the database driver specific placeholders and returns
	// the updated query.
	ConvertGenericPlaceholders(query string, argsCount int) string
	// TableExists checks if the tableName exists within the
	// database currently connected to.
	TableExists(executor sqlExecutor, tableName string) (bool, error)
	// DatabaseName retrieves the currently selected database name.
	DatabaseName(executor sqlExecutor) (string, error)
}

type DBDriver struct {
	name    string
	adapter DBAdapter
}

var PostgresDriverName = "postgres"
var PostgresqlDriverName = "postgresql"
var MysqlDriverName = "mysql"
var MssqlDriverName = "mssql"
var SqliteDriverName = "sqlite3"

var SupportedDrivers = map[string]DBDriver{
	PostgresDriverName:   {name: "pgx", adapter: PostgresqlAdapter{}},
	PostgresqlDriverName: {name: "pgx", adapter: PostgresqlAdapter{}},
	MysqlDriverName:      {name: "mysql", adapter: MySQLAdapter{}},
	MssqlDriverName:      {name: "sqlserver", adapter: MSSQLAdapter{}},
	SqliteDriverName:     {name: "sqlite3", adapter: SqliteAdapter{}},
}

var ErrUnsupportedDriver = fmt.Errorf(
	"unsupported driver, supported drivers are %s",
	SupportedDrivers,
	[]string{
		PostgresDriverName,
		PostgresqlDriverName,
		MysqlDriverName,
		MssqlDriverName,
		SqliteDriverName,
	},
)

// NewDBDriverFromDSN attempts to infer the driver name from a given DSN.
func NewDBDriverFromDSN(dsn string) (DBDriver, error) {
	if u, err := url.Parse(dsn); err == nil {
		driver, exists := SupportedDrivers[u.Scheme]
		if exists {
			return driver, nil
		}
	}
	return DBDriver{}, ErrUnsupportedDriver
}
