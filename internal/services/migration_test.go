package services

import (
	"errors"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/checkmate/assert"
	"github.com/eugenetriguba/checkmate/check"
)

func TestListMigrations(t *testing.T) {
	type test struct {
		localFilesystemMigrations []*models.Migration
		remoteAppliedMigrations   []*models.Migration
		cfg                       configloader.Config
		sortOrder                 sortOrder
		expectedMigrations        []*models.Migration
	}

	tests := []test{
		// Ensure no local or remote applied migrations means none are returned.
		{
			localFilesystemMigrations: []*models.Migration{},
			remoteAppliedMigrations:   []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder:          SortOrderAsc,
			expectedMigrations: []*models.Migration{},
		},
		// Ensure one local fs migration shows up as expected.
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "001", Applied: false},
			},
			remoteAppliedMigrations: []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder: SortOrderAsc,
			expectedMigrations: []*models.Migration{
				{Version: "001", Applied: false},
			},
		},
		// Ensure that one local fs migration and one remote applied migration
		// for the same version returns one applied migration.
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "001", Applied: false},
			},
			remoteAppliedMigrations: []*models.Migration{
				{Version: "001", Applied: true},
			},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder: SortOrderAsc,
			expectedMigrations: []*models.Migration{
				{Version: "001", Applied: true},
			},
		},
		// Ensure that any applied migrations that don't exist locally
		// do not show up in the response.
		{
			localFilesystemMigrations: []*models.Migration{},
			remoteAppliedMigrations: []*models.Migration{
				{Version: "001", Applied: true},
			},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder:          SortOrderAsc,
			expectedMigrations: []*models.Migration{},
		},
		// Ensure that migrations are sorted in asc order for sequential migrations
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "20000"},
				{Version: "10000"},
				{Version: "1010"},
				{Version: "1009"},
				{Version: "190"},
				{Version: "110"},
				{Version: "001"},
			},
			remoteAppliedMigrations: []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder: SortOrderAsc,
			expectedMigrations: []*models.Migration{
				{Version: "001"},
				{Version: "110"},
				{Version: "190"},
				{Version: "1009"},
				{Version: "1010"},
				{Version: "10000"},
				{Version: "20000"},
			},
		},
		// Ensure that migrations are sorted in desc order for sequential migrations
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "001"},
				{Version: "110"},
				{Version: "190"},
				{Version: "1009"},
				{Version: "1010"},
				{Version: "10000"},
				{Version: "20000"},
			},
			remoteAppliedMigrations: []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleSequential,
				},
			},
			sortOrder: SortOrderDesc,
			expectedMigrations: []*models.Migration{
				{Version: "20000"},
				{Version: "10000"},
				{Version: "1010"},
				{Version: "1009"},
				{Version: "190"},
				{Version: "110"},
				{Version: "001"},
			},
		},
		// Ensure that migrations are sorted in asc order for timestamp migrations
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "20080509220905"},
				{Version: "20080509220405"},
				{Version: "20080509150405"},
				{Version: "20080502150405"},
				{Version: "20080102150405"},
				{Version: "20070102150405"},
				{Version: "20060102150405"},
			},
			remoteAppliedMigrations: []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleTimestamp,
				},
			},
			sortOrder: SortOrderAsc,
			expectedMigrations: []*models.Migration{
				{Version: "20060102150405"},
				{Version: "20070102150405"},
				{Version: "20080102150405"},
				{Version: "20080502150405"},
				{Version: "20080509150405"},
				{Version: "20080509220405"},
				{Version: "20080509220905"},
			},
		},
		// Ensure that migrations are sorted in desc order for timestamp migrations
		{
			localFilesystemMigrations: []*models.Migration{
				{Version: "20060102150405"},
				{Version: "20070102150405"},
				{Version: "20080102150405"},
				{Version: "20080502150405"},
				{Version: "20080509150405"},
				{Version: "20080509220405"},
				{Version: "20080509220905"},
			},
			remoteAppliedMigrations: []*models.Migration{},
			cfg: configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: configloader.VersionStyleTimestamp,
				},
			},
			sortOrder: SortOrderDesc,
			expectedMigrations: []*models.Migration{
				{Version: "20080509220905"},
				{Version: "20080509220405"},
				{Version: "20080509150405"},
				{Version: "20080502150405"},
				{Version: "20080102150405"},
				{Version: "20070102150405"},
				{Version: "20060102150405"},
			},
		},
	}

	for _, tc := range tests {
		localFsMigrationMap := make(map[string]*models.Migration, 0)
		for _, localFsMigration := range tc.localFilesystemMigrations {
			localFsMigrationMap[localFsMigration.Version] = localFsMigration
		}
		migrationFsRepo := &bolttest.MockMigrationFsRepo{
			ListReturnValue: bolttest.ListReturnValue{
				Migrations: localFsMigrationMap,
				Err:        nil,
			},
		}

		remoteAppliedMigrationMap := make(map[string]*models.Migration, 0)
		for _, remoteAppliedMigration := range tc.remoteAppliedMigrations {
			remoteAppliedMigrationMap[remoteAppliedMigration.Version] = remoteAppliedMigration
		}
		migrationDbRepo := &bolttest.MockMigrationDBRepo{
			ListReturnValue: bolttest.ListReturnValue{
				Migrations: remoteAppliedMigrationMap,
				Err:        nil,
			},
		}

		svc := NewMigrationService(
			migrationDbRepo,
			migrationFsRepo,
			tc.cfg,
			bolttest.NullOutputter{},
		)
		migrations, err := svc.ListMigrations(tc.sortOrder)

		assert.Nil(t, err)
		assert.DeepEqual(t, migrations, tc.expectedMigrations)
	}
}

