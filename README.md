# xn

`xn` is the [xeger.net](https://xeger.net) command-line launcher: `xn foo args...` makes sure the `foo` tool is installed and current, then runs it with `args`.

It is also a working reference for shipping a Go command-line tool the way users expect on every platform:

- **Installable everywhere.** A macOS `.pkg`, Windows `.msi` (x64 and ARM64, adds itself to `PATH`), and static `.tar.gz` archives for macOS, Linux, and Windows on amd64 and arm64.
- **Statically linked.** Every binary is built with `CGO_ENABLED=0`, so it runs on any distribution and needs no runtime.
- **Chained launching with auto-upgrade.** `xn foo` installs `foo` on first use, checks for a newer release at most once a week, and then hands the process over to it (`exec` on Unix; on Windows it waits and exits with the tool's status). `xn` also tells you when a newer `xn` is out.

## Install

Download the installer for your platform from the [latest release](https://github.com/xeger/installer/releases/latest):

| Platform | Asset |
|----------|-------|
| macOS (Intel and Apple silicon) | `xn.pkg` |
| Windows x64 | `xn-x64.msi` |
| Windows ARM64 | `xn-arm64.msi` |
| Linux, or anywhere without an installer | `xn_<version>_<os>_<arch>.tar.gz` — unpack `xn` onto your `PATH` |

The installers are not yet signed. On macOS, the first attempt to open `xn.pkg` is blocked; open **System Settings → Privacy & Security**, find the message about `xn.pkg`, and choose **Open Anyway**. On Windows, choose **More info → Run anyway** in SmartScreen.

## Pre-Requisites

`xn` downloads releases with the [GitHub CLI](https://cli.github.com) (`gh`), which must be installed and signed in:

```sh
brew install gh                   # macOS
winget install --id GitHub.cli    # Windows
gh auth login
```

If `gh` is missing or signed out, `xn` tells you how to fix it.

## Usage

- `xn <tool> [args...]` installs `<tool>` if needed, checks for updates at most weekly, then runs it. The tool's exit status becomes `xn`'s.
- `xn install <owner>/<tool>` installs a tool from any GitHub repository; afterwards run it as `xn <tool>`. `xn` knows one tool by name out of the box: [`semdiff`](https://github.com/xeger/semdiff).
- `xn upgrade [tool...]` upgrades the named tools, or every installed tool.
- `xn version` and `xn help` do what you'd expect.

## Publishing a Tool

`xn install <owner>/<tool>` works with any repository whose releases attach one asset per platform named `<tool>_<tag>_<GOOS>_<GOARCH>.tar.gz` (e.g. `semdiff_v1.2.3_darwin_arm64.tar.gz`), containing an executable named `<tool>` (`<tool>.exe` on Windows) anywhere in the archive. `xn`'s own releases follow the same convention.
