package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/db"
	"github.com/google/subcommands"
)

type MigrateCmd struct{}

func (*MigrateCmd) Name() string     { return "migrate" }
func (*MigrateCmd) Synopsis() string { return "Migrate the database to the latest migration." }
func (*MigrateCmd) Usage() string {
	return `migrate:
	Migrate the database to the latest migration.
  `
}

func (m *MigrateCmd) SetFlags(f *flag.FlagSet) {}

func (m *MigrateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	db, err := db.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	defer db.Close()

	sqlStatement := `
		CREATE TABLE IF NOT EXISTS applied_migration(
			version CHARACTER(32) PRIMARY KEY NOT NULL
		);
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
