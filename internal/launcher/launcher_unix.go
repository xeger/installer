//go:build !windows

// Package launcher runs an installed tool in place of the launcher.
package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Exec replaces the launcher process with binary, so the tool inherits its
// stdio, signals, and exit status. It returns only if the exec fails.
func Exec(binary string, args []string) error {
	argv := append([]string{filepath.Base(binary)}, args...)
	if err := syscall.Exec(binary, argv, os.Environ()); err != nil { //nolint:gosec // running tools is the launcher's job
		return fmt.Errorf("exec %s: %w", binary, err)
	}
	return nil
}
