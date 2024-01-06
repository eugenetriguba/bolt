package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/db"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/google/subcommands"
)

type ListCmd struct{}

func (*ListCmd) Name() string     { return "list" }
func (*ListCmd) Synopsis() string { return "List the database migrations and their statuses." }
func (*ListCmd) Usage() string {
	return `list:
	List the database migrations and their statuses.
  `
}

func (m *ListCmd) SetFlags(f *flag.FlagSet) {}

func (m *ListCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	db, err := db.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	defer db.Close()

	cfg, err := configloader.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	migrationRepo, err := repositories.NewMigrationRepo(db, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	migrations, err := migrationRepo.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	for _, migration := range migrations {
		fmt.Println(migration)
	}

	return subcommands.ExitSuccess
}
