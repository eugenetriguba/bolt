package storage

import "github.com/eugenetriguba/bolt/internal/configloader"

type DBAdapter interface {
	ConvertGenericPlaceholders(query string, argsCount int) string
	TableExists(executor sqlExecutor, tableName string) (bool, error)
	DatabaseName(executor sqlExecutor) (string, error)
	CreateDSN(cfg configloader.ConnectionConfig) string
}
