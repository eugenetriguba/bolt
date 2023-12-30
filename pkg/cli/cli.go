package cli

import (
	"context"
	"flag"

	"github.com/eugenetriguba/bolt/internal/commands"
	"github.com/google/subcommands"
)

func Run() int {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&commands.MigrateCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	return int(subcommands.Execute(ctx))
}
