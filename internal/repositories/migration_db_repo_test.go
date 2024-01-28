package repositories_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/models"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/checkmate"
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
	checkmate.AssertErrorIs(t, err, sql.ErrNoRows)

	_, err = repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	row = db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'bolt_migrations'
	`)
	err = row.Scan(&scanResult)
	checkmate.AssertNil(t, err)
}

func TestMigrationDBRepo_NewMigrationDBRepoLeavesCurrentTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	_, err := db.Exec(`CREATE TABLE bolt_migrations(id INT NOT NULL PRIMARY KEY)`)
	checkmate.AssertNil(t, err)
	_, err = db.Exec(`INSERT INTO bolt_migrations(id) VALUES (1);`)
	checkmate.AssertNil(t, err)

	_, err = repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	row := db.QueryRow(`SELECT id FROM bolt_migrations`)
	var scanResult int
	err = row.Scan(&scanResult)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, scanResult, 1)
}

func TestMigrationDBRepo_ListWithEmptyTable(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	migrations, err := repo.List()
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, len(migrations), 0)
}

func TestMigrationDBRepo_ListWithSingleResult(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	version := "20230101000000"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	checkmate.AssertNil(t, err)

	migrations, err := repo.List()
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, len(migrations), 1)
	checkmate.AssertDeepEqual(
		t,
		migrations[version],
		&models.Migration{Version: version, Message: "", Applied: true},
	)
}

func TestMigrationDBRepo_ListWithShortVersion(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	version := "20230101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	checkmate.AssertNil(t, err)

	migrations, err := repo.List()
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, len(migrations), 1)
	checkmate.AssertDeepEqual(
		t,
		migrations[version],
		&models.Migration{Version: version, Message: "", Applied: true},
	)
}

func TestMigrationDBRepo_IsAppliedWithNotApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	version := "20230101010101"
	applied, err := repo.IsApplied(version)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, applied, false)
}

func TestMigrationDBRepo_IsAppliedWithApplied(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	version := "20230101010101"
	_, err = db.Exec(`INSERT INTO bolt_migrations(version) VALUES ($1);`, version)
	checkmate.AssertNil(t, err)

	applied, err := repo.IsApplied(version)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, applied, true)
}

func TestMigrationDBRepo_Apply(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	err = repo.Apply(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`, migration)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, migration.Applied, true)

	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'tmp'
	`)
	var scanResult int
	err = row.Scan(&scanResult)
	checkmate.AssertNil(t, err)
	applied, err := repo.IsApplied(migration.Version)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, applied, true)
}

func TestMigrationDBRepo_Revert(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)

	_, err = db.Exec(`CREATE TABLE tmp(id INT NOT NULL PRIMARY KEY)`)
	checkmate.AssertNil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	_, err = db.Exec(
		`INSERT INTO bolt_migrations(version) VALUES ($1);`,
		migration.Version,
	)
	checkmate.AssertNil(t, err)
	migration.Applied = true

	err = repo.Revert(`DROP TABLE tmp;`, migration)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, migration.Applied, false)

	row := db.QueryRow(`
		SELECT 1 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'public' 
		AND TABLE_NAME = 'tmp'
	`)
	var scanResult int
	err = row.Scan(&scanResult)
	checkmate.AssertErrorIs(t, err, sql.ErrNoRows)

	row = db.QueryRow(`SELECT count(*) FROM bolt_migrations`)
	var count int
	err = row.Scan(&count)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, count, 0)
}

func TestMigrationDBRepo_ApplyAndRevertWithoutTransaction(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)
	_, err = db.Exec("CREATE TABLE tmp(id INT PRIMARY KEY, id2 INT, id3 INT)")
	checkmate.AssertNil(t, err)

	migration := models.NewTimestampMigration(time.Now(), "test")
	// CREATE INDEX CONCURRENTLY cannot run inside a transaction.
	err = repo.Apply(
		`-- bolt: no-transaction\CREATE INDEX CONCURRENTLY i1 ON tmp(id2);`,
		migration,
	)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, migration.Applied, true)

	err = repo.Revert(
		"-- bolt: no-transaction\nCREATE INDEX CONCURRENTLY i2 ON tmp(id3);",
		migration,
	)
	checkmate.AssertNil(t, err)
	checkmate.AssertEqual(t, migration.Applied, false)
}

func TestMigrationDBRepo_ApplyMalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")

	err = repo.Apply("this is not SQL", migration)
	checkmate.AssertErrorContains(t, err, "syntax error")
	checkmate.AssertEqual(t, migration.Applied, false)

	err = repo.Apply("-- bolt: no-transaction\nthis is not SQL", migration)
	checkmate.AssertErrorContains(t, err, "syntax error")
	checkmate.AssertEqual(t, migration.Applied, false)
}

func TestMigrationDBRepo_RevertMalformedSql(t *testing.T) {
	db := bolttest.NewTestDB(t, "postgres")
	repo, err := repositories.NewMigrationDBRepo(db)
	checkmate.AssertNil(t, err)
	migration := models.NewTimestampMigration(time.Now(), "test")
	migration.Applied = true

	err = repo.Revert("this is not SQL", migration)
	checkmate.AssertErrorContains(t, err, "syntax error")
	checkmate.AssertEqual(t, migration.Applied, true)

	err = repo.Revert("-- bolt: no-transaction\nthis is not SQL", migration)
	checkmate.AssertErrorContains(t, err, "syntax error")
	checkmate.AssertEqual(t, migration.Applied, true)
}
