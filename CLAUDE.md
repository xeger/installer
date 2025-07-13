# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a meta-CLI launcher written in Go that dynamically discovers and executes subcommands. The CLI acts as a lightweight launcher that finds binaries in a standardized directory structure and forwards arguments to them.

## Architecture

**Core Components:**
- `cmd/ck/main.go` - CLI entrypoint that handles argument parsing and help requests
- `internal/launcher/launcher.go` - Core launcher logic with subcommand discovery and execution

**Key Design Patterns:**
- **Binary Discovery**: Looks for executables at `$XDG_STATE_HOME/crossnokaye/cli/cmd/{subcommand}/{subcommand}`
- **Argument Forwarding**: Passes all arguments after the subcommand directly to the target binary
- **Dynamic Help**: Scans the filesystem to generate help text with available subcommands
- **XDG Compliance**: Uses `XDG_STATE_HOME` with fallback to `~/.local/state`

## Build and Development Commands

```bash
# Build the main CLI
go build -o ck ./cmd/ck

# Run directly during development
go run ./cmd/ck [subcommand] [args...]

# Build for installation
go build -o ck ./cmd/ck
```

## Key Functions

**launcher.Execute(subcommand, args)** - Main execution function that:
1. Finds the binary path using `findBinary()`
2. Creates an `exec.Command` with stdio forwarding
3. Runs the target binary with provided arguments

**launcher.GetAvailableCommands()** - Filesystem scanner that:
1. Reads the commands directory structure
2. Validates that binaries exist and are executable
3. Returns sorted list of available subcommands

**launcher.PrintUsage()** - Dynamic help generator that:
1. Calls `GetAvailableCommands()` to get current subcommands
2. Formats usage information with discovered commands
3. Handles missing command directory gracefully

## Directory Structure Convention

The launcher expects subcommands to be installed as:
```
$XDG_STATE_HOME/crossnokaye/cli/cmd/
├── foo/
│   └── foo          # executable binary
├── bar/
│   └── bar          # executable binary
└── ...
```

## Future Architecture Notes

This codebase is designed to be extended with `install` and `upgrade` commands that will use GitHub packages and the GitHub CLI as a distribution mechanism. The current launcher provides the foundation for this plugin-style architecture.