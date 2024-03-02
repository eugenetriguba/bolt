package repositories_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewMigrationDBRepo_CreatesTable(t *testing.T) {
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

func TestNewMigrationDBRepo_TableAlreadyExists(t *testing.T) {
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

func TestNewMigrationDBRepo_TableCreationErr(t *testing.T) {
	expectedErr := errors.New("error!")
	db := bolttest.MockSqlDb{
		ExecReturnValue: bolttest.ExecReturnValue{
			Err: expectedErr,
		},
	}

	_, err := repositories.NewMigrationDBRepo(&db)

	assert.ErrorIs(t, err, expectedErr)
}

func TestList_EmptyTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	migrations, err := repo.List()
	assert.Nil(t, err)
	assert.Equal(t, len(migrations), 0)
}

func TestList_QueryErr(t *testing.T) {
	expectedErr := errors.New("error!")
	db := bolttest.MockSqlDb{
		QueryReturnValue: bolttest.QueryReturnValue{
			Err: expectedErr,
		},
	}
	repo, err := repositories.NewMigrationDBRepo(&db)
	assert.Nil(t, err)

	_, err = repo.List()

	assert.ErrorIs(t, err, expectedErr)
}

func TestList_SingleResult(t *testing.T) {
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

func TestList_ShortVersion(t *testing.T) {
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

func TestIsApplied_WithNotApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	version := "20230101010101"
	applied, err := repo.IsApplied(version)
	assert.Nil(t, err)
	assert.Equal(t, applied, false)
}

func TestIsApplied_WithApplied(t *testing.T) {
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

func TestApply(t *testing.T) {
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

func TestApply_MalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.Apply("this is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, false)
}

func TestApplyWithTx_BeginErr(t *testing.T) {
	expectedErr := errors.New("error!")
	db := bolttest.MockSqlDb{
		BeginReturnValue: bolttest.BeginReturnValue{
			Err: expectedErr,
		},
	}
	repo, err := repositories.NewMigrationDBRepo(&db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.ApplyWithTx("SELECT 1 FROM bolt_migrations;", migration)

	assert.ErrorIs(t, err, expectedErr)
}

func TestApplyWithTx_ExecErr(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.ApplyWithTx("SELECT 1 FROM abc123donotexist;", migration)

	assert.ErrorContains(t, err, `relation "abc123donotexist" does not exist`)
}

func TestApplyWithTx_SuccessfullyApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	err = repo.ApplyWithTx(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`, migration)
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

func TestRevert(t *testing.T) {
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

func TestRevert_MalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")
	migration.Applied = true

	err = repo.Revert("this is not SQL", migration)
	assert.ErrorContains(t, err, "syntax error")
	assert.Equal(t, migration.Applied, true)
}

func TestRevertWithTx_BeginErr(t *testing.T) {
	expectedErr := errors.New("error!")
	db := bolttest.MockSqlDb{
		BeginReturnValue: bolttest.BeginReturnValue{
			Err: expectedErr,
		},
	}
	repo, err := repositories.NewMigrationDBRepo(&db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.RevertWithTx("DROP TABLE bolt_migrations;", migration)

	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertWithTx_ExecErr(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	assert.Nil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.RevertWithTx("DROP TABLE abc123donotexist;", migration)

	assert.ErrorContains(t, err, `table "abc123donotexist" does not exist`)
}

func TestRevertWithTx_SuccessfullyReverted(t *testing.T) {
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

	err = repo.RevertWithTx(`DROP TABLE tmp;`, migration)
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
