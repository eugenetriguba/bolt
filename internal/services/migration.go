package services

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/output"
	"github.com/eugenetriguba/bolt/internal/repositories"
)

type MigrationService struct {
	dbRepo    *repositories.MigrationDBRepo
	fsRepo    *repositories.MigrationFsRepo
	outputter output.Outputter
}

func NewMigrationService(
	dbRepo *repositories.MigrationDBRepo,
	fsRepo *repositories.MigrationFsRepo,
	outputter output.Outputter,
) *MigrationService {
	return &MigrationService{dbRepo: dbRepo, fsRepo: fsRepo, outputter: outputter}
}

type sortOrder int

const (
	SortOrderDesc sortOrder = iota
	SortOrderAsc
)

func (ms *MigrationService) ApplyAllMigrations() error {
	migrations, err := ms.ListMigrations(SortOrderAsc)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if !migration.Applied {
			err := ms.ApplyMigration(migration)
			if err != nil {
				return fmt.Errorf(
					"unable to apply migration %s: %w",
					migration.Name(),
					err,
				)
			}
		}
	}

	return nil
}

func (ms *MigrationService) ApplyUpToVersion(version string) error {
	migrations, err := ms.ListMigrations(SortOrderAsc)
	if err != nil {
		return err
	}

	var targetMigration *models.Migration
	for _, migration := range migrations {
		if migration.Version == version {
			targetMigration = migration
			break
		}
	}
	if targetMigration == nil {
		return fmt.Errorf("migration with version %s does not exist", version)
	}
	if targetMigration.Applied {
		return fmt.Errorf(
			"migration with version %s is already applied, nothing to apply",
			version,
		)
	}

	for _, migration := range migrations {
		if !migration.Applied {
			err := ms.ApplyMigration(migration)
			if err != nil {
				return fmt.Errorf(
					"unable to apply migration %s: %w",
					migration.Name(),
					err,
				)
			}
		}

		if migration.Version == version {
			break
		}
	}

	return nil
}

func (ms *MigrationService) ApplyMigration(migration *models.Migration) error {
	ms.outputter.Output(
		fmt.Sprintf(
			"Applying migration %s_%s..",
			migration.Version,
			migration.Message,
		),
	)

	scriptContents, err := ms.fsRepo.ReadUpgradeScript(migration)
	if err != nil {
		return err
	}

	err = ms.dbRepo.Apply(scriptContents, migration)
	if err != nil {
		return err
	}
	ms.outputter.Output(
		fmt.Sprintf(
			"Successfully applied migration %s_%s!",
			migration.Version,
			migration.Message,
		),
	)

	return nil
}

func (ms *MigrationService) RevertAllMigrations() error {
	migrations, err := ms.ListMigrations(SortOrderDesc)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if migration.Applied {
			err := ms.RevertMigration(migration)
			if err != nil {
				return fmt.Errorf(
					"unable to revert migration %s: %w",
					migration.Name(),
					err,
				)
			}
		}
	}

	return nil
}

func (ms *MigrationService) RevertDownToVersion(version string) error {
	migrations, err := ms.ListMigrations(SortOrderDesc)
	if err != nil {
		return err
	}

	var targetMigration *models.Migration
	for _, migration := range migrations {
		if migration.Version == version {
			targetMigration = migration
			break
		}
	}
	if targetMigration == nil {
		return fmt.Errorf("migration with version %s does not exist", version)
	}
	if !targetMigration.Applied {
		return fmt.Errorf(
			"migration with version %s isn't applied, nothing to revert",
			version,
		)
	}

	for _, migration := range migrations {
		if migration.Applied {
			err := ms.RevertMigration(migration)
			if err != nil {
				return fmt.Errorf(
					"unable to revert migration %s: %w",
					migration.Name(),
					err,
				)
			}
		}

		if migration.Version == version {
			break
		}
	}

	return nil
}

func (ms *MigrationService) RevertMigration(migration *models.Migration) error {
	ms.outputter.Output(
		fmt.Sprintf(
			"Reverting migration %s_%s..",
			migration.Version,
			migration.Message,
		),
	)
	scriptContents, err := ms.fsRepo.ReadDowngradeScript(migration)
	if err != nil {
		return err
	}

	err = ms.dbRepo.Revert(scriptContents, migration)
	if err != nil {
		return err
	}
	ms.outputter.Output(
		fmt.Sprintf(
			"Successfully reverted migration %s_%s!",
			migration.Version,
			migration.Message,
		),
	)

	return nil
}

func (ms *MigrationService) CreateMigration(
	versionStyle configloader.VersionStyle,
	message string,
) error {
	var migration *models.Migration

	if versionStyle == configloader.VersionStyleTimestamp {
		migration = models.NewTimestampMigration(time.Now(), message)
	} else {
		var currentVerison uint64 = 0
		latestMigration, err := ms.fsRepo.Latest()
		if err != nil {
			return err
		}
		if latestMigration != nil {
			currentVerison, err = strconv.ParseUint(latestMigration.Version, 10, 64)
			if err != nil {
				return err
			}
		}
		migration = models.NewSequentialMigration(currentVerison+1, message)
	}

	err := ms.fsRepo.Create(migration)
	if err != nil {
		return err
	}
	ms.outputter.Output(
		fmt.Sprintf("Created migration %s - %s.", migration.Version, migration.Message),
	)
	return nil
}

func (ms *MigrationService) ListMigrations(order sortOrder) ([]*models.Migration, error) {
	// Assumption: All migration versions that have been applied
	// to the database exist locally in the list of filesystem migrations.
	// This would hold true unless someone decided to apply a migration and,
	// for whatever reason, delete it later from their filesystem.
	migrations, err := ms.fsRepo.List()
	if err != nil {
		return nil, err
	}

	appliedMigrations, err := ms.dbRepo.List()
	if err != nil {
		return nil, err
	}

	var flatMigrations []*models.Migration
	for _, migration := range migrations {
		_, ok := appliedMigrations[migration.Version]
		if ok {
			migration.Applied = true
		}
		flatMigrations = append(flatMigrations, migration)
	}

	sort.Slice(flatMigrations, func(i, j int) bool {
		if order == SortOrderAsc {
			return flatMigrations[i].Version < flatMigrations[j].Version
		} else {
			return flatMigrations[i].Version > flatMigrations[j].Version
		}
	})

	return flatMigrations, nil
}
