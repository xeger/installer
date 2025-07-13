package main

import (
	"fmt"
	"os"

	"github.com/crossnokaye/cli/internal/commands"
	"github.com/crossnokaye/cli/internal/launcher"
)

func main() {
	if len(os.Args) < 2 {
		launcher.PrintUsage(os.Args[0])
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	if subcommand == "--help" || subcommand == "help" {
		launcher.PrintUsage(os.Args[0])
		return
	}

	if subcommand == "install" {
		if err := commands.Install(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if subcommand == "upgrade" {
		if err := commands.Upgrade(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := launcher.Execute(subcommand, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}