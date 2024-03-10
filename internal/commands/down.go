package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/output"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/bolt/internal/services"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/google/subcommands"
)

type DownCmd struct {
	version string
}

func (*DownCmd) Name() string {
	return "down"
}

func (*DownCmd) Synopsis() string {
	return "downgrade migrations against the database"
}

func (*DownCmd) Usage() string {
	return `down [-version|-v]:
	Downgrade migrations against the database
  `
}

func (cmd *DownCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(
		&cmd.version,
		"version",
		"",
		"The version to downgrade down and including to.",
	)
	f.StringVar(&cmd.version, "v", cmd.version, "alias for -version")
}

func (cmd *DownCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...interface{},
) subcommands.ExitStatus {
	consoleOutputter := output.NewConsoleOutputter()

	cfg, err := configloader.NewConfig()
	if err != nil {
		consoleOutputter.Error(fmt.Errorf("unable to retrieve configuration: %w", err))
		return subcommands.ExitFailure
	}

	db, err := storage.NewDB(cfg.Connection)
	if err != nil {
		consoleOutputter.Error(fmt.Errorf("unable to connect to database: %w", err))
		return subcommands.ExitFailure
	}
	defer db.Close()

	migrationDBRepo, err := repositories.NewMigrationDBRepo(db)
	if err != nil {
		consoleOutputter.Error(err)
		return subcommands.ExitFailure
	}

	migrationFsRepo, err := repositories.NewMigrationFsRepo(&cfg.Migrations)
	if err != nil {
		consoleOutputter.Error(err)
		return subcommands.ExitFailure
	}

	migrationService := services.NewMigrationService(
		migrationDBRepo,
		migrationFsRepo,
		*cfg,
		consoleOutputter,
	)

	if cmd.version == "" {
		err = migrationService.RevertAllMigrations()
		if err != nil {
			consoleOutputter.Error(fmt.Errorf("unable to revert all migrations: %w", err))
			return subcommands.ExitFailure
		}
	} else {
		err = migrationService.RevertDownToVersion(cmd.version)
		if err != nil {
			consoleOutputter.Error(fmt.Errorf("unable to revert migrations down to %s: %w", cmd.version, err))
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
