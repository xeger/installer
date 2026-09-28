package clikit

import (
	"os"

	"golang.org/x/term"
)

// ColorEnabled reports whether output written to f may use ANSI color and
// styling. It encodes terminal policy only; clikit does not render output.
//
// The policy, in order of precedence:
//
//   - NO_COLOR set to any non-empty value disables color (https://no-color.org).
//   - CLICOLOR_FORCE set to a non-empty value other than "0" enables color
//     even when f is not a terminal, e.g. for piped output in CI.
//   - Otherwise color is enabled only when f is a terminal and TERM is not
//     "dumb".
//
// On Windows, when the answer is yes and f is a console, ColorEnabled also
// turns on the console's virtual terminal processing so ANSI sequences
// render; if that fails it reports false. Call it once per stream at
// startup rather than per write.
func ColorEnabled(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	forced := colorForced()
	if f == nil {
		return forced
	}
	if !forced && (os.Getenv("TERM") == "dumb" || !term.IsTerminal(int(f.Fd()))) {
		return false
	}
	return colorPrepare(f)
}

// colorForced reports whether CLICOLOR_FORCE requests color.
func colorForced() bool {
	v := os.Getenv("CLICOLOR_FORCE")
	return v != "" && v != "0"
}