func TestListMigrations_FsRepoListError(t *testing.T) {
	expectedErr := errors.New("fs repo error")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        nil,
		},
	}

	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)
	migrations, err := svc.ListMigrations(SortOrderAsc)

	assert.ErrorIs(t, err, expectedErr)
	var expectedMigrations []*models.Migration
	assert.DeepEqual(t, migrations, expectedMigrations)
}

func TestListMigrations_DbRepoListError(t *testing.T) {
	expectedErr := errors.New("db repo error")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}

	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)
	migrations, err := svc.ListMigrations(SortOrderAsc)

	assert.ErrorIs(t, err, expectedErr)
	var expectedMigrations []*models.Migration
	assert.DeepEqual(t, migrations, expectedMigrations)
}

func TestListMigrations_ParseErrDuringSort(t *testing.T) {
	type test struct {
		m1           models.Migration
		m2           models.Migration
		versionStyle configloader.VersionStyle
	}

	tests := []test{
		{
			m1:           models.Migration{Version: "abc"},
			m2:           models.Migration{Version: "123"},
			versionStyle: configloader.VersionStyleSequential,
		},
		{
			m1:           models.Migration{Version: "123"},
			m2:           models.Migration{Version: "abc"},
			versionStyle: configloader.VersionStyleSequential,
		},
		{
			m1:           models.Migration{Version: "20060102150405"},
			m2:           models.Migration{Version: "abc"},
			versionStyle: configloader.VersionStyleTimestamp,
		},
		{
			m1:           models.Migration{Version: "abc"},
			m2:           models.Migration{Version: "20060102150405"},
			versionStyle: configloader.VersionStyleTimestamp,
		},
	}

	for _, tc := range tests {
		migrationFsRepo := &bolttest.MockMigrationFsRepo{
			ListReturnValue: bolttest.ListReturnValue{
				Migrations: map[string]*models.Migration{
					tc.m1.Version: {Version: tc.m1.Version},
					tc.m2.Version: {Version: tc.m2.Version},
				},
				Err: nil,
			},
		}
		migrationDbRepo := &bolttest.MockMigrationDBRepo{
			ListReturnValue: bolttest.ListReturnValue{
				Migrations: nil,
				Err:        nil,
			},
		}

		svc := NewMigrationService(
			migrationDbRepo,
			migrationFsRepo,
			configloader.Config{
				Migrations: configloader.MigrationsConfig{
					VersionStyle: tc.versionStyle,
				},
			},
			bolttest.NullOutputter{},
		)
		migrations, err := svc.ListMigrations(SortOrderAsc)

		assert.ErrorContains(t, err, "unable to sort migrations")
		var expectedMigrations []*models.Migration
		assert.DeepEqual(t, migrations, expectedMigrations)
	}
}

func TestCreateMigration_VersionStyleTimestamp(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleTimestamp,
			},
		},
		bolttest.NullOutputter{},
	)

	beforeCreateTs := time.Now().Add(30 * -time.Second)
	migration, err := svc.CreateMigration("new migration")
	afterCreateTs := time.Now().Add(30 * time.Second)

	assert.Nil(t, err)
	version, err := time.ParseInLocation("20060102150405", migration.Version, time.Local)
	assert.Nil(t, err)
	check.True(t, beforeCreateTs.Before(version))
	check.True(t, afterCreateTs.After(version))
	check.Equal(t, migration.Message, "new migration")
	check.Equal(t, migration.Applied, false)
}

func TestCreateMigration_VersionStyleSequential(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.Nil(t, err)
	check.Equal(t, migration.Version, "001")
	check.Equal(t, migration.Message, "new migration")
	check.Equal(t, migration.Applied, false)
}

func TestCreateMigration_NilDbRepo(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	svc := NewMigrationService(
		nil,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.Nil(t, err)
	check.Equal(t, migration.Version, "001")
	check.Equal(t, migration.Message, "new migration")
	check.Equal(t, migration.Applied, false)
}

func TestCreateMigration_CreateErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, migration)
}

func TestCreateMigration_LatestErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Err: expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, migration)
}

func TestCreateMigration_SequentialVersionIncrements(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: map[string]*models.Migration{
				"001": {Version: "001"},
			},
			Err: nil,
		},
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.Nil(t, err)
	assert.Equal(t, migration.Version, "002")
}

func TestCreateMigration_SequentialVersionIncrementsLatestVersion(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: map[string]*models.Migration{
				"001": {Version: "001"},
				"002": {Version: "002"},
				"003": {Version: "003"},
			},
			Err: nil,
		},
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.Nil(t, err)
	assert.Equal(t, migration.Version, "004")
}

func TestCreateMigration_SequentialVersionIncrementsParsingErr(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: map[string]*models.Migration{
				"abc": {Version: "abc"},
			},
			Err: nil,
		},
		CreateReturnValue: bolttest.CreateReturnValue{
			Err: nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	migration, err := svc.CreateMigration("new migration")

	assert.ErrorContains(t, err, `parsing "abc": invalid syntax`)
	assert.Nil(t, migration)
}

func TestApplyMigration_WithTransaction(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadUpgradeScriptReturnValue: bolttest.ReadUpgradeScriptReturnValue{
			ScriptContents: "CREATE TABLE tmp(id INT PRIMARY KEY, id2 INT, id3 INT);",
			Err:            nil,
		},
	}

	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		ApplyWithTxReturnValue: bolttest.ApplyWithTxReturnValue{
			Err: nil,
		},
	}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyMigration(&models.Migration{})

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.ApplyWithTxCallCount, 1)
	assert.Equal(t, migrationDbRepo.ApplyCallCount, 0)
}

func TestApplyMigration_WithoutTransaction(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadUpgradeScriptReturnValue: bolttest.ReadUpgradeScriptReturnValue{
			ScriptContents: "-- bolt: no-transaction\nCREATE TABLE tmp(id INT PRIMARY KEY, id2 INT, id3 INT);",
			Err:            nil,
		},
	}

	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		ApplyReturnValue: bolttest.ApplyReturnValue{
			Err: nil,
		},
	}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyMigration(&models.Migration{})

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.ApplyWithTxCallCount, 0)
	assert.Equal(t, migrationDbRepo.ApplyCallCount, 1)
}

func TestApplyMigration_ReadUpgradeScriptErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadUpgradeScriptReturnValue: bolttest.ReadUpgradeScriptReturnValue{
			ScriptContents: "",
			Err:            expectedErr,
		},
	}

	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyMigration(&models.Migration{})

	assert.ErrorIs(t, err, expectedErr)
}

func TestApplyAllMigrations_ListMigrationsErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyAllMigrations()

	assert.ErrorIs(t, err, expectedErr)
}

func TestApplyAllMigrations_AppliesNothingIfAllApplied(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
		"002": {Version: "002", Applied: true},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyAllMigrations()

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.ApplyCallCount, 0)
	assert.Equal(t, migrationDbRepo.ApplyWithTxCallCount, 0)
}

func TestApplyAllMigrations_ApplyMigrationErr(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
	}
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
		ReadUpgradeScriptReturnValue: bolttest.ReadUpgradeScriptReturnValue{
			Err: expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyAllMigrations()

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, migrationDbRepo.ApplyCallCount, 0)
	assert.Equal(t, migrationDbRepo.ApplyWithTxCallCount, 0)
}

func TestApplyUpToVersion_ListMigrationsErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyUpToVersion("001")

	assert.ErrorIs(t, err, expectedErr)
}

func TestApplyUpToVersion_TargetMigrationNotFound(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyUpToVersion("002")

	assert.ErrorContains(t, err, "migration with version 002 does not exist")
}

func TestApplyUpToVersion_TargetMigrationAlreadyApplied(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyUpToVersion("001")

	assert.ErrorContains(t, err, "migration with version 001 is already applied")
}

func TestApplyUpToVersion_ApplyMigrationErr(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
	}
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
		ReadUpgradeScriptReturnValue: bolttest.ReadUpgradeScriptReturnValue{
			Err: expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyUpToVersion("001")

	assert.ErrorIs(t, err, expectedErr)
}

