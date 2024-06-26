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
	dbRepo    repositories.MigrationDBRepo
	fsRepo    repositories.MigrationFsRepo
	cfg       configloader.Config
	outputter output.Outputter
}

func NewMigrationService(
	dbRepo repositories.MigrationDBRepo,
	fsRepo repositories.MigrationFsRepo,
	cfg configloader.Config,
	outputter output.Outputter,
) MigrationService {
	return MigrationService{
		dbRepo:    dbRepo,
		fsRepo:    fsRepo,
		cfg:       cfg,
		outputter: outputter,
	}
}

type sortOrder int

const (
	SortOrderDesc sortOrder = iota
	SortOrderAsc
)

func (ms MigrationService) ApplyAllMigrations() error {
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

func (ms MigrationService) ApplyUpToVersion(version string) error {
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
		// Assumption: If the target migration is applied, all migrations before
		// it must also be applied.
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

func (ms MigrationService) ApplyMigration(migration *models.Migration) error {
	ms.outputter.Output(fmt.Sprintf("Applying migration %s..", migration.Name()))
	startTime := time.Now()

	upgradeScript, err := ms.fsRepo.ReadUpgradeScript(migration)
	if err != nil {
		return fmt.Errorf("unable to read upgrade script: %w", err)
	}

	if upgradeScript.Options.UseTransaction {
		err = ms.dbRepo.ApplyWithTx(upgradeScript.Contents, migration)
	} else {
		err = ms.dbRepo.Apply(upgradeScript.Contents, migration)
	}

	if err != nil {
		return fmt.Errorf("unable to apply migration %s: %w", migration.Name(), err)
	}

	ms.outputter.Output(
		fmt.Sprintf(
			"Successfully applied migration %s in %s!",
			migration.Name(),
			time.Since(startTime),
		),
	)

	return nil
}

func (ms MigrationService) RevertAllMigrations() error {
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

func (ms MigrationService) RevertDownToVersion(version string) error {
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
		// Assumption: Every migration from the latest down to the target
		// migration hasn't been applied.
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

func (ms MigrationService) RevertMigration(migration *models.Migration) error {
	ms.outputter.Output(fmt.Sprintf("Reverting migration %s..", migration.Name()))
	startTime := time.Now()

	downgradeScript, err := ms.fsRepo.ReadDowngradeScript(migration)
	if err != nil {
		return fmt.Errorf("unable to read downgrade script: %w", err)
	}

	if downgradeScript.Options.UseTransaction {
		err = ms.dbRepo.RevertWithTx(downgradeScript.Contents, migration)
	} else {
		err = ms.dbRepo.Revert(downgradeScript.Contents, migration)
	}

	if err != nil {
		return err
	}

	ms.outputter.Output(
		fmt.Sprintf(
			"Successfully reverted migration %s in %s!",
			migration.Name(),
			time.Since(startTime),
		),
	)

	return nil
}

func (ms MigrationService) CreateMigration(message string) (*models.Migration, error) {
	var migration *models.Migration

	if ms.cfg.Migrations.VersionStyle == configloader.VersionStyleTimestamp {
		migration = models.NewTimestampMigration(time.Now(), message)
	} else {
		currentVersion, err := ms.getCurrentSequentialMigrationVersion()
		if err != nil {
			return nil, err
		}
		migration = models.NewSequentialMigration(currentVersion+1, message)
	}

	err := ms.fsRepo.Create(migration)
	if err != nil {
		return nil, err
	}
	ms.outputter.Output(
		fmt.Sprintf("Created migration %s - %s.", migration.Version, migration.Message),
	)
	return migration, nil
}

func (ms MigrationService) getCurrentSequentialMigrationVersion() (uint64, error) {
	var currentVersion uint64 = 0

	localMigrations, err := ms.fsRepo.List()
	if err != nil {
		return 0, fmt.Errorf("unable to list out local filesystem migrations: %w", err)
	}

	migrations, err := ms.combineMigrations(
		localMigrations,
		map[string]*models.Migration{},
		SortOrderDesc,
	)
	if err != nil {
		return 0, err
	}

	if len(migrations) > 0 {
		latestMigration := migrations[0]
		currentVersion, err = strconv.ParseUint(latestMigration.Version, 10, 64)
		if err != nil {
			return 0, err
		}
	}

	return currentVersion, nil
}

func (ms MigrationService) ListMigrations(order sortOrder) ([]*models.Migration, error) {
	// Assumption: All migration versions that have been applied
	// to the database exist locally in the list of filesystem migrations.
	// This would hold true unless someone decided to apply a migration and,
	// for whatever reason, delete it later from their filesystem. If there is
	// an applied migration that doesn't exist locally, it would not be shown
	// here.
	localMigrations, err := ms.fsRepo.List()
	if err != nil {
		return nil, fmt.Errorf("unable to list out local filesystem migrations: %w", err)
	}

	appliedMigrations, err := ms.dbRepo.List()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to list out applied migrations from remote db: %w",
			err,
		)
	}

	return ms.combineMigrations(localMigrations, appliedMigrations, order)
}

func (ms MigrationService) combineMigrations(
	localMigrations map[string]*models.Migration,
	appliedMigrations map[string]*models.Migration,
	order sortOrder,
) ([]*models.Migration, error) {
	migrations := make([]*models.Migration, 0)
	for _, localMigration := range localMigrations {
		_, ok := appliedMigrations[localMigration.Version]
		if ok {
			localMigration.Applied = true
		}
		migrations = append(migrations, localMigration)
	}

	err := ms.sortMigrations(migrations, order)
	if err != nil {
		return nil, err
	}

	return migrations, err
}

func (ms MigrationService) sortMigrations(
	migrations []*models.Migration,
	order sortOrder,
) error {
	sortErrs := make([]error, 0)
	sort.Slice(migrations, func(i, j int) bool {
		var comparison bool
		var err error

		if ms.cfg.Migrations.VersionStyle == configloader.VersionStyleSequential {
			comparison, err = ms.compareSequentialMigrations(
				migrations[i],
				migrations[j],
				order,
			)
		} else {
			comparison, err = ms.compareTimestampMigrations(migrations[i], migrations[j], order)
		}

		if err != nil {
			sortErrs = append(sortErrs, err)
		}

		return comparison
	})
	if len(sortErrs) != 0 {
		return fmt.Errorf("unable to sort migrations: %v", sortErrs)
	}
	return nil
}

func (ms MigrationService) compareSequentialMigrations(
	m1 *models.Migration,
	m2 *models.Migration,
	order sortOrder,
) (bool, error) {
	m1Version, err := strconv.ParseInt(m1.Version, 10, 64)
	if err != nil {
		return false, err
	}

	m2Version, err := strconv.ParseInt(m2.Version, 10, 64)
	if err != nil {
		return false, err
	}

	if order == SortOrderAsc {
		return m1Version < m2Version, nil
	} else {
		return m1Version > m2Version, nil
	}
}

func (ms MigrationService) compareTimestampMigrations(
	m1 *models.Migration,
	m2 *models.Migration,
	order sortOrder,
) (bool, error) {
	m1Version, err := time.Parse("20060102150405", m1.Version)
	if err != nil {
		return false, err
	}

	m2Version, err := time.Parse("20060102150405", m2.Version)
	if err != nil {
		return false, err
	}

	if order == SortOrderAsc {
		return m1Version.Before(m2Version), nil
	} else {
		return m1Version.After(m2Version), nil
	}
}
