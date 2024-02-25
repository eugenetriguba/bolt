package commands

import (
	"context"
	"flag"

	"github.com/eugenetriguba/bolt/internal/output"
	"github.com/google/subcommands"
)

type VersionCmd struct{}

func (*VersionCmd) Name() string {
	return "version"
}

func (*VersionCmd) Synopsis() string {
	return "show the current version of bolt"
}

func (*VersionCmd) Usage() string {
	return `version:
	Show the current version of Bolt
  `
}

func (cmd *VersionCmd) SetFlags(f *flag.FlagSet) {}

func (cmd *VersionCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...interface{},
) subcommands.ExitStatus {
	consoleOutputter := output.NewConsoleOutputter()
	err := consoleOutputter.Output("bolt v0.2.1")
	if err != nil {
		consoleOutputter.Error(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
