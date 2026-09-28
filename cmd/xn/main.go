// Command main is the launcher: "<name> foo args..." makes sure the foo
// tool is installed and current, then runs it with args. Its name and
// identity come from internal/platform.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/xeger/installer/internal/commands"
	"github.com/xeger/installer/internal/launcher"
	"github.com/xeger/installer/internal/platform"
	"github.com/xeger/installer/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdout)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout io.Writer) int {
	ui := platform.UI
	if len(args) == 0 {
		usage(os.Stderr)
		return 2
	}

	switch cmd := args[0]; cmd {
	case "help", "-h", "--help":
		usage(stdout)
		return 0

	case "version", "--version":
		fmt.Fprintln(stdout, platform.VersionInfo())
		return 0

	case "install":
		if len(args) < 2 {
			ui.Error("Usage:", platform.Name, "install <tool|owner/tool>...")
			return 2
		}
		commands.NotifySelfUpdate(ctx)
		failed := false
		for _, tool := range args[1:] {
			if err := commands.Install(ctx, tool); err != nil {
				ui.ErrorDetail(err, "Could not install", tool)
				failed = true
			}
		}
		return exitCode(failed)

	case "upgrade":
		commands.NotifySelfUpdate(ctx)
		if err := commands.Upgrade(ctx, args[1:]); err != nil {
			ui.ErrorDetail(err, "Upgrade failed")
			return 1
		}
		return 0

	default:
		if strings.HasPrefix(cmd, "-") {
			ui.Error("Unknown option", cmd)
			usage(os.Stderr)
			return 2
		}
		commands.NotifySelfUpdate(ctx)
		bin, err := commands.Ensure(ctx, cmd)
		if err != nil {
			ui.ErrorDetail(err, "Cannot run", cmd)
			return 1
		}
		err = launcher.Exec(bin, args[1:])
		ui.ErrorDetail(err, "Cannot run", cmd)
		return 1
	}
}

func exitCode(failed bool) int {
	if failed {
		return 1
	}
	return 0
}

func usage(w io.Writer) {
	n := platform.Name
	fmt.Fprintf(w, `Usage: %[1]s <tool> [args...]

%[2]s

Commands:
  install <tool>...   Install the latest release of each tool
                      (use <owner>/<tool> for a tool %[1]s doesn't know yet)
  upgrade [tool...]   Upgrade the named tools, or all installed tools
  version             Show %[1]s's version
  help                Show this help
`, n, platform.Description)
	tools, err := store.List()
	if err != nil {
		fmt.Fprintf(w, "\nCould not list installed tools: %v\n", err)
		return
	}
	if len(tools) == 0 {
		fmt.Fprintf(w, "\nNo tools installed yet. Run '%[1]s <tool>' or '%[1]s install <tool>'.\n", n)
		return
	}
	fmt.Fprintln(w, "\nInstalled tools:")
	for _, t := range tools {
		tag := t.Tag
		if tag == "" {
			tag = "unknown version"
		}
		fmt.Fprintf(w, "  %-20s %s\n", t.Name, tag)
	}
}