func TestApplyUpToVersion_AppliesMigrations(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
		"002": {Version: "002", Applied: false},
		"003": {Version: "003", Applied: false},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.ApplyUpToVersion("002")

	assert.Nil(t, err)
	assert.Equal(t, migrationFsRepo.ReadUpgradeScriptCallCount, 2)
}

func TestRevertMigration_WithTransaction(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadDowngradeScriptReturnValue: bolttest.ReadDowngradeScriptReturnValue{
			ScriptContents: "DROP TABLE tmp;",
			Err:            nil,
		},
	}

	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		RevertWithTxReturnValue: bolttest.RevertWithTxReturnValue{
			Err: nil,
		},
	}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertMigration(&models.Migration{})

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.RevertWithTxCallCount, 1)
	assert.Equal(t, migrationDbRepo.RevertCallCount, 0)
}

func TestRevertMigration_WithoutTransaction(t *testing.T) {
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadDowngradeScriptReturnValue: bolttest.ReadDowngradeScriptReturnValue{
			ScriptContents: "-- bolt: no-transaction\nDROP TABLE tmp;",
			Err:            nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{
		RevertReturnValue: bolttest.RevertReturnValue{
			Err: nil,
		},
	}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertMigration(&models.Migration{})

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.RevertWithTxCallCount, 0)
	assert.Equal(t, migrationDbRepo.RevertCallCount, 1)
}

func TestRevertMigration_ReadUpgradeScriptErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ReadDowngradeScriptReturnValue: bolttest.ReadDowngradeScriptReturnValue{
			ScriptContents: "",
			Err:            expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertMigration(&models.Migration{})

	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertDownToVersion_ListMigrationsErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertDownToVersion("001")

	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertDownToVersion_TargetMigrationNotFound(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertDownToVersion("002")

	assert.ErrorContains(t, err, "migration with version 002 does not exist")
}

func TestRevertDownToVersion_TargetMigrationNotApplied(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertDownToVersion("001")

	assert.ErrorContains(t, err, "migration with version 001 isn't applied")
}

func TestRevertDownToVersion_RevertMigrationErr(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
	}
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
		ReadDowngradeScriptReturnValue: bolttest.ReadDowngradeScriptReturnValue{
			Err: expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertDownToVersion("001")

	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertDownToVersion_RevertsMigrations(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
		"002": {Version: "002", Applied: true},
		"003": {Version: "003", Applied: false},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.RevertDownToVersion("001")

	assert.Nil(t, err)
	assert.Equal(t, migrationFsRepo.ReadDowngradeScriptCallCount, 2)
}

func TestRevertAllMigrations_ListMigrationsErr(t *testing.T) {
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: nil,
			Err:        expectedErr,
		},
	}
	svc := NewMigrationService(
		&bolttest.MockMigrationDBRepo{},
		migrationFsRepo,
		configloader.Config{},
		bolttest.NullOutputter{},
	)

	err := svc.RevertAllMigrations()

	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertAllMigrations_RevertsNothingIfNoneApplied(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: false},
		"002": {Version: "002", Applied: false},
	}
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.RevertAllMigrations()

	assert.Nil(t, err)
	assert.Equal(t, migrationDbRepo.RevertCallCount, 0)
	assert.Equal(t, migrationDbRepo.RevertWithTxCallCount, 0)
}

func TestRevertAllMigrations_RevertMigrationErr(t *testing.T) {
	migrations := map[string]*models.Migration{
		"001": {Version: "001", Applied: true},
	}
	expectedErr := errors.New("error!")
	migrationFsRepo := &bolttest.MockMigrationFsRepo{
		ListReturnValue: bolttest.ListReturnValue{
			Migrations: migrations,
			Err:        nil,
		},
		ReadDowngradeScriptReturnValue: bolttest.ReadDowngradeScriptReturnValue{
			Err: expectedErr,
		},
	}
	migrationDbRepo := &bolttest.MockMigrationDBRepo{}
	svc := NewMigrationService(
		migrationDbRepo,
		migrationFsRepo,
		configloader.Config{
			Migrations: configloader.MigrationsConfig{
				VersionStyle: configloader.VersionStyleSequential,
			},
		},
		bolttest.NullOutputter{},
	)

	err := svc.RevertAllMigrations()

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, migrationDbRepo.RevertCallCount, 0)
	assert.Equal(t, migrationDbRepo.RevertWithTxCallCount, 0)
}
