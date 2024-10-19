package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/microsoft/go-mssqldb/msdsn"
)

type MSSQLAdapter struct{}

func (m MSSQLAdapter) ConvertGenericPlaceholders(
	query string,
	argsCount int,
) string {
	for i := 0; i <= argsCount; i++ {
		placeholder := m.createPlaceholder(i + 1)
		query = strings.Replace(query, "?", placeholder, 1)
	}
	return query
}

func (m MSSQLAdapter) createPlaceholder(index int) string {
	return fmt.Sprintf("@p%d", index)
}

func (m MSSQLAdapter) TableExists(
	executor sqlExecutor,
	tableName string,
) (bool, error) {
	var exists bool

	schemaName := "dbo"
	parts := strings.Split(tableName, ".")
	if len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	} else {
		tableName = parts[0]
	}

	err := executor.QueryRow(`
		SELECT CASE WHEN EXISTS (
			SELECT * 
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = @p1
			AND TABLE_NAME = @p2
		) THEN 1 ELSE 0 END
	`, schemaName, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf(
			"unable to check if %s exists: %w",
			tableName,
			err,
		)
	}

	return exists, nil
}

func (m MSSQLAdapter) DatabaseName(executor sqlExecutor) (string, error) {
	var name string
	err := executor.QueryRow("SELECT DB_NAME();").Scan(&name)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve database name: %w", err)
	}

	return name, nil
}

func (m MSSQLAdapter) CreateDSN(cfg configloader.DatabaseConfig) string {
	port, err := strconv.ParseUint(cfg.Port, 10, 64)
	if err != nil {
		// Use default port if we can't parse it out.
		// The mssql driver requires an int port to be passed.
		port = 1433
	}
	dsnCfg := msdsn.Config{
		Host:     cfg.Host,
		Port:     port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.DBName,
	}
	return dsnCfg.URL().String()
}
