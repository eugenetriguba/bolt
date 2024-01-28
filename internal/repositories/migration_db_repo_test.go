package repositories_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/checkmate/assert"
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
	assert.Nil(t, err)

	row = db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'bolt_migrations'
	`)
	err = row.Scan(&scanResult)
	assert.Nil(t, err)
}

func TestMigrationDBRepo_NewMigrationDBRepoLeavesCurrentTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	_, err := db.Exec(`CREATE TABLE bolt_migrations(id INT NOT NULL PRIMARY KEY)`)
	assert.Nil(t, err)
	_, err = db.Exec(`INSERT INTO bolt_migrations(id) VALUES (1);`)
	assert.Nil(t, err)

	_, err = repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	row := db.QueryRow(`SELECT id FROM bolt_migrations`)
	var scanResult int
	err = row.Scan(&scanResult)
	assert.Nil(t, err)
	assert.Equal(t, scanResult, 1)
}

func TestMigrationDBRepo_ListWithEmptyTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	migrations, err := repo.List()
	assert.Nil(t, err)
	assert.Equal(t, len(migrations), 0)
}

func TestMigrationDBRepo_ListWithSingleResult(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	version := "20230101000000"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.Nil(t, err)

	migrations, err := repo.List()
	assert.Nil(t, err)
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
	assert.Nil(t, err)

	version := "20230101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.Nil(t, err)

	migrations, err := repo.List()
	assert.Nil(t, err)
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
	assert.Nil(t, err)

	version := "20230101010101"
	applied, err := repo.IsApplied(version)
	assert.Nil(t, err)
	assert.Equal(t, applied, false)
}

func TestMigrationDBRepo_IsAppliedWithApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	version := "20230101010101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	assert.Nil(t, err)

	applied, err := repo.IsApplied(version)
	assert.Nil(t, err)
	assert.Equal(t, applied, true)
}

func TestMigrationDBRepo_Apply(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	err = repo.Apply(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`, migration)
	assert.Nil(t, err)
	assert.Equal(t, migration.Applied, true)

	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'tmp'
	`)
	var scanResult int
	err = row.Scan(&scanResult)
	assert.Nil(t, err)
	applied, err := repo.IsApplied(migration.Version)
	assert.Nil(t, err)
	assert.Equal(t, applied, true)
}

func TestMigrationDBRepo_Revert(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	_, err = db.Exec(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`)
	assert.Nil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	_, err = db.Exec(
		`INSERT INTO bolt_migrations(version) VALUES ($1);`,
		migration.Version,
	)
	assert.Nil(t, err)
	migration.Applied = true

	err = repo.Revert(`DROP TABLE tmp;`, migration)
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}

func TestMigrationDBRepo_ApplyAndRevertWithoutTransaction(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	_, err = db.Exec("CREATE TABLE tmp(id INT PRIMARY KEY, id2 INT, id3 INT)")
	assert.Nil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	// CREATE INDEX CONCURRENTLY cannot run inside a transaction.
	err = repo.Apply(
		`-- bolt: no-transaction\CREATE INDEX CONCURRENTLY i1 ON tmp(id2);`,
		migration,
	)
	assert.Nil(t, err)
	assert.Equal(t, migration.Applied, true)

	err = repo.Revert(
		"-- bolt: no-transaction\nCREATE INDEX CONCURRENTLY i2 ON tmp(id3);",
		migration,
	)
	assert.Nil(t, err)
	assert.Equal(t, migration.Applied, false)
}

func TestMigrationDBRepo_ApplyMalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.Apply("this is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, false)

	err = repo.Apply("-- bolt: no-transaction\nthis is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, false)
}

func TestMigrationDBRepo_RevertMalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")
	migration.Applied = true

	err = repo.Revert("this is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, true)

	err = repo.Revert("-- bolt: no-transaction\nthis is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, true)
}
