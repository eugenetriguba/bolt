package cli

import (
	"context"
	"flag"

	"github.com/eugenetriguba/bolt/internal/commands"
	"github.com/google/subcommands"
)

// Run runs the Bolt CLI and returns the exit code.
func Run() int {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&commands.NewCmd{}, "")
	subcommands.Register(&commands.UpCmd{}, "")
	subcommands.Register(&commands.DownCmd{}, "")
	subcommands.Register(&commands.StatusCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	return int(subcommands.Execute(ctx))
}
