//go:build windows

// Package launcher runs an installed tool in place of the launcher.
package launcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

// Exec runs binary with the launcher's stdio and exits with its exit status.
// Windows has no exec(2), so the launcher waits for the tool instead of being
// replaced by it.
// It returns only if the tool cannot be started.
func Exec(binary string, args []string) error {
	// The tool outlives any launcher context: like exec(2) on Unix, nothing
	// cancels it, and Ctrl+C is the tool's to handle.
	cmd := exec.CommandContext(context.Background(), binary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// The console delivers Ctrl+C to the tool too; let it decide what to do.
	signal.Ignore(os.Interrupt)

	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		return fmt.Errorf("run %s: %w", binary, err)
	}
	os.Exit(0)
	return nil
}
