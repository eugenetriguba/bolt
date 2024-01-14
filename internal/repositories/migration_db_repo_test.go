package repositories_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"gotest.tools/v3/assert"
)

func TestMigrationDBRepo_NewMigrationDBRepoCreatesTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'bolt_migrations'
	`)
	var scanResult int
	err := row.Scan(&scanResult)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	_, err = repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	row = db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'bolt_migrations'
	`)
	err = row.Scan(&scanResult)
	assert.NilError(t, err)
}

func TestMigrationDBRepo_NewMigrationDBRepoLeavesCurrentTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	_, err := db.Exec(`CREATE TABLE bolt_migrations(id INT NOT NULL PRIMARY KEY)`)
	assert.NilError(t, err)
	_, err = db.Exec(`INSERT INTO bolt_migrations(id) VALUES (1);`)
	assert.NilError(t, err)

	_, err = repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	row := db.QueryRow(`SELECT id FROM bolt_migrations`)
	var scanResult int
	err = row.Scan(&scanResult)
	assert.NilError(t, err)
	assert.Equal(t, scanResult, 1)
}

func TestMigrationDBRepo_ListWithEmptyTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	migrations, err := repo.List()
	assert.NilError(t, err)
	assert.Equal(t, len(migrations), 0)
}

func TestMigrationDBRepo_ListWithSingleResult(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	version := "20230101000000"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.NilError(t, err)

	migrations, err := repo.List()
	assert.NilError(t, err)
	assert.Equal(t, len(migrations), 1)
	assert.DeepEqual(
		t,
		migrations[version],
		&models.Migration{Version: version, Message: "", Applied: true},
	)
}

func TestMigrationDBRepo_ListWithShortVersion(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	version := "20230101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.NilError(t, err)

	migrations, err := repo.List()
	assert.NilError(t, err)
	assert.Equal(t, len(migrations), 1)
	assert.DeepEqual(
		t,
		migrations[version],
		&models.Migration{Version: version, Message: "", Applied: true},
	)
}

func TestMigrationDBRepo_IsAppliedWithNotApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	version := "20230101010101"
	applied, err := repo.IsApplied(version)
	assert.NilError(t, err)
	assert.Equal(t, applied, false)
}

func TestMigrationDBRepo_IsAppliedWithApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	version := "20230101010101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.NilError(t, err)

	applied, err := repo.IsApplied(version)
	assert.NilError(t, err)
	assert.Equal(t, applied, true)
}

func TestMigrationDBRepo_Apply(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	migration := models.NewMigration(time.Now(), "test")
	err = repo.Apply(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`, migration)
	assert.NilError(t, err)
	assert.Equal(t, migration.Applied, true)

	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'tmp'
	`)
	var scanResult int
	err = row.Scan(&scanResult)
	assert.NilError(t, err)
	applied, err := repo.IsApplied(migration.Version)
	assert.NilError(t, err)
	assert.Equal(t, applied, true)
}

func TestMigrationDBRepo_Revert(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.NilError(t, err)

	_, err = db.Exec(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`)
	assert.NilError(t, err)

	migration := models.NewMigration(time.Now(), "test")
	_, err = db.Exec(
		`INSERT INTO bolt_migrations(version) VALUES ($1);`,
		migration.Version,
	)
	assert.NilError(t, err)
	migration.Applied = true

	err = repo.Revert(`DROP TABLE tmp;`, migration)
	assert.NilError(t, err)
	assert.Equal(t, migration.Applied, false)

	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'tmp'
	`)
	var scanResult int
	err = row.Scan(&scanResult)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	row = db.QueryRow(`SELECT count(*) FROM bolt_migrations`)
	var count int
	err = row.Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 0)
}