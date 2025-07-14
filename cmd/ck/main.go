// Test comment to verify Claude hook functionality
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/crossnokaye/cli/internal/commands"
	"github.com/crossnokaye/cli/internal/launcher"
	"github.com/crossnokaye/cli/internal/updatecheck"
)

func main() {
	// Check for updates (runs max once per week)
	updatecheck.CheckForUpdates()

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
		// Check if this is a "command not found" error
		if isCommandNotFoundError(err) {
			fmt.Printf("Command '%s' not found. Attempting to install...\n", subcommand)

			// Try to install the command
			if installErr := commands.Install([]string{subcommand}); installErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to install '%s': %v\n", subcommand, installErr)
				os.Exit(1)
			}

			// Installation successful, now try to execute the command again
			if execErr := launcher.Execute(subcommand, args); execErr != nil {
				fmt.Fprintf(os.Stderr, "Error executing '%s' after installation: %v\n", subcommand, execErr)
				os.Exit(1)
			}
			return
		}

		// For other errors, just print and exit
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isCommandNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "not found at")
}
