package repositories

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/eugenetriguba/bolt/internal/models"
)

type MigrationDBRepo struct {
	db *sql.DB
}

// NewMigrationDBRepo initializes the MigrationRepo with a
// database and ensures the bolt_migration table exists.
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
		migrations[version] = &models.Migration{
			Version: strings.TrimSpace(version),
			Message: "",
			Applied: true,
		}
	}

	return migrations, nil
}

// IsApplied checks if the given version has been applied.
//
// Note that applied will be false when the version isn't
// applied and when the version might be applied, but there
// was an error. Check err first before looking at whether the
// version is applied.
func (mr *MigrationDBRepo) IsApplied(version string) (applied bool, err error) {
	row := mr.db.QueryRow(`SELECT 1 FROM bolt_migrations WHERE version = $1`, version)
	err = row.Scan()
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Apply applies a migration by executing the corresponding upgrade script
// and added the applied migration version into the bolt_migrations table.
func (mr *MigrationDBRepo) Apply(upgradeScript string, migration *models.Migration) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(upgradeScript)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO bolt_migrations(version) VALUES ($1);`,
		migration.Version,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	migration.Applied = true
	return nil
}

// Revert reverts a migration by executing the corresponding downgrade script
// and deleting the migration version into the bolt_migrations table.
func (mr *MigrationDBRepo) Revert(downgradeScript string, migration *models.Migration) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(downgradeScript)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`DELETE FROM bolt_migrations WHERE version = $1;`,
		migration.Version,
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	migration.Applied = false
	return nil
}
