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
		consoleOutputter.Error(fmt.Errorf("unable to retrieve configuration: %w", err))
		return subcommands.ExitFailure
	}

	db, err := storage.DBConnect(cfg.Connection)
	if err != nil {
		consoleOutputter.Error(fmt.Errorf("unable to connect to database: %w", err))
		return subcommands.ExitFailure
	}
	defer db.Session.Close()

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

	migrations, err := migrationService.ListMigrations(services.SortOrderAsc)
	if err != nil {
		consoleOutputter.Error(fmt.Errorf("unable to list migrations: %w", err))
		return subcommands.ExitFailure
	}

	if len(migrations) == 0 {
		consoleOutputter.Output(
			"No migrations have been created.\n" +
				"Run 'bolt new' to create your first migration.",
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

	err = consoleOutputter.Table(headers, rows)
	if err != nil {
		consoleOutputter.Error(
			fmt.Errorf("unable to output migrations as table: %w", err),
		)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
