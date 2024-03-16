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
	"github.com/eugenetriguba/bolt/internal/sqlparse"
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
	ReadUpgradeScript(migration *models.Migration) (sqlparse.MigrationScript, error)
	ReadDowngradeScript(migration *models.Migration) (sqlparse.MigrationScript, error)
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
	newMigrationPath := filepath.Join(mr.migrationsDirPath, migration.Name()+".sql")
	file, err := os.Create(newMigrationPath)
	if err != nil {
		return fmt.Errorf("unable to create file at %s: %w", newMigrationPath, err)
	}
	_, err = file.WriteString("-- migrate:up\n\n-- migrate:down\n")
	if err != nil {
		return fmt.Errorf("unable to write template to new migration script: %w", err)
	}
	return nil
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
	version := parts[0]
	message := strings.TrimSuffix(parts[1], filepath.Ext(parts[1])) // Remove .sql
	return &models.Migration{
		Version: version,
		Message: message,
		Applied: false,
	}, nil
}

func (mr migrationFsRepo) ReadUpgradeScript(migration *models.Migration) (sqlparse.MigrationScript, error) {
	scriptPath := filepath.Join(
		mr.migrationsDirPath,
		migration.Name()+".sql",
	)
	upgradeScript, _, err := mr.getMigrationScripts(scriptPath)
	return upgradeScript, err
}

func (mr migrationFsRepo) ReadDowngradeScript(
	migration *models.Migration,
) (sqlparse.MigrationScript, error) {
	scriptPath := filepath.Join(
		mr.migrationsDirPath,
		migration.Name()+".sql",
	)
	_, downgradeScript, err := mr.getMigrationScripts(scriptPath)
	return downgradeScript, err
}

func (mr migrationFsRepo) getMigrationScripts(scriptPath string) (sqlparse.MigrationScript, sqlparse.MigrationScript, error) {
	scriptContents, err := mr.readScriptContents(scriptPath)
	if err != nil {
		return sqlparse.MigrationScript{}, sqlparse.MigrationScript{}, fmt.Errorf("unable to read %s script: %w", scriptPath, err)
	}
	sqlParser := sqlparse.NewSqlParser()
	return sqlParser.Parse(strings.NewReader(scriptContents))
}

func (mr migrationFsRepo) readScriptContents(scriptPath string) (string, error) {
	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
