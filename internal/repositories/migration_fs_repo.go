package repositories

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func NewMigrationFsRepo(migrationsDirPath string) (*MigrationFsRepo, error) {
	fileInfo, err := os.Stat(migrationsDirPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(migrationsDirPath, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if err == nil && !fileInfo.IsDir() {
		return nil, &ErrIsNotDir{path: migrationsDirPath}
	}

	return &MigrationFsRepo{migrationsDirPath: migrationsDirPath}, nil
}

func (mr *MigrationFsRepo) Create(m *models.Migration) error {
	path := filepath.Join(
		mr.migrationsDirPath,
		fmt.Sprintf("%s_%s", m.Version, m.NormalizedMessage()),
	)

	_, err := os.Create(filepath.Join(path, "upgrade.sql"))
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(path, "downgrade.sql"))
	if err != nil {
		return err
	}

	return nil
}

func (mr *MigrationFsRepo) List() (map[string]*models.Migration, error) {
	entries, err := os.ReadDir(mr.migrationsDirPath)
	if err != nil {
		return nil, err
	}

	var migrations = make(map[string]*models.Migration)
	for _, entry := range entries {
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"%s is an invalid migration name: expected a "+
					"migration directory of the format <version>_<message>",
				entry.Name(),
			)
		}
		migrations[parts[0]] = &models.Migration{
			Version: parts[0],
			Message: parts[1],
			Applied: false,
		}
	}

	return migrations, nil
}

func (mr *MigrationFsRepo) ReadUpgradeScript(migration *models.Migration) (string, error) {
	upgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		fmt.Sprintf("%s_%s", migration.Version, migration.NormalizedMessage()),
		"upgrade.sql",
	)
	return mr.readScriptContents(upgradeScriptPath)
}

func (mr *MigrationFsRepo) ReadDowngradeScript(migration *models.Migration) (string, error) {
	downgradeScriptPath := filepath.Join(
		mr.migrationsDirPath,
		fmt.Sprintf("%s_%s", migration.Version, migration.NormalizedMessage()),
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
