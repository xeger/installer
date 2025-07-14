# CrossnoKaye Command Line Interface

This is a meta-CLI launcher and version manager written in Go that dynamically discovers and executes subcommands. The CLI acts as a lightweight launcher that finds binaries in a standardized directory structure, forwards arguments to them, and manages installation/upgrades via GitHub releases.

## One-Click Install

Download and run a suitable installer for your platform.

- Mac OS X
    - [All processors](https://github.com/xeger/ck/releases/latest/download/CrossnoKaye-CLI.pkg)
- Windows
    - [Intel processors](https://github.com/xeger/ck/releases/latest/download/CrossnoKaye-CLI-x64.msi)
    - [ARM processor](https://github.com/xeger/ck/releases/latest/download/CrossnoKaye-CLI-arm64.msi)

## Features

- Discover and execute subcommands from GitHub releases
- Manage installation/upgrades via GitHub releases
- Dynamic help generation

## Built-in Commands

**install** - `ck install <command>`
- Searches GitHub repositories: `crossnokaye/cli-{command}`, then `crossnokaye/{command}`
- Downloads latest release asset matching platform: `{command}_{tag}_{GOOS}_{GOARCH}.tar.gz`
- Extracts to `$XDG_DATA_HOME/crossnokaye/cli/cmd/{command}/`
- Writes version tracking info to `release.json`
- Requires GitHub CLI (`gh`) to be installed

**upgrade** - `ck upgrade <command>`
- Upgrades specified subcommand to latest release
- Checks current version and skips if already up-to-date
- Updates version tracking info

## GitHub CLI Integration

The system requires GitHub CLI (`gh`) and provides helpful error messages when missing:
- Link to https://cli.github.com/ for installation
- Uses `gh release view` and `gh release download` commands
- Handles repository not found errors gracefully
