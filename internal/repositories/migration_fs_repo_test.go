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
	"github.com/eugenetriguba/checkmate"
)

func assertFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	checkmate.AssertNil(t, err)
}

func TestNewMigrationFsRepoCreatesDirIfNotExists(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}

	_, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	checkmate.AssertNil(t, err)

	assertFileExists(t, migrationsDir)
}

func TestNewMigrationFsRepoMigrationDirIsFile(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: migrationsDir}
	_, err := os.Create(migrationsDir)
	checkmate.AssertNil(t, err)

	_, err = repositories.NewMigrationFsRepo(&migrationsConfig)
	checkmate.AssertErrorContains(t, err, "is not a directory")
}

func TestMigrationFsRepoCreate(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	checkmate.AssertNil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "add users table")

	err = repo.Create(migration)
	checkmate.AssertNil(t, err)

	migrationDirName := fmt.Sprintf("%s_add_users_table", migration.Version)
	assertFileExists(t, filepath.Join(tempDir, migrationDirName))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "upgrade.sql"))
	assertFileExists(t, filepath.Join(tempDir, migrationDirName, "downgrade.sql"))
}

func TestMigrationFsRepoReadUpgradeAndDowngradeScript(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	checkmate.AssertNil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "add users table")
	err = repo.Create(migration)
	checkmate.AssertNil(t, err)

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
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, upgradeScriptContents, string(expectedUpgradeScriptContents))
	downgradeScriptContents, err := repo.ReadDowngradeScript(migration)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, downgradeScriptContents, string(expectedDowngradeScriptContents))
}

func TestMigrationFsRepo_List(t *testing.T) {
	tempDir := t.TempDir()
	migrationsConfig := configloader.MigrationsConfig{DirectoryPath: tempDir}
	repo, err := repositories.NewMigrationFsRepo(&migrationsConfig)
	checkmate.AssertNil(t, err)

	migration1 := models.NewTimestampMigration(
		time.Date(2020, 10, 12, 1, 1, 1, 1, time.UTC),
		"migration_1",
	)
	migration2 := models.NewTimestampMigration(
		time.Date(2022, 10, 12, 1, 1, 1, 1, time.UTC),
		"migration_2",
	)
	err = repo.Create(migration1)
	checkmate.AssertNil(t, err)
	repo.Create(migration2)
	checkmate.AssertNil(t, err)

	migrations, err := repo.List()
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, len(migrations), 2)
	checkmate.AssertDeepEqual(t, migrations[migration1.Version], migration1)
	checkmate.AssertDeepEqual(t, migrations[migration2.Version], migration2)
}
