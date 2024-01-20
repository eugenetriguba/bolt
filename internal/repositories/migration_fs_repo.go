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

type MigrationFsRepo struct {
	migrationsDirPath string
}

func NewMigrationFsRepo(migrationsConfig *configloader.MigrationsConfig) (*MigrationFsRepo, error) {
	fileInfo, err := os.Stat(migrationsConfig.DirectoryPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(migrationsConfig.DirectoryPath, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if err == nil && !fileInfo.IsDir() {
		return nil, &ErrIsNotDir{path: migrationsConfig.DirectoryPath}
	}

	return &MigrationFsRepo{migrationsDirPath: migrationsConfig.DirectoryPath}, nil
}

func (mr *MigrationFsRepo) Create(migration *models.Migration) error {
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

func (mr *MigrationFsRepo) Exists(version string) (bool, error) {
	migrations, err := mr.List()
	if err != nil {
		return false, err
	}

	_, ok := migrations[version]
	return ok, nil
}

func (mr *MigrationFsRepo) Get(version string) (*models.Migration, error) {
	migrations, err := mr.List()
	if err != nil {
		return nil, err
	}

	migration, ok := migrations[version]
	if !ok {
		return nil, fmt.Errorf("migration %s does not exist", version)
	}

	return migration, nil
}

func (mr *MigrationFsRepo) Latest() (*models.Migration, error) {
	entries, err := os.ReadDir(mr.migrationsDirPath)
	if err != nil {
		return nil, err
	}

	if len(entries) > 0 {
		return dirEntryToMigration(entries[len(entries)-1])
	}

	return nil, nil
}

func (mr *MigrationFsRepo) List() (map[string]*models.Migration, error) {
	entries, err := os.ReadDir(mr.migrationsDirPath)
	if err != nil {
		return nil, err
	}

	var migrations = make(map[string]*models.Migration)
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

func (mr *MigrationFsRepo) ReadUpgradeScript(
	migration *models.Migration,
) (string, error) {
	upgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		mr.migrationDirname(migration),
		"upgrade.sql",
	)
	return mr.readScriptContents(upgradeScriptPath)
}

func (mr *MigrationFsRepo) ReadDowngradeScript(
	migration *models.Migration,
) (string, error) {
	downgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		mr.migrationDirname(migration),
		"downgrade.sql",
	)
	return mr.readScriptContents(downgradeScriptPath)
}

func (mr *MigrationFsRepo) readScriptContents(scriptPath string) (string, error) {
	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

// migrationDirname creates the directory name that
// should be used for this migration.
func (mr *MigrationFsRepo) migrationDirname(migration *models.Migration) string {
	return fmt.Sprintf("%s_%s", migration.Version, mr.normalizeMessage(migration))
}

// normalizeMessage normalizes the Message of a migration
// to be filesystem friendly.
func (mr *MigrationFsRepo) normalizeMessage(migration *models.Migration) string {
	message := strings.ToLower(migration.Message)
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, " ", "_")
	return message
}
