package storage

import (
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
)

type SqliteAdapter struct{}

func (s SqliteAdapter) ConvertGenericPlaceholders(
	query string,
	argsCount int,
) string {
	// The sqlite driver uses ? as a generic placeholder,
	// so no translation is necessary.
	return query
}

func (s SqliteAdapter) TableExists(
	executor sqlExecutor,
	tableName string,
) (bool, error) {
	var count int
	err := executor.QueryRow(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type='table' AND name=?;
	`, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf(
			"unable to check if %s exists: %w",
			tableName,
			err,
		)
	}

	return count > 0, nil
}

func (s SqliteAdapter) DatabaseName(executor sqlExecutor) (string, error) {
	return "main", nil
}

func (s SqliteAdapter) CreateDSN(cfg configloader.DatabaseConfig) string {
	// Note: Use the dbname as the sqlite db name/path
	return cfg.DBName
}
