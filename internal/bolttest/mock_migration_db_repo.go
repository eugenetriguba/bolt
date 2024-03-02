package bolttest

import "github.com/eugenetriguba/bolt/internal/models"

type MockMigrationDBRepo struct {
	ListReturnValue         ListReturnValue
	ListCallCount           int
	IsAppliedReturnValue    IsAppliedReturnValue
	IsAppliedCallCount      int
	ApplyReturnValue        ApplyReturnValue
	ApplyCallCount          int
	ApplyWithTxReturnValue  ApplyWithTxReturnValue
	ApplyWithTxCallCount    int
	RevertReturnValue       RevertReturnValue
	RevertCallCount         int
	RevertWithTxReturnValue RevertWithTxReturnValue
	RevertWithTxCallCount   int
}

type ListReturnValue struct {
	Migrations map[string]*models.Migration
	Err        error
}

type IsAppliedReturnValue struct {
	IsApplied bool
	Err       error
}

type ApplyReturnValue struct {
	Err error
}

type ApplyWithTxReturnValue = ApplyReturnValue
type RevertReturnValue = ApplyReturnValue
type RevertWithTxReturnValue = ApplyReturnValue

func (repo *MockMigrationDBRepo) List() (map[string]*models.Migration, error) {
	repo.ListCallCount += 1
	return repo.ListReturnValue.Migrations, repo.ListReturnValue.Err
}

func (repo *MockMigrationDBRepo) IsApplied(version string) (bool, error) {
	repo.IsAppliedCallCount += 1
	return repo.IsAppliedReturnValue.IsApplied, repo.IsAppliedReturnValue.Err
}

func (repo *MockMigrationDBRepo) Apply(
	upgradeScript string,
	migration *models.Migration,
) error {
	repo.ApplyCallCount += 1
	return repo.ApplyReturnValue.Err
}

func (repo *MockMigrationDBRepo) ApplyWithTx(
	upgradeScript string,
	migration *models.Migration,
) error {
	repo.ApplyWithTxCallCount += 1
	return repo.ApplyWithTxReturnValue.Err
}

func (repo *MockMigrationDBRepo) Revert(
	downgradeScript string,
	migration *models.Migration,
) error {
	repo.RevertCallCount += 1
	return repo.RevertReturnValue.Err
}

func (repo *MockMigrationDBRepo) RevertWithTx(
	downgradeScript string,
	migration *models.Migration,
) error {
	repo.RevertWithTxCallCount += 1
	return repo.RevertWithTxReturnValue.Err
}
