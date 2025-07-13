# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a meta-CLI launcher and version manager written in Go that dynamically discovers and executes subcommands. The CLI acts as a lightweight launcher that finds binaries in a standardized directory structure, forwards arguments to them, and manages installation/upgrades via GitHub releases.

## Architecture

**Core Components:**
- `cmd/ck/main.go` - CLI entrypoint that handles argument parsing, built-in commands, and help requests
- `internal/launcher/launcher.go` - Core launcher logic with subcommand discovery and execution
- `internal/commands/install.go` - Install subcommands from GitHub releases with multiple repository patterns
- `internal/commands/upgrade.go` - Upgrade subcommands or ck itself with version checking
- `internal/github/github.go` - GitHub CLI wrapper for release querying and asset downloading
- `internal/archive/extractor.go` - Streaming tar.gz extraction with in-place upgrade support
- `internal/version/version.go` - Version tracking database using release.json files

**Key Design Patterns:**
- **Binary Discovery**: Looks for executables at `$XDG_STATE_HOME/crossnokaye/cli/cmd/{subcommand}/{subcommand}`
- **Argument Forwarding**: Passes all arguments after the subcommand directly to the target binary
- **Dynamic Help**: Scans the filesystem to generate help text with available subcommands and versions
- **XDG Compliance**: Uses `XDG_STATE_HOME` with fallback to `~/.local/state`
- **Version Tracking**: Maintains `release.json` database files alongside binaries
- **Repository Patterns**: Tries multiple GitHub repository naming conventions (`crossnokaye/cli-{cmd}`, `crossnokaye/{cmd}`)
- **Self-Upgrade**: Can upgrade its own binary in-place with safety testing

## Build and Development Commands

```bash
# Build the main CLI
go build -o ck ./cmd/ck

# Run directly during development
go run ./cmd/ck [subcommand] [args...]

# Test built-in commands
go run ./cmd/ck help
go run ./cmd/ck install <command>
go run ./cmd/ck upgrade <command>

# Check for code issues
go vet ./...
```

ALWAYS CHECK FOR CODE ISSUES AFTER GENERATING CODE.

## Built-in Commands

**install** - `ck install <command>`
- Searches GitHub repositories: `crossnokaye/cli-{command}`, then `crossnokaye/{command}`
- Downloads latest release asset matching platform: `{command}_{tag}_{GOOS}_{GOARCH}.tar.gz`
- Extracts to `$XDG_STATE_HOME/crossnokaye/cli/cmd/{command}/`
- Writes version tracking info to `release.json`
- Requires GitHub CLI (`gh`) to be installed

**upgrade** - `ck upgrade <command>`
- Upgrades specified subcommand to latest release
- Checks current version and skips if already up-to-date
- Updates version tracking info

## Key Functions

**launcher.Execute(subcommand, args)** - Main execution function that:
1. Finds the binary path using `findBinary()`
2. Creates an `exec.Command` with stdio forwarding
3. Runs the target binary with provided arguments

**version.GetSortedCommandsWithVersions(stateHome)** - Version-aware command scanner that:
1. Reads the commands directory structure
2. Validates that binaries exist and are executable
3. Returns sorted list of commands with version info from `release.json`
4. Shows "unknown" for commands without version tracking

**launcher.PrintUsage()** - Dynamic help generator that:
1. Shows built-in commands (install, upgrade)
2. Lists installed subcommands with versions
3. Handles missing command directory gracefully

**github.GetLatestRelease(repo)** - GitHub API wrapper that:
1. Uses `gh release view` to get latest release info
2. Parses JSON response for tag name and assets
3. Returns structured release data

**archive.ExtractTarGzStream(reader, opts)** - Streaming extractor that:
1. File filtering for targeted extraction
2. Atomic binary replacement to prevent corruption

## Directory Structure Convention

The launcher expects subcommands to be installed as:
```
$XDG_STATE_HOME/crossnokaye/cli/cmd/
├── foo/
│   ├── foo          # executable binary
│   └── release.json # version tracking database
├── bar/
│   ├── bar          # executable binary
│   └── release.json # version tracking database
└── ...
```

## Version Tracking

Each command directory contains a `release.json` file:
```json
{
  "foo": {
    "tag_name": "v1.2.3",
    "repository": "crossnokaye/cli-foo",
    "timestamp": "1673123456"
  }
}
```

This enables:
- Version display in help output
- Skip upgrades when already current
- Repository tracking for multiple naming patterns

## GitHub CLI Integration

The system requires GitHub CLI (`gh`) and provides helpful error messages when missing:
- Link to https://cli.github.com/ for installation
- Uses `gh release view` and `gh release download` commands
- Handles repository not found errors gracefully
