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
	"github.com/xeger/installer/internal/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdout)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout io.Writer) int {
	out := ui.Default
	if len(args) == 0 {
		usage(os.Stderr)
		return 2
	}

	switch cmd := args[0]; cmd {
	case "help", "-h", "--help":
		usage(stdout)
		return 0

	case "version", "--version":
		_, _ = fmt.Fprintln(stdout, platform.VersionInfo()) // nowhere to report a failed write
		return 0

	case "install":
		if len(args) < 2 {
			out.Error("Usage:", platform.Name, "install <tool|owner/tool>...")
			return 2
		}
		commands.NotifySelfUpdate(ctx)
		failed := false
		for _, tool := range args[1:] {
			if err := commands.Install(ctx, tool); err != nil {
				out.ErrorDetail(err, "Could not install", tool)
				failed = true
			}
		}
		return exitCode(failed)

	case "upgrade":
		commands.NotifySelfUpdate(ctx)
		if err := commands.Upgrade(ctx, args[1:]); err != nil {
			out.ErrorDetail(err, "Upgrade failed")
			return 1
		}
		return 0

	default:
		if strings.HasPrefix(cmd, "-") {
			out.Error("Unknown option", cmd)
			usage(os.Stderr)
			return 2
		}
		commands.NotifySelfUpdate(ctx)
		bin, err := commands.Ensure(ctx, cmd)
		if err != nil {
			out.ErrorDetail(err, "Cannot run", cmd)
			return 1
		}
		err = launcher.Exec(bin, args[1:])
		out.ErrorDetail(err, "Cannot run", cmd)
		return 1
	}
}

func exitCode(failed bool) int {
	if failed {
		return 1
	}
	return 0
}

// usage writes help to w. Write errors are ignored: there is nowhere left
// to report them.
func usage(w io.Writer) {
	var b strings.Builder
	defer func() { _, _ = io.WriteString(w, b.String()) }()

	n := platform.Name
	fmt.Fprintf(&b, `Usage: %[1]s <tool> [args...]

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
		fmt.Fprintf(&b, "\nCould not list installed tools: %v\n", err)
		return
	}
	if len(tools) == 0 {
		fmt.Fprintf(&b, "\nNo tools installed yet. Run '%[1]s <tool>' or '%[1]s install <tool>'.\n", n)
		return
	}
	fmt.Fprintln(&b, "\nInstalled tools:")
	for _, t := range tools {
		tag := t.Tag
		if tag == "" {
			tag = "unknown version"
		}
		fmt.Fprintf(&b, "  %-20s %s\n", t.Name, tag)
	}
}
