package commands

import (
	"context"
	"flag"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/output"
	"github.com/eugenetriguba/bolt/internal/repositories"
	"github.com/eugenetriguba/bolt/internal/services"
	"github.com/eugenetriguba/bolt/internal/storage"
	"github.com/google/subcommands"
)

type DownCmd struct{}

func (*DownCmd) Name() string {
	return "down"
}

func (*DownCmd) Synopsis() string {
	return "downgrade migrations against the database"
}

func (*DownCmd) Usage() string {
	return `down:
	Downgrade migrations against the database.
  `
}

func (m *DownCmd) SetFlags(f *flag.FlagSet) {}

func (m *DownCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...interface{},
) subcommands.ExitStatus {
	consoleOutputter := output.ConsoleOutputter{}

	cfg, err := configloader.NewConfig()
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}

	db, err := storage.DBConnect(
		cfg.Connection.Driver,
		storage.DBConnectionString(&cfg.Connection),
	)
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}
	defer db.Close()

	migrationDBRepo, err := repositories.NewMigrationDBRepo(db)
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}

	migrationFsRepo, err := repositories.NewMigrationFsRepo(cfg.MigrationsDir)
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}

	migrationService := services.NewMigrationService(
		migrationDBRepo,
		migrationFsRepo,
		consoleOutputter,
	)
	err = migrationService.RevertAllMigrations()
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
