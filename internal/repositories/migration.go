package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/models"
)

type ErrIsNotDir struct {
	path string
}

func (e *ErrIsNotDir) Error() string {
	return fmt.Sprintf(
		"The specified migrations directory path '%s' is not a directory.",
		e.path,
	)
}

type MigrationRepo struct {
	// The database connection that will be used
	// to look up migrations that have already been
	// applied.
	db  *sql.DB
	cfg *configloader.Config
}

// Create a new Migration Repo.
//
// This handles the interactions with applying
// and reverting migrations.
func NewMigrationRepo(db *sql.DB, cfg *configloader.Config) (*MigrationRepo, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bolt_migrations(
			version CHARACTER(14) PRIMARY KEY NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(cfg.MigrationsDir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(cfg.MigrationsDir, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if err == nil && !fileInfo.IsDir() {
		return nil, &ErrIsNotDir{path: cfg.MigrationsDir}
	}

	return &MigrationRepo{db: db, cfg: cfg}, nil
}

func (mr *MigrationRepo) Create(m *models.Migration) error {
	path := filepath.Join(mr.cfg.MigrationsDir, m.Dirname())
	err := os.Mkdir(path, 0755)
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(path, "upgrade.sql"))
	if err != nil {
		return err
	}

	return nil
}

func (mr *MigrationRepo) List() ([]*models.Migration, error) {
	rows, err := mr.db.Query(`SELECT version FROM bolt_migrations;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[string]*models.Migration)
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}
		version = strings.TrimSpace(version)
		migrations[version] = &models.Migration{
			Version: version,
			Message: "",
			Applied: true,
		}
	}

	entries, err := os.ReadDir(mr.cfg.MigrationsDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			return nil, errors.New(
				fmt.Sprintf(
					"%s is an invalid migration name. Expected a "+
						"migration directory of the format <version>_<message>.",
					entry.Name(),
				),
			)
		}
		version := parts[0]
		message := parts[1]
		val, ok := migrations[version]
		if ok {
			val.Message = message
		} else {
			migrations[version] = &models.Migration{
				Version: version,
				Message: message,
				Applied: false,
			}
		}
	}

	values := make([]*models.Migration, 0, len(migrations))
	for _, value := range migrations {
		values = append(values, value)
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Dirname() < values[j].Dirname()
	})

	return values, nil
}

func (mr *MigrationRepo) Apply(migration *models.Migration) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	upgradeScriptPath := filepath.Join(
		mr.cfg.MigrationsDir, migration.Dirname(), "upgrade.sql",
	)
	contents, err := os.ReadFile(upgradeScriptPath)
	if err != nil {
		return err
	}

	_, err = tx.Exec(string(contents))
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
	migration.Applied = true
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
