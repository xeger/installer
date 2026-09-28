# Contributor's Guide

## Architecture

| Package | Responsibility |
|---------|----------------|
| `internal/platform` | Product identity (`Name`, `Description`, `SelfRepo`, `Candidates`) and services (`DataDir`, `StateDir`, `Version`, `UI`) |
| `internal/platform/clikit` | Per-user directories (XDG everywhere, XDG layout on macOS, AppData on Windows) and version stamping; `term` subpackage for terminal output |
| `cmd/xn` | Argument dispatch and help |
| `internal/commands` | `Ensure` (install if missing, weekly upgrade check), `Install`, `Upgrade`, self-update notice |
| `internal/store` | Installed-tool layout, name validation, `release.json` records |
| `internal/github` | `gh` wrapper: release lookup, download, and setup guidance |
| `internal/archive` | Extracts a tool's executable from its release tarball, atomically |
| `internal/launcher` | Replaces `xn` with the tool (`exec` on Unix; run-and-exit-with-status on Windows) |
| `internal/updates` | Semantic version comparison and the weekly check interval |

`internal/platform` is the only package that knows it is `xn`. Every other package is product-neutral, so the same launcher code can be reused under another name by supplying a different `internal/platform`. `scripts/check-sync <other-checkout>` verifies that a sibling launcher's shared code is identical (ignoring module path and the `cmd/<name>` directory). Keep product names, owners, and paths out of every other package.

## Files on Disk

```
$XDG_DATA_HOME/xn/tools/<tool>/<tool>[.exe]
$XDG_DATA_HOME/xn/tools/<tool>/release.json   # tag, repository, last update check
$XDG_STATE_HOME/xn/self-update.json           # last check for a newer xn
```

Defaults: `~/.local/share` and `~/.local/state` on macOS and Linux, `%AppData%` and `%LocalAppData%` on Windows.

## Development

```sh
go test -race ./...
go vet ./... && GOOS=windows go vet ./...
go run ./cmd/xn help
```

## Packaging

| Script | Output |
|--------|--------|
| `scripts/build-archives.sh <tag>` | `dist/xn_<tag>_<os>_<arch>.tar.gz` for darwin/linux/windows × amd64/arm64 |
| `scripts/build-darwin.sh <tag>` | `dist/xn.pkg` (both architectures; postinstall picks one and copies it to `/usr/local/bin`) |
| `scripts/build-windows.ps1 -Version <tag>` | `dist/xn-x64.msi`, `dist/xn-arm64.msi` (installs to Program Files, adds to `PATH`) |

All binaries are static (`CGO_ENABLED=0`) and stamped with `-ldflags "-X github.com/xeger/installer/internal/platform.version=<tag>"`. Publishing a GitHub release runs `.github/workflows/release.yml`, which builds and attaches everything.

WiX is pinned to 5.0.2: WiX 6 and later require the Open Source Maintenance Fee for commercial use, and v7 enforces it.
