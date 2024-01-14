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

type StatusCmd struct{}

func (*StatusCmd) Name() string {
	return "status"
}

func (*StatusCmd) Synopsis() string {
	return "list the database migrations and their statuses"
}

func (*StatusCmd) Usage() string {
	return `status:
	List the database migrations and their statuses
  `
}

func (m *StatusCmd) SetFlags(f *flag.FlagSet) {}

func (m *StatusCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...interface{},
) subcommands.ExitStatus {
	consoleOutputter := output.NewConsoleOutputter()

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
	migrations, err := migrationService.ListMigrations(services.SortOrderAsc)
	if err != nil {
		consoleOutputter.Error(err.Error())
		return subcommands.ExitFailure
	}

	if len(migrations) == 0 {
		consoleOutputter.Output(
			"No migrations have been created.\n" +
				"Run 'bolt create' to create your first migration.",
		)
		return subcommands.ExitSuccess
	}

	headers := []string{"Version", "Message", "Applied"}
	rows := make([][]string, len(migrations))
	for i, migration := range migrations {
		applied := ""
		if migration.Applied {
			applied = "X"
		}

		rows[i] = []string{migration.Version, migration.Message, applied}
	}

	consoleOutputter.Table(headers, rows)
	return subcommands.ExitSuccess
}
