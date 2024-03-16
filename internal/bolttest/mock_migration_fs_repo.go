package bolttest

import (
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/sqlparse"
)

type MockMigrationFsRepo struct {
	CreateReturnValue              CreateReturnValue
	CreateCallCount                int
	ListReturnValue                ListReturnValue
	ListCallCount                  int
	ReadUpgradeScriptReturnValue   ReadUpgradeScriptReturnValue
	ReadUpgradeScriptCallCount     int
	ReadDowngradeScriptReturnValue ReadDowngradeScriptReturnValue
	ReadDowngradeScriptCallCount   int
}

type CreateReturnValue struct {
	Err error
}

type ReadUpgradeScriptReturnValue struct {
	Script sqlparse.MigrationScript
	Err    error
}

type ReadDowngradeScriptReturnValue = ReadUpgradeScriptReturnValue

func (repo *MockMigrationFsRepo) Create(migration *models.Migration) error {
	repo.CreateCallCount += 1
	return repo.CreateReturnValue.Err
}

func (repo *MockMigrationFsRepo) List() (map[string]*models.Migration, error) {
	repo.ListCallCount += 1
	return repo.ListReturnValue.Migrations, repo.ListReturnValue.Err
}

func (repo *MockMigrationFsRepo) ReadUpgradeScript(
	migration *models.Migration,
) (sqlparse.MigrationScript, error) {
	repo.ReadUpgradeScriptCallCount += 1
	return repo.ReadUpgradeScriptReturnValue.Script, repo.ReadUpgradeScriptReturnValue.Err
}

func (repo *MockMigrationFsRepo) ReadDowngradeScript(
	migration *models.Migration,
) (sqlparse.MigrationScript, error) {
	repo.ReadDowngradeScriptCallCount += 1
	return repo.ReadDowngradeScriptReturnValue.Script, repo.ReadDowngradeScriptReturnValue.Err
}
