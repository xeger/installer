package term

import (
	"io"
	"os"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	xterm "golang.org/x/term"
)

// levelStyle bundles a leveled message's symbol together with the styling
// applied to the symbol and to the message text.
type levelStyle struct {
	symbol      string
	symbolStyle lipgloss.Style
	textStyle   lipgloss.Style
}

// Level symbols. These are narrow, single-width glyphs (unlike the emoji set
// they replace, which included variation-selector characters such as ℹ️ and
// ⚠️ that render at inconsistent widths across terminals).
const (
	verboseSymbol = "·"
	infoSymbol    = "›"
	successSymbol = "✓"
	warnSymbol    = "!"
	errorSymbol   = "✗"
)

// Level styles. The symbol is always bold and colored. Message text is
// colored for verbose (faint gray, since it's low-priority debug chatter)
// and warn/error (yellow/red, to draw the eye to trouble); info and success
// message text is left at the default foreground color so it stays readable
// against any terminal background.
var (
	verboseLevel = levelStyle{
		symbol:      verboseSymbol,
		symbolStyle: lipgloss.NewStyle().Bold(true).Faint(true).Foreground(lipgloss.Color("8")),
		textStyle:   lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8")),
	}
	infoLevel = levelStyle{
		symbol:      infoSymbol,
		symbolStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
		textStyle:   lipgloss.NewStyle(),
	}
	successLevel = levelStyle{
		symbol:      successSymbol,
		symbolStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
		textStyle:   lipgloss.NewStyle(),
	}
	warnLevel = levelStyle{
		symbol:      warnSymbol,
		symbolStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")),
		textStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	}
	errorLevel = levelStyle{
		symbol:      errorSymbol,
		symbolStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")),
		textStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	}
)

// styleMessage renders msg (which may contain "\n") with level's symbol and
// styling. Unlike a single Style.Render call over the whole message, each
// line is styled individually: lipgloss pads every line of a multi-line
// Render to the widest line's width with trailing spaces and, separately,
// expands tabs to 4 spaces, neither of which is wanted for status text.
// Rendering line-by-line sidesteps both, since alignment padding only
// kicks in for a Render call that itself spans multiple lines. Tab
// conversion is disabled explicitly (lipgloss.NoTabConversion) so tabs
// pass through untouched. The symbol appears only on the first line;
// continuation lines are indented to align under the first line's text,
// using the symbol's display width (via lipgloss.Width, which measures
// cells rather than bytes or runes) plus one for the space that follows
// it. No line carries trailing whitespace.
func styleMessage(level levelStyle, msg string) string {
	symbolWidth := lipgloss.Width(level.symbol)
	indent := strings.Repeat(" ", symbolWidth+1)
	textStyle := level.textStyle.TabWidth(lipgloss.NoTabConversion)
	symbolRendered := level.symbolStyle.Render(level.symbol)

	var b strings.Builder
	for i, line := range strings.Split(msg, "\n") {
		if i > 0 {
			b.WriteString("\n")
		}
		switch {
		case i == 0 && line == "":
			// No message text at all: just the symbol, no trailing space.
			b.WriteString(symbolRendered)
		case i == 0:
			b.WriteString(symbolRendered)
			b.WriteString(" ")
			b.WriteString(textStyle.Render(line))
		case line == "":
			// Blank continuation line: nothing to indent.
		default:
			b.WriteString(indent)
			b.WriteString(textStyle.Render(line))
		}
	}
	return b.String()
}

// autoColor reports whether styling should be enabled for w when
// Options.Color was not set.
//
// NO_COLOR (see https://no-color.org), when set to any non-empty value,
// always disables color and takes priority over everything else, including
// a forcing CLICOLOR_FORCE.
//
// Otherwise, CLICOLOR_FORCE (set to a non-empty value other than "0")
// forces color on unconditionally — including when w is not a terminal, or
// isn't even an *os.File — since its purpose is to force color in contexts
// (e.g. piped output in CI) where auto-detection would normally say no.
//
// Absent either override, color is enabled only when w is an *os.File
// attached to a terminal (per golang.org/x/term) and TERM is not "dumb".
func autoColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	forced := cliColorForced()

	f, ok := w.(*os.File)
	if !ok {
		return forced
	}
	if forced {
		return true
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return xterm.IsTerminal(int(f.Fd()))
}

// cliColorForced reports whether CLICOLOR_FORCE is set to a non-empty value
// other than "0", per the convention originated by BSD's cliColor.
func cliColorForced() bool {
	v := os.Getenv("CLICOLOR_FORCE")
	return v != "" && v != "0"
}
