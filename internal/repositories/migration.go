package repositories

import (
	"database/sql"
)

type MigrationRepo struct {
	// The database connection that will be used
	// to look up migrations that have already been
	// applied.
	db *sql.DB
}

// Create a new Migration Repo.
//
// This handles the interactions with applying
// and reverting migrations.
func NewMigrationRepo(db *sql.DB) (*MigrationRepo, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bolt_migrations(
			version CHARACTER(32) PRIMARY KEY NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &MigrationRepo{db: db}, nil
}

func (m *MigrationRepo) Apply() error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sqlStatement := `INSERT INTO applied_migration(version) VALUES ($1);`
	_, err = m.db.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
