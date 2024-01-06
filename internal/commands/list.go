package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/bolt/internal/repositories"
<<<<<<< Updated upstream
	"github.com/eugenetriguba/bolt/internal/storage"
=======
	"github.com/eugenetriguba/bolt/internal/services"
>>>>>>> Stashed changes
	"github.com/google/subcommands"
)

type ListCmd struct{}

func (*ListCmd) Name() string { return "list" }

func (*ListCmd) Synopsis() string { return "List the database migrations and their statuses." }
func (*ListCmd) Usage() string {
	return `list:
	List the database migrations and their statuses.
  `
}

func (m *ListCmd) SetFlags(f *flag.FlagSet) {}

<<<<<<< Updated upstream
func (m *ListCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cfg, err := configloader.NewConfig()
=======
func (m *ListCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...interface{},
) subcommands.ExitStatus {
	c, err := config.NewConfig()
>>>>>>> Stashed changes
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

<<<<<<< Updated upstream
	db, err := storage.DBConnect(
		cfg.Connection.Driver,
		storage.DBConnectionString(&cfg.Connection),
	)
=======
	db, err := db.Connect(c)
>>>>>>> Stashed changes
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	defer db.Close()

<<<<<<< Updated upstream
	migrationRepo, err := repositories.NewMigrationRepo(db, cfg)
=======
	migrationRepo, err := repositories.NewMigrationRepo(db)
>>>>>>> Stashed changes
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	migrationService, err := services.NewMigrationService(migrationRepo, c)

	migrations, err := migrationService.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	for _, migration := range migrations {
		fmt.Println(migration)
	}

	return subcommands.ExitSuccess
}
