package storage

import (
	"fmt"
	"strings"

	"github.com/eugenetriguba/bolt/internal/configloader"
)

type PostgresqlAdapter struct{}

func (p PostgresqlAdapter) ConvertGenericPlaceholders(
	query string,
	argsCount int,
) string {
	for i := 0; i <= argsCount; i++ {
		placeholder := p.createPlaceholder(i + 1)
		query = strings.Replace(query, "?", placeholder, 1)
	}
	return query
}

func (p PostgresqlAdapter) createPlaceholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (p PostgresqlAdapter) TableExists(
	executor sqlExecutor,
	tableName string,
) (bool, error) {
	var exists bool

	schemaName := "public"
	parts := strings.Split(tableName, ".")
	if len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	} else {
		tableName = parts[0]
	}

	err := executor.QueryRow(`
		SELECT EXISTS (
			SELECT FROM pg_catalog.pg_class c
			JOIN   pg_catalog.pg_namespace n ON n.oid = c.relnamespace
			WHERE  n.nspname = $1
			AND    c.relname = $2
			AND    c.relkind = 'r'  -- Only tables
		);
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

func (p PostgresqlAdapter) DatabaseName(executor sqlExecutor) (string, error) {
	var name string
	err := executor.QueryRow("SELECT current_database();").Scan(&name)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve database name: %w", err)
	}

	return name, nil
}

func (p PostgresqlAdapter) CreateDSN(cfg configloader.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
}
