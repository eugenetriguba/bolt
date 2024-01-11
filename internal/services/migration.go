package services

import (
	"sort"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
)

type MigrationService struct {
	dbRepo *repositories.MigrationDBRepo
	fsRepo *repositories.MigrationFsRepo
}

func NewMigrationService(
	dbRepo *repositories.MigrationDBRepo,
	fsRepo *repositories.MigrationFsRepo,
) *MigrationService {
	return &MigrationService{dbRepo: dbRepo, fsRepo: fsRepo}
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
			ms.ApplyMigration(migration)
		}
	}

	return nil
}

func (ms *MigrationService) ApplyMigration(migration *models.Migration) error {
	scriptContents, err := ms.fsRepo.ReadUpgradeScript(migration)
	if err != nil {
		return err
	}

	err = ms.dbRepo.Apply(scriptContents, migration)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MigrationService) RevertAllMigrations() error {
	migrations, err := ms.ListMigrations(SortOrderDesc)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if migration.Applied {
			ms.RevertMigration(migration)
		}
	}

	return nil
}

func (ms *MigrationService) RevertMigration(migration *models.Migration) error {
	scriptContents, err := ms.fsRepo.ReadDowngradeScript(migration)
	if err != nil {
		return err
	}

	err = ms.dbRepo.Revert(scriptContents, migration)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MigrationService) CreateMigration(migration *models.Migration) error {
	return ms.fsRepo.Create(migration)
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
