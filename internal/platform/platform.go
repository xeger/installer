// Package platform holds everything that makes this launcher xn: its
// identity, where it finds tools, and the local clikit services it builds
// on.
//
// Every other package is product-neutral and shared verbatim with other
// launchers built from the same code; only this package differs. Its
// exported variables are swappable, which tests use to isolate behavior.
package platform

import "github.com/xeger/installer/internal/platform/clikit"

// version is stamped at build time:
//
//	go build -ldflags "-X github.com/xeger/installer/internal/platform.version=v1.2.3"
var version string

func init() {
	clikit.SetVersion(version)
}

var app = clikit.App{Name: "xn"}

var (
	// Name is the launcher's command and product name.
	Name = "xn"

	// Description is the one-line summary shown in help.
	Description = "Runs xeger.net tools, installing or updating them first as needed."

	// SelfRepo is the GitHub repository that publishes the launcher itself.
	SelfRepo = "xeger/installer"

	// Candidates returns the repositories that may publish tool, in lookup
	// order. A nil result means the launcher does not know the tool; users
	// can still install it by repository ("xn install <owner>/<tool>").
	Candidates = func(tool string) []string {
		switch tool {
		case "semdiff":
			return []string{"xeger/semdiff"}
		default:
			return nil
		}
	}

	// DataDir returns a path under the launcher's per-user data directory.
	DataDir = app.DataDir

	// StateDir returns a path under the launcher's per-user state directory.
	StateDir = app.StateDir

	// Version returns the launcher's version.
	Version = clikit.Version

	// VersionInfo returns a one-line "<program> <version> (<commit>)" summary.
	VersionInfo = clikit.VersionInfo

	// ColorEnabled reports whether output to a stream may use color.
	ColorEnabled = clikit.ColorEnabled
)
