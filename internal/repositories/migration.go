package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/config"
)

type IsNotDirError struct {
	path string
}

func (e *IsNotDirError) Error() string {
	return fmt.Sprintf(
		"The specified migrations directory path '%s' is not a directory.",
		e.path,
	)
}

type MigrationRepo struct {
	// The database connection that will be used
	// to look up migrations that have already been
	// applied.
	db     *sql.DB
	config *config.Config
}

// Create a new Migration Repo.
//
// This handles the interactions with applying
// and reverting migrations.
func NewMigrationRepo(db *sql.DB, c *config.Config) (*MigrationRepo, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bolt_migrations(
			version CHARACTER(32) PRIMARY KEY NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(c.MigrationsDir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(c.MigrationsDir, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if err == nil && !fileInfo.IsDir() {
		return nil, &IsNotDirError{path: c.MigrationsDir}
	}

	return &MigrationRepo{db: db, config: c}, nil
}

func (mr *MigrationRepo) Create(message string) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sqlStatement := `INSERT INTO applied_migration(version) VALUES ($1);`
	_, err = mr.db.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
