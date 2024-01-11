package repositories_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"gotest.tools/v3/assert"
)

func assertFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.NilError(t, err)
}

func TestNewMigrationFsRepoCreatesDirIfNotExists(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")

	_, err := repositories.NewMigrationFsRepo(migrationsDir)
	assert.NilError(t, err)

	assertFileExists(t, migrationsDir)
}

func TestNewMigrationFsRepoMigrationDirIsFile(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	_, err := os.Create(migrationsDir)
	assert.NilError(t, err)

	_, err = repositories.NewMigrationFsRepo(migrationsDir)
	assert.ErrorContains(t, err, "is not a directory")
}

func TestMigrationFsRepoCreate(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := repositories.NewMigrationFsRepo(tempDir)
	assert.NilError(t, err)
	migration := models.NewMigration(time.Now(), "add users table")

	err = repo.Create(migration)
	assert.NilError(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	assertFileExists(t, filepath.Join(tempDir, migrationDirName))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "upgrade.sql"))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "downgrade.sql"))
}

func TestMigrationFsRepoReadUpgradeAndDowngradeScript(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := repositories.NewMigrationFsRepo(tempDir)
	assert.NilError(t, err)

	migration := models.NewMigration(time.Now(), "add users table")
	err = repo.Create(migration)
	assert.NilError(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	expectedUpgradeScriptContents := []byte("CREATE TABLE users(id int PRIMARY KEY);")
	os.WriteFile(
		filepath.Join(tempDir, migrationDirName, "upgrade.sql"),
		expectedUpgradeScriptContents,
		0755,
	)
	expectedDowngradeScriptContents := []byte("DROP TABLE users;")
	os.WriteFile(
		filepath.Join(tempDir, migrationDirName, "downgrade.sql"),
		expectedDowngradeScriptContents,
		0755,
	)

	upgradeScriptContents, err := repo.ReadUpgradeScript(migration)
	assert.NilError(t, err)
	assert.Equal(t, upgradeScriptContents, string(expectedUpgradeScriptContents))
	downgradeScriptContents, err := repo.ReadDowngradeScript(migration)
	assert.NilError(t, err)
	assert.Equal(t, downgradeScriptContents, string(expectedDowngradeScriptContents))
}

func TestMigrationFsRepo_List(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := repositories.NewMigrationFsRepo(tempDir)
	assert.NilError(t, err)

	migration1 := models.NewMigration(time.Date(2020, 10, 12, 1, 1, 1, 1, time.UTC), "migration_1")
	migration2 := models.NewMigration(time.Date(2022, 10, 12, 1, 1, 1, 1, time.UTC), "migration_2")
	err = repo.Create(migration1)
	assert.NilError(t, err)
	repo.Create(migration2)
	assert.NilError(t, err)

	migrations, err := repo.List()
	assert.NilError(t, err)
	assert.Equal(t, len(migrations), 2)
	assert.DeepEqual(t, migrations[migration1.Version], migration1)
	assert.DeepEqual(t, migrations[migration2.Version], migration2)
}
