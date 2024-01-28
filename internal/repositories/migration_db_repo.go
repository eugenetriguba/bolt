package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/sqlparse"
)

type MigrationDBRepo struct {
	db *sql.DB
}

// NewMigrationDBRepo initializes the MigrationDBRepo with a
// database. Furthermore, it ensures the migration table it
// operates on exists. If it is unable to create or confirm
// the table exists, an error is returned.
func NewMigrationDBRepo(db *sql.DB) (*MigrationDBRepo, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bolt_migrations(
			version CHARACTER(14) PRIMARY KEY NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &MigrationDBRepo{db: db}, nil
}

// List retrieves a map of migration models that can be
// looked up by their `Version`. All migrations retrieved
// will be ones that have been applied, and their message
// will always be an empty string.
func (mr *MigrationDBRepo) List() (map[string]*models.Migration, error) {
	rows, err := mr.db.Query(`SELECT version FROM bolt_migrations;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations = make(map[string]*models.Migration)
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}
		trimmedVersion := strings.TrimSpace(version)
		migrations[trimmedVersion] = &models.Migration{
			Version: trimmedVersion,
			// Note: We don't store the user-friendly message for
			// the migration in the database. It's purely for the
			// user to understand what the migration was locally.
			Message: "",
			Applied: true,
		}
	}

	return migrations, nil
}

// IsApplied checks if the given version has been applied.
//
// applied will be false when the version isn't applied and
// when the version might be applied, but there was an error.
// Check err first before looking at whether the version is applied.
func (mr *MigrationDBRepo) IsApplied(version string) (applied bool, err error) {
	row := mr.db.QueryRow(`SELECT 1 FROM bolt_migrations WHERE version = $1`, version)
	var scanResult int
	err = row.Scan(&scanResult)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Apply applies a migration by executing the corresponding upgrade script
// and adding the applied migration version into the migrations table. When
// successfully applied, the `migration` model's `Applied` field will be set
// to true.
func (mr *MigrationDBRepo) Apply(
	upgradeScript string,
	migration *models.Migration,
) error {
	sqlParser := sqlparse.NewSqlParser(strings.NewReader(upgradeScript))
	execOptions, err := sqlParser.Parse()
	if err != nil {
		return fmt.Errorf("unable to parse sql file for upgrade script: %w", err)
	}

	if execOptions.UseTransaction {
		tx, err := mr.db.Begin()
		if err != nil {
			return fmt.Errorf(
				"unable to start transaction during upgrade script execution: %w",
				err,
			)
		}
		defer tx.Rollback()

		_, err = tx.Exec(upgradeScript)
		if err != nil {
			return fmt.Errorf("unable to execute upgrade script: %w", err)
		}

		_, err = tx.Exec(
			`INSERT INTO bolt_migrations(version) VALUES ($1);`,
			migration.Version,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to insert applied migration to bolt_migrations table: %w",
				err,
			)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf(
				"unable to commit transaction during upgrade script execution: %w",
				err,
			)
		}
	} else {
		_, err = mr.db.Exec(upgradeScript)
		if err != nil {
			return fmt.Errorf("unable to execute upgrade script: %w", err)
		}

		_, err = mr.db.Exec(
			`INSERT INTO bolt_migrations(version) VALUES ($1);`,
			migration.Version,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to insert applied migration to bolt_migrations table: %w",
				err,
			)
		}
	}

	migration.Applied = true
	return nil
}

// Revert reverts a migration by executing the corresponding downgrade script
// and deleting the migration version from the migrations table. When successfully
// reverted, the `migration` model's `Applied` field will be set to false.
func (mr *MigrationDBRepo) Revert(
	downgradeScript string,
	migration *models.Migration,
) error {
	sqlParser := sqlparse.NewSqlParser(strings.NewReader(downgradeScript))
	execOptions, err := sqlParser.Parse()
	if err != nil {
		return fmt.Errorf("unable to parse sql file for downgrade script: %w", err)
	}

	if execOptions.UseTransaction {
		tx, err := mr.db.Begin()
		if err != nil {
			return fmt.Errorf(
				"unable to start transaction during downgrade script execution: %w",
				err,
			)
		}
		defer tx.Rollback()

		_, err = tx.Exec(downgradeScript)
		if err != nil {
			return fmt.Errorf("unable to execute downgrade script: %w", err)
		}

		_, err = tx.Exec(
			`DELETE FROM bolt_migrations WHERE version = $1;`,
			migration.Version,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to remove reverted migration from bolt_migrations table: %w",
				err,
			)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf(
				"unable to commit transaction during downgrade script execution: %w",
				err,
			)
		}
	} else {
		_, err = mr.db.Exec(downgradeScript)
		if err != nil {
			return fmt.Errorf("unable to execute downgrade script: %w", err)
		}

		_, err = mr.db.Exec(
			`DELETE FROM bolt_migrations WHERE version = $1;`,
			migration.Version,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to remove reverted migration from bolt_migrations table: %w",
				err,
			)
		}
	}

	migration.Applied = false
	return nil
}
