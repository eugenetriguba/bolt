package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/storage"
)

type MigrationDBRepo interface {
	List() (map[string]*models.Migration, error)
	IsApplied(version string) (bool, error)
	Apply(upgradeScript string, migration *models.Migration) error
	ApplyWithTx(upgradeScript string, migration *models.Migration) error
	Revert(downgradeScript string, migration *models.Migration) error
	RevertWithTx(downgradeScript string, migration *models.Migration) error
}

type migrationDBRepo struct {
	migrationTableName string
	db                 storage.DB
}

// NewMigrationDBRepo initializes the MigrationDBRepo with a
// database. Furthermore, it ensures the migration table it
// operates on exists. If it is unable to create or confirm
// the table exists, an error is returned.
func NewMigrationDBRepo(migrationTableName string, db storage.DB) (MigrationDBRepo, error) {
	err := sanitizeTableName(migrationTableName)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid migration table name: %w",
			err,
		)
	}

	migrationTableExists, err := db.TableExists(migrationTableName)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to confirm '%s' database table exists: %w",
			migrationTableName,
			err,
		)
	}

	if !migrationTableExists {
		_, err := db.Exec(fmt.Sprintf(`
			CREATE TABLE %s (
				version VARCHAR(255) PRIMARY KEY NOT NULL
			);
		`, migrationTableName))
		if err != nil {
			return nil, fmt.Errorf(
				"unable to create '%s' database table: %w",
				migrationTableName,
				err,
			)
		}
	}

	return &migrationDBRepo{migrationTableName: migrationTableName, db: db}, nil
}

func sanitizeTableName(tableName string) error {
	// Allow alphanumeric characters, underscores, and a single dot for schema.table
	validTableName := regexp.MustCompile(`^[a-zA-Z0-9_]+(\.[a-zA-Z0-9_]+)?$`)
	if !validTableName.MatchString(tableName) {
		return errors.New(
			"a migration table name must only contain alphanumeric or underscore characters " +
				"and optionally a single dot for schema-qualified names")
	}
	return nil
}

// List retrieves a map of migration models that can be
// looked up by their `Version`. All migrations retrieved
// will be ones that have been applied, and their message
// will always be an empty string.
func (mr migrationDBRepo) List() (map[string]*models.Migration, error) {
	rows, err := mr.db.Query(fmt.Sprintf("SELECT version FROM %s;", mr.migrationTableName))
	if err != nil {
		return nil, fmt.Errorf(
			"unable to execute query to select versions from "+
				"'%s' database table: %w",
			mr.migrationTableName,
			err,
		)
	}
	defer rows.Close()

	var migrations = make(map[string]*models.Migration, 0)
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to scan version row from applied migrations: %w",
				err,
			)
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
func (mr migrationDBRepo) IsApplied(version string) (bool, error) {
	var scanResult int
	err := mr.db.QueryRow(fmt.Sprintf("SELECT 1 FROM %s WHERE version = ?", mr.migrationTableName), version).
		Scan(&scanResult)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf(
			"unable to check whether the migration %s is applied: %w",
			version,
			err,
		)
	}
	return true, nil
}

// Apply applies a migration by executing the corresponding upgrade script
// and adding the applied migration version into the migrations table. When
// successfully applied, the `migration` model's `Applied` field will be set
// to true.
func (mr migrationDBRepo) Apply(
	upgradeScript string,
	migration *models.Migration,
) error {
	err := mr.applyMigration(upgradeScript, *migration)
	if err != nil {
		return err
	}

	migration.Applied = true
	return nil
}

// ApplyWithTx applies a migration like Apply. However, it
// wraps the operation is a database transaction.
func (mr migrationDBRepo) ApplyWithTx(
	upgradeScript string,
	migration *models.Migration,
) error {
	err := mr.db.Tx(func(db storage.DB) error {
		err := mr.applyMigration(upgradeScript, *migration)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	migration.Applied = true
	return nil
}

func (mr migrationDBRepo) applyMigration(
	upgradeScript string,
	migration models.Migration,
) error {
	_, err := mr.db.Exec(upgradeScript)
	if err != nil {
		return fmt.Errorf("unable to execute upgrade script: %w", err)
	}

	_, err = mr.db.Exec(fmt.Sprintf("INSERT INTO %s(version) VALUES(?)", mr.migrationTableName), migration.Version)
	if err != nil {
		return fmt.Errorf(
			"unable to insert migration: %w",
			err,
		)
	}

	return nil
}

// Revert reverts a migration by executing the corresponding downgrade script
// and deleting the migration version from the migrations table. When successfully
// reverted, the `migration` model's `Applied` field will be set to false.
func (mr migrationDBRepo) Revert(
	downgradeScript string,
	migration *models.Migration,
) error {
	err := mr.revertMigration(downgradeScript, *migration)
	if err != nil {
		return err
	}

	migration.Applied = false
	return nil
}

// RevertWithTx reverts a migration like Revert. However, it
// wraps the operation is a database transaction.
func (mr migrationDBRepo) RevertWithTx(
	downgradeScript string,
	migration *models.Migration,
) error {
	err := mr.db.Tx(func(db storage.DB) error {
		err := mr.revertMigration(downgradeScript, *migration)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	migration.Applied = false
	return nil
}

func (mr migrationDBRepo) revertMigration(
	downgradeScript string,
	migration models.Migration,
) error {
	_, err := mr.db.Exec(downgradeScript)
	if err != nil {
		return fmt.Errorf("unable to execute downgrade script: %w", err)
	}

	_, err = mr.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE version = ?", mr.migrationTableName), migration.Version)
	if err != nil {
		return fmt.Errorf(
			"unable to remove reverted migration from %s table: %w",
			mr.migrationTableName,
			err,
		)
	}

	return nil
}
