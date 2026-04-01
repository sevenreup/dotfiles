package main

import (
	"doot/src/internal/commands"
	"os"
)

func main() {
	if err := commands.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
