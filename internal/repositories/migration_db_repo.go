package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/upper/db/v4"
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
	db storage.DB
}

// NewMigrationDBRepo initializes the MigrationDBRepo with a
// database. Furthermore, it ensures the migration table it
// operates on exists. If it is unable to create or confirm
// the table exists, an error is returned.
func NewMigrationDBRepo(db storage.DB) (MigrationDBRepo, error) {
	_, err := db.Session.SQL().Exec(`
		CREATE TABLE IF NOT EXISTS bolt_migrations(
			version CHARACTER(14) PRIMARY KEY NOT NULL
		);
	`)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to confirm bolt_migrations database table exists: %w",
			err,
		)
	}

	return &migrationDBRepo{db: db}, nil
}

// List retrieves a map of migration models that can be
// looked up by their `Version`. All migrations retrieved
// will be ones that have been applied, and their message
// will always be an empty string.
func (mr migrationDBRepo) List() (map[string]*models.Migration, error) {
	rows, err := mr.db.Session.SQL().Select("version").From("bolt_migrations").Query()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to execute query to select versions from "+
				"bolt_migrations database table: %w",
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
	row, err := mr.db.Session.SQL().Select(1).From("bolt_migrations").Where("version = ?", version).QueryRow()
	if err != nil {
		return false, fmt.Errorf("unable to execute query to check if version exists in bolt_migrations: %w", err)
	}
	var scanResult int
	err = row.Scan(&scanResult)
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
	err := applyMigration(mr.db.Session, upgradeScript, *migration)
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
	err := mr.db.Session.Tx(func(sess db.Session) error {
		err := applyMigration(sess, upgradeScript, *migration)
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

func applyMigration(
	db db.Session,
	upgradeScript string,
	migration models.Migration,
) error {
	_, err := db.SQL().Exec(upgradeScript)
	if err != nil {
		return fmt.Errorf("unable to execute upgrade script: %w", err)
	}

	_, err = db.SQL().InsertInto("bolt_migrations").Columns("version").Values(migration.Version).Exec()
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
	err := revertMigration(mr.db.Session, downgradeScript, *migration)
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
	err := mr.db.Session.Tx(func(sess db.Session) error {
		err := revertMigration(sess, downgradeScript, *migration)
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

func revertMigration(
	db db.Session,
	downgradeScript string,
	migration models.Migration,
) error {
	_, err := db.SQL().Exec(downgradeScript)
	if err != nil {
		return fmt.Errorf("unable to execute downgrade script: %w", err)
	}

	_, err = db.SQL().DeleteFrom("bolt_migrations").Where("version = ?", migration.Version).Exec()
	if err != nil {
		return fmt.Errorf(
			"unable to remove reverted migration from bolt_migrations table: %w",
			err,
		)
	}

	return nil
}
