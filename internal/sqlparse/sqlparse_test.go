package sqlparse_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/eugenetriguba/bolt/internal/sqlparse"
	"github.com/eugenetriguba/checkmate/assert"
	"github.com/eugenetriguba/checkmate/check"
)

func TestSqlParser_Parse(t *testing.T) {
	testCases := []struct {
		migration               string
		expectedUpgradeScript   sqlparse.MigrationScript
		expectedDowngradeScript sqlparse.MigrationScript
	}{
		{
			migration: "",
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- migrate:up
			CREATE TABLE users(id int PRIMARY KEY);`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- migrate:up transaction:false
			CREATE TABLE users(id int PRIMARY KEY);`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- migrate:up transaction:true
			CREATE TABLE users(id int PRIMARY KEY);`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- transaction:true migrate:up
			CREATE TABLE users(id int PRIMARY KEY);`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- migrate:down
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
		},
		{
			migration: `
			-- migrate:down transaction:false
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
		},
		{
			migration: `
			-- migrate:down transaction:true
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
		},
		{
			migration: `
			-- transaction:true migrate:down
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "",
				Options:  sqlparse.ExecutionOptions{},
			},
		},
		{
			migration: `
			-- migrate:up
			CREATE TABLE users(id int PRIMARY KEY);
			-- migrate:down
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: true},
			},
		},
		{
			migration: `
			-- migrate:up transaction:false
			CREATE TABLE users(id int PRIMARY KEY);
			-- migrate:down transaction:false
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
		},
		{
			migration: `
			-- MIGRATE:UP TRANSACTION:FALSE
			CREATE TABLE users(id int PRIMARY KEY);
			-- MIGRATE:DOWN TRANSACTION:FALSE
			DROP TABLE users;`,
			expectedUpgradeScript: sqlparse.MigrationScript{
				Contents: "CREATE TABLE users(id int PRIMARY KEY);\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
			expectedDowngradeScript: sqlparse.MigrationScript{
				Contents: "DROP TABLE users;\n",
				Options:  sqlparse.ExecutionOptions{UseTransaction: false},
			},
		},
	}

	for _, tc := range testCases {
		sqlParser := sqlparse.NewSqlParser()
		upgradeScript, downgradeScript, err := sqlParser.Parse(strings.NewReader(tc.migration))
		assert.Nil(t, err)
		check.DeepEqual(t, upgradeScript, tc.expectedUpgradeScript)
		check.DeepEqual(t, downgradeScript, tc.expectedDowngradeScript)
	}
}

type ErrReader struct {
	Reader  io.Reader
	ErrCond string
}

func (er *ErrReader) Read(p []byte) (int, error) {
	n, err := er.Reader.Read(p)
	if err != nil {
		return n, err
	}

	if bytes.Contains(p[:n], []byte(er.ErrCond)) {
		return n, errors.New("error: unwanted input encountered")
	}

	return n, nil
}

func TestSqlParserParseError(t *testing.T) {
	reader := &ErrReader{
		Reader:  strings.NewReader("    "),
		ErrCond: " ",
	}
	sqlParser := sqlparse.NewSqlParser()
	_, _, err := sqlParser.Parse(reader)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unwanted input encountered")
}
