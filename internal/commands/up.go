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

type UpCmd struct {
	version string
}

func (*UpCmd) Name() string {
	return "up"
}

func (*UpCmd) Synopsis() string {
	return "apply migrations against the database"
}

func (*UpCmd) Usage() string {
	return `up [-version|-v]:
	Apply migrations against the database
  `
}

func (cmd *UpCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(
		&cmd.version,
		"version",
		"",
		"The version to upgrade up and including to.",
	)
	f.StringVar(&cmd.version, "v", cmd.version, "alias for -version")
}

func (cmd *UpCmd) Execute(
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

	db, err := storage.DBConnect(
		cfg.Connection.Driver,
		storage.DBConnectionString(&cfg.Connection),
	)
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
		err = migrationService.ApplyAllMigrations()
		if err != nil {
			consoleOutputter.Error(fmt.Errorf("unable to apply all migrations: %w", err))
			return subcommands.ExitFailure
		}
	} else {
		err = migrationService.ApplyUpToVersion(cmd.version)
		if err != nil {
			consoleOutputter.Error(fmt.Errorf("unable to apply migrations up to %s: %w", cmd.version, err))
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
