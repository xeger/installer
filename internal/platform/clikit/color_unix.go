//go:build !windows

package clikit

import "os"

// colorPrepare readies f for ANSI output. Unix terminals need nothing.
func colorPrepare(*os.File) bool { return true }
