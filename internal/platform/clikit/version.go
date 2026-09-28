package clikit

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// version is set at build time via:
//
//	go build -ldflags "-X github.com/xeger/installer/internal/platform/clikit.version=v1.2.3"
//
// The linker silently ignores -X flags whose target does not exist, so a
// misspelled symbol (or a leftover "-X main.Tag=...") produces no error;
// Version just falls back to the build info. SetVersion is the runtime
// alternative for tools that already stamp their own variable.
var version string

// versionReadBuildInfo is a seam over debug.ReadBuildInfo for testing.
var versionReadBuildInfo = debug.ReadBuildInfo

// SetVersion overrides the version reported by Version. It is intended for
// tools that stamp their own variable at build time (e.g. "-X main.Tag=...")
// and call clikit.SetVersion(Tag) early in main. An empty v is ignored, so
// an unstamped build still falls back to the build info. SetVersion is not
// safe for concurrent use with Version.
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

// Version reports the build version of the running binary. It returns, in
// order of preference:
//
//   - the value from SetVersion or the -ldflags incantation documented on
//     the version variable;
//   - the main module version from the Go build info. "Main" is the module
//     of the binary being built (the tool), not clikit. Since Go 1.24 this
//     is filled for local builds in a VCS checkout as well as for
//     `go install example.com/tool@v1.2.3`: a tag such as "v1.2.3", a
//     pseudo-version for untagged commits, with "+dirty" for modified trees;
//   - "dev", e.g. for `go run` or -buildvcs=false builds.
func Version() string {
	if version != "" {
		return version
	}
	if info, ok := versionReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

// Commit returns the VCS revision the binary was built from, as recorded by
// runtime/debug.ReadBuildInfo (the "vcs.revision" build setting). It returns
// "" if that information is unavailable, e.g. for a binary built outside a
// VCS checkout or without Go's VCS stamping.
func Commit() string {
	info, ok := versionReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
}

// VersionInfo returns a human-readable summary of the running binary
// suitable for a --version flag: "<program> <version> (<commit>)". The
// program name is the basename of os.Args[0] without any ".exe" suffix, so
// tools launched through a wrapper (e.g. a launcher) report a clean
// name rather than a full path. The parenthetical commit is omitted when no
// commit is known.
func VersionInfo() string {
	program := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	v := Version()
	if c := Commit(); c != "" {
		return program + " " + v + " (" + c + ")"
	}
	return program + " " + v
}
