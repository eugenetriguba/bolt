package repositories_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/checkmate/assert"
)

func assertFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.Nil(t, err)
}

func TestNewMigrationFsRepo_CreatesDirIfNotExists(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}

	_, err := repositories.NewMigrationFsRepo(&migrationsConfig)

	assert.Nil(t, err)
	assertFileExists(t, migrationsDir)
}

func TestNewMigrationFsRepo_MigrationDirIsFile(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	_, err := os.Create(migrationsDir)
	assert.Nil(t, err)

	_, err = repositories.NewMigrationFsRepo(&migrationsConfig)

	assert.ErrorContains(t, err, "is not a directory")
}

func TestNewMigrationFsRepo_UnknownStatErr(t *testing.T) {
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: "\000x"}

	_, err := repositories.NewMigrationFsRepo(&migrationsConfig)

	assert.ErrorContains(t, err, "unable to check if migration directory")
}

func TestCreate_SuccessfullyCreated(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "add users table")

	err = repo.Create(migration)
	assert.Nil(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	assertFileExists(t, filepath.Join(tempDir, migrationDirName))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "upgrade.sql"))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "downgrade.sql"))
}

func TestCreate_FailsToCreateDir(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "add users table")
	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	_, err = os.Create(filepath.Join(tempDir, migrationDirName))
	assert.Nil(t, err)

	err = repo.Create(migration)

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "file exists")
}

func TestReadUpgradeScript_SuccessfullyRead(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)

	migration := models.NewSequentialMigration(1, "add users table")
	err = repo.Create(migration)
	assert.Nil(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	expectedUpgradeScriptContents := []byte("CREATE TABLE users(id int PRIMARY KEY);")
	os.WriteFile(
		filepath.Join(tempDir, migrationDirName, "upgrade.sql"),
		expectedUpgradeScriptContents,
		0755,
	)

	upgradeScriptContents, err := repo.ReadUpgradeScript(migration)
	assert.Nil(t, err)
	assert.Equal(t, upgradeScriptContents, string(expectedUpgradeScriptContents))
}

func TestReadUpgradeScript_FileDoesNotExist(t *testing.T) {
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: t.TempDir()}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	migration := models.NewSequentialMigration(1, "add users table")

	_, err = repo.ReadUpgradeScript(migration)

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestReadDowngradeScript_SuccessfullyRead(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)

	migration := models.NewSequentialMigration(1, "add users table")
	err = repo.Create(migration)
	assert.Nil(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	expectedDowngradeScriptContents := []byte("DROP TABLE users;")
	os.WriteFile(
		filepath.Join(tempDir, migrationDirName, "downgrade.sql"),
		expectedDowngradeScriptContents,
		0755,
	)

	downgradeScriptContents, err := repo.ReadDowngradeScript(migration)
	assert.Nil(t, err)
	assert.Equal(t, downgradeScriptContents, string(expectedDowngradeScriptContents))
}

func TestReadDowngradeScript_FileDoesNotExist(t *testing.T) {
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: t.TempDir()}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	migration := models.NewSequentialMigration(1, "add users table")

	_, err = repo.ReadDowngradeScript(migration)

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestList_Success(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)

	migration1 := models.NewTimestampMigration(
		time.Date(2020, 10, 12, 1, 1, 1, 1, time.UTC),
		"migration_1",
	)
	migration2 := models.NewTimestampMigration(
		time.Date(2022, 10, 12, 1, 1, 1, 1, time.UTC),
		"migration_2",
	)
	err = repo.Create(migration1)
	assert.Nil(t, err)
	repo.Create(migration2)
	assert.Nil(t, err)

	migrations, err := repo.List()
	assert.Nil(t, err)
	assert.Equal(t, len(migrations), 2)
	assert.DeepEqual(t, migrations[migration1.Version], migration1)
	assert.DeepEqual(t, migrations[migration2.Version], migration2)
}

func TestList_DirDoesNotExist(t *testing.T) {
	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	err = os.RemoveAll(migrationsDir)
	assert.Nil(t, err)

	_, err = repo.List()

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestList_InvalidMigrationName(t *testing.T) {
	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	err = os.Mkdir(filepath.Join(migrationsDir, "invalid"), 0755)
	assert.Nil(t, err)

	_, err = repo.List()

	assert.NotNil(t, err)
	assert.ErrorContains(
		t,
		err,
		"expected a migration directory of the format <version>_<message>",
	)
}

func TestLatest_DirDoesNotExist(t *testing.T) {
	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)
	err = os.RemoveAll(migrationsDir)
	assert.Nil(t, err)

	_, err = repo.Latest()

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestLatest_NoMigrations(t *testing.T) {
	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	assert.Nil(t, err)

	latestMigration, err := repo.Latest()

	assert.Nil(t, err)
	assert.Nil(t, latestMigration)
}
