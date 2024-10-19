package storage

import (
	"github.com/eugenetriguba/bolt/internal/configloader"
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
	// CreateDSN creates a DSN to be used with sql.Open in the database
	// driver specific format.
	CreateDSN(cfg configloader.DatabaseConfig) string
}
