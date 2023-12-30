package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/db"
	"github.com/google/subcommands"
)

type CreateCmd struct{}

func (*CreateCmd) Name() string     { return "create" }
func (*CreateCmd) Synopsis() string { return "Create a new database migration." }
func (*CreateCmd) Usage() string {
	return `create:
	Create a new database migration.
  `
}

func (m *CreateCmd) SetFlags(f *flag.FlagSet) {}

func (m *CreateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
