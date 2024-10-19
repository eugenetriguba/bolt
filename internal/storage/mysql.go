package storage

import (
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/go-sql-driver/mysql"
)

type MySQLAdapter struct{}

func (m MySQLAdapter) ConvertGenericPlaceholders(query string, argsCount int) string {
	// MySQL uses ? as its query argument placeholder, which we're using
	// as our generic placeholder so no transformation is necessary.
	return query
}

func (m MySQLAdapter) TableExists(
	executor sqlExecutor,
	tableName string,
) (bool, error) {
	// Note: MySQL doesn't have schemas. So the "table_schema"
	// should be the currently selected database.
	databaseName, err := m.DatabaseName(executor)
	if err != nil {
		return false, fmt.Errorf("unable to retrieve database name: %w", err)
	}

	var exists bool
	err = executor.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
		);
	`, databaseName, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf(
			"unable to check if %s exists: %w",
			tableName,
			err,
		)
	}

	return exists, nil
}

func (m MySQLAdapter) DatabaseName(executor sqlExecutor) (string, error) {
	var name string
	err := executor.QueryRow("SELECT DATABASE();").Scan(&name)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve database name: %w", err)
	}

	// MySQL allows connecting without a database selected.
	if name == "" {
		return "", fmt.Errorf("no database is currently selected")
	}

	return name, nil
}

func (m MySQLAdapter) CreateDSN(cfg configloader.DatabaseConfig) string {
	addr := cfg.Host
	if cfg.Port != "" {
		addr += fmt.Sprintf(":%s", cfg.Port)
	}
	dsnCfg := mysql.Config{
		Net:    "tcp",
		Addr:   addr,
		User:   cfg.User,
		Passwd: cfg.Password,
		DBName: cfg.DBName,
	}
	return dsnCfg.FormatDSN()
}
