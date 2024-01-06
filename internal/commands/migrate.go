package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/bolt/internal/storage"
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
	cfg, err := configloader.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	db, err := storage.DBConnect(
		cfg.Connection.Driver,
		storage.DBConnectionString(&cfg.Connection),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	defer db.Close()

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
		if !migration.Applied {
			fmt.Printf("Applying migration for %s..\n", migration.Dirname())
			err = migrationRepo.Apply(migration)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return subcommands.ExitFailure
			}
			fmt.Printf("%s [x]\n", migration.Dirname())
		}
	}

	return subcommands.ExitSuccess
}
