//go:build !windows

package clikit

import "path/filepath"

// xdgDefaultBase returns the Unix default location for kind: a path under
// the user's home directory. This build tag also covers macOS -- we
// deliberately use this same XDG-style layout (~/.config, ~/.local/state,
// ~/.cache, ~/.local/share) there rather than the ~/Library/* locations
// os.UserConfigDir would choose, so tools have one
// consistent directory layout across Linux and macOS.
func xdgDefaultBase(kind dirKind) (string, error) {
	home, err := dirUserHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, kind.unixRel...)...), nil
}
