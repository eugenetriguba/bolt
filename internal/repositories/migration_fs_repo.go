package repositories

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

type MigrationFsRepo interface {
	Create(migration *models.Migration) error
	List() (map[string]*models.Migration, error)
	Latest() (*models.Migration, error)
	ReadUpgradeScript(migration *models.Migration) (string, error)
	ReadDowngradeScript(migration *models.Migration) (string, error)
}

type migrationFsRepo struct {
	migrationsDirPath string
}

func NewMigrationFsRepo(
	migrationsConfig *configloader.MigrationsConfig,
) (MigrationFsRepo, error) {
	fileInfo, err := os.Stat(migrationsConfig.DirectoryPath)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(migrationsConfig.DirectoryPath, 0755)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to create migration directory at %s: %w",
				migrationsConfig.DirectoryPath,
				err,
			)
		}
	} else if err != nil {
		return nil, fmt.Errorf(
			"unable to check if migration directory at %s exists: %w",
			migrationsConfig.DirectoryPath,
			err,
		)
	} else if !fileInfo.IsDir() {
		return nil, &ErrIsNotDir{path: migrationsConfig.DirectoryPath}
	}

	return &migrationFsRepo{migrationsDirPath: migrationsConfig.DirectoryPath}, nil
}

func (mr migrationFsRepo) Create(migration *models.Migration) error {
	newMigrationDir := filepath.Join(mr.migrationsDirPath, mr.migrationDirname(migration))

	err := os.Mkdir(newMigrationDir, 0755)
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(newMigrationDir, "upgrade.sql"))
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(newMigrationDir, "downgrade.sql"))
	if err != nil {
		return err
	}

	return nil
}

func (mr migrationFsRepo) Latest() (*models.Migration, error) {
	// TODO: ReadDir returns migrations sorted by filename.
	// To be correct, it seems like we will need to sort these
	// migrations by the version style that is configured and then
	// grab the latest.
	entries, err := os.ReadDir(mr.migrationsDirPath)
	if err != nil {
		return nil, err
	}

	if len(entries) > 0 {
		return dirEntryToMigration(entries[len(entries)-1])
	}

	return nil, nil
}

func (mr migrationFsRepo) List() (map[string]*models.Migration, error) {
	entries, err := os.ReadDir(mr.migrationsDirPath)
	if err != nil {
		return nil, err
	}

	var migrations = make(map[string]*models.Migration, 0)
	for _, entry := range entries {
		migration, err := dirEntryToMigration(entry)
		if err != nil {
			return nil, err
		}
		migrations[migration.Version] = migration
	}

	return migrations, nil
}

func dirEntryToMigration(entry fs.DirEntry) (*models.Migration, error) {
	parts := strings.SplitN(entry.Name(), "_", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"%s is an invalid migration name: expected a "+
				"migration directory of the format <version>_<message>",
			entry.Name(),
		)
	}
	return &models.Migration{
		Version: parts[0],
		Message: parts[1],
		Applied: false,
	}, nil
}

func (mr migrationFsRepo) ReadUpgradeScript(migration *models.Migration) (string, error) {
	upgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		mr.migrationDirname(migration),
		"upgrade.sql",
	)
	return mr.readScriptContents(upgradeScriptPath)
}

func (mr migrationFsRepo) ReadDowngradeScript(
	migration *models.Migration,
) (string, error) {
	downgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		mr.migrationDirname(migration),
		"downgrade.sql",
	)
	return mr.readScriptContents(downgradeScriptPath)
}

func (mr migrationFsRepo) readScriptContents(scriptPath string) (string, error) {
	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

// migrationDirname creates the directory name that
// should be used for this migration.
func (mr migrationFsRepo) migrationDirname(migration *models.Migration) string {
	return fmt.Sprintf("%s_%s", migration.Version, mr.normalizeMessage(migration))
}

// normalizeMessage normalizes the Message of a migration
// to be filesystem friendly.
func (mr migrationFsRepo) normalizeMessage(migration *models.Migration) string {
	message := strings.ToLower(migration.Message)
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, " ", "_")
	return message
}
