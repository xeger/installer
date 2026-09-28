//go:build windows

package clikit

import (
	"os"

	"golang.org/x/sys/windows"
)

// colorPrepare enables virtual terminal processing when f is a console, so
// ANSI sequences render instead of printing as text. A non-console f (such
// as a pipe with CLICOLOR_FORCE set) needs nothing.
func colorPrepare(f *os.File) bool {
	h := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return true
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	return windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
