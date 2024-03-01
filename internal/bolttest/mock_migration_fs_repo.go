package bolttest

import "github.com/eugenetriguba/bolt/internal/models"

type MockMigrationFsRepo struct {
	CreateReturnValue              CreateReturnValue
	CreateCallCount                int
	ExistsReturnValue              ExistsReturnValue
	ExistsCallCount                int
	GetReturnValue                 GetReturnValue
	GetCallCount                   int
	ListReturnValue                ListReturnValue
	ListCallCount                  int
	LatestReturnValue              LatestReturnValue
	LatestCallCount                int
	ReadUpgradeScriptReturnValue   ReadUpgradeScriptReturnValue
	ReadUpgradeScriptCallCount     int
	ReadDowngradeScriptReturnValue ReadDowngradeScriptReturnValue
	ReadDowngradeScriptCallCount   int
}

type CreateReturnValue struct {
	Err error
}

type ExistsReturnValue struct {
	Exists bool
	Err    error
}

type GetReturnValue struct {
	Migration *models.Migration
	Err       error
}

type LatestReturnValue = GetReturnValue

type ReadUpgradeScriptReturnValue struct {
	ScriptContents string
	Err            error
}

type ReadDowngradeScriptReturnValue = ReadUpgradeScriptReturnValue

func (repo *MockMigrationFsRepo) Create(migration *models.Migration) error {
	repo.CreateCallCount += 1
	return repo.CreateReturnValue.Err
}

func (repo *MockMigrationFsRepo) Exists(version string) (bool, error) {
	repo.ExistsCallCount += 1
	return repo.ExistsReturnValue.Exists, repo.ExistsReturnValue.Err
}

func (repo *MockMigrationFsRepo) Get(version string) (*models.Migration, error) {
	repo.GetCallCount += 1
	return repo.GetReturnValue.Migration, repo.GetReturnValue.Err
}

func (repo *MockMigrationFsRepo) List() (map[string]*models.Migration, error) {
	repo.ListCallCount += 1
	return repo.ListReturnValue.Migrations, repo.ListReturnValue.Err
}

func (repo *MockMigrationFsRepo) Latest() (*models.Migration, error) {
	repo.LatestCallCount += 1
	return repo.LatestReturnValue.Migration, repo.LatestReturnValue.Err
}

func (repo *MockMigrationFsRepo) ReadUpgradeScript(
	migration *models.Migration,
) (string, error) {
	repo.ReadUpgradeScriptCallCount += 1
	return repo.ReadUpgradeScriptReturnValue.ScriptContents, repo.ReadUpgradeScriptReturnValue.Err
}

func (repo *MockMigrationFsRepo) ReadDowngradeScript(
	migration *models.Migration,
) (string, error) {
	repo.ReadDowngradeScriptCallCount += 1
	return repo.ReadDowngradeScriptReturnValue.ScriptContents, repo.ReadDowngradeScriptReturnValue.Err
}
