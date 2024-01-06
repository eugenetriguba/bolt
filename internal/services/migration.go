package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
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

type MigrationService struct {
	repo *repositories.MigrationRepo
	cfg  *configloader.Config
}

func NewMigrationService(repo *repositories.MigrationRepo, cfg *configloader.Config) (*MigrationService, error) {
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

	return &MigrationService{repo: repo, cfg: cfg}, nil
}

func (ms *MigrationService) ApplyAllMigrations() error {
	migrations, err := ms.repo.List(repositories.SortOrderAsc)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		ms.ApplyMigration(migration)
	}

	return nil
}

func (ms *MigrationService) ApplyMigration(migration *models.Migration) error {
	upgradeScriptPath := filepath.Join(
		ms.cfg.MigrationsDir,
		fmt.Sprintf("%s_%s", migration.Version, migration.NormalizedMessage()),
		"upgrade.sql",
	)

	upgradeScript, err := os.ReadFile(upgradeScriptPath)
	if err != nil {
		return err
	}

	err = ms.repo.Apply(string(upgradeScript), migration)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MigrationService) RevertAllMigrations() error {
	migrations, err := ms.repo.List(repositories.SortOrderDesc)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		ms.RevertMigration(migration)
	}

	return nil
}

func (ms *MigrationService) RevertMigration(migration *models.Migration) error {
	downgradeScriptPath := filepath.Join(
		ms.cfg.MigrationsDir,
		fmt.Sprintf("%s_%s", migration.Version, migration.NormalizedMessage()),
		"downgrade.sql",
	)

	downgradeScript, err := os.ReadFile(downgradeScriptPath)
	if err != nil {
		return err
	}

	err = ms.repo.Apply(string(downgradeScript), migration)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MigrationService) CreateMigration(m *models.Migration) error {
	path := filepath.Join(
		ms.cfg.MigrationsDir,
		fmt.Sprintf("%s_%s", m.Version, m.NormalizedMessage()),
	)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(path, "upgrade.sql"))
	if err != nil {
		return err
	}

	_, err = os.Create(filepath.Join(path, "downgrade.sql"))
	if err != nil {
		return err
	}

	return nil
}

func (ms *MigrationService) ListMigrations() ([]*models.Migration, error) {
	entries, err := os.ReadDir(ms.cfg.MigrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations = make(map[string]*models.Migration)
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
		migrations[version] = &models.Migration{
			Version: version,
			Message: message,
			Applied: false,
		}
	}

	// if order == SortOrderDesc {
	// 	sort.Slice(values, func(i, j int) bool {
	// 		return values[i].Dirname() < values[j].Dirname()
	// 	})
	// }

	return migrations, nil
}
