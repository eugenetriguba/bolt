package main

import (
	"os"

	"github.com/eugenetriguba/bolt/internal/cli"
)

func main() {
	os.Exit(cli.Run())
}
