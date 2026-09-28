//go:build windows

package clikit

import (
	"os"
	"path/filepath"
)

// xdgDefaultBase returns the Windows default location for kind: the
// platform environment variable (APPDATA for config/data, LOCALAPPDATA
// for state/cache) when it is set to an absolute path, else the
// corresponding AppData subdirectory of the user's home directory. A
// relative APPDATA/LOCALAPPDATA value is treated as unset.
func xdgDefaultBase(kind dirKind) (string, error) {
	if v := os.Getenv(kind.winEnv); v != "" && filepath.IsAbs(v) {
		return filepath.Clean(v), nil
	}
	home, err := dirUserHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, kind.winRel...)...), nil
}
