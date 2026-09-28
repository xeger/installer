package term

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	lipgloss "charm.land/lipgloss/v2"
)

// Options configures a Printer.
type Options struct {
	// Data is the stream written by Table and Data. It defaults to
	// os.Stdout.
	Data io.Writer
	// UI is the stream written by Verbose, Info, Success, Warn, Error and
	// ErrorDetail. It defaults to os.Stderr.
	UI io.Writer
	// Verbose enables output from the Verbose method. It can also be
	// changed later with SetVerbose.
	Verbose bool
	// Color overrides styling auto-detection. nil (the default) means
	// auto: styling is enabled when UI is an *os.File attached to a
	// terminal and the NO_COLOR environment variable is unset or empty
	// (see autoColor for the full set of rules, including TERM=dumb and
	// CLICOLOR_FORCE). A non-nil value forces styling on or off; it can
	// also be changed later with SetColor.
	Color *bool
}

// Printer prints leveled, styled status messages to a UI stream, and
// tabular or raw data to a separate Data stream. The split keeps a tool's
// data output (suitable for piping) free of status chatter. A *Printer is
// safe for concurrent use.
type Printer struct {
	mu      sync.Mutex
	data    io.Writer
	ui      io.Writer
	verbose bool
	color   bool
}

// New creates a Printer from opts. Data defaults to os.Stdout and UI
// defaults to os.Stderr when left nil.
func New(opts Options) *Printer {
	data := opts.Data
	if data == nil {
		data = os.Stdout
	}
	ui := opts.UI
	if ui == nil {
		ui = os.Stderr
	}

	color := autoColor(ui)
	if opts.Color != nil {
		color = *opts.Color
	}

	if color {
		// Get conhost's console mode into VT-processing so ANSI escapes
		// work there too. A no-op off Windows and for non-*os.File
		// streams.
		if f, ok := ui.(*os.File); ok {
			lipgloss.EnableLegacyWindowsANSI(f)
		}
	}

	return &Printer{
		data:    data,
		ui:      ui,
		verbose: opts.Verbose,
		color:   color,
	}
}

// Default is a ready-to-use Printer for tools that previously relied on a
// package-level global (such as the various "TUI" variables it replaces).
// It writes Data output to os.Stdout and UI output (Verbose, Info, Success,
// Warn, Error, ErrorDetail) to os.Stderr. Color is auto-detected once, at
// package init time (see autoColor); call Default.SetColor to override that
// later, for example after parsing a --no-color flag.
var Default = New(Options{})

// SetVerbose enables or disables output from Verbose.
func (p *Printer) SetVerbose(verbose bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.verbose = verbose
}

// Verbosity reports whether Verbose-level output is currently enabled.
func (p *Printer) Verbosity() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.verbose
}

// SetColor enables or disables styling (color and symbols), overriding
// whatever New auto-detected or Options.Color specified.
func (p *Printer) SetColor(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.color = enabled
}

// colorEnabled reports whether styling is currently enabled.
func (p *Printer) colorEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.color
}

// Verbose prints debug text to the UI stream, only when verbose output is
// enabled (see SetVerbose).
func (p *Printer) Verbose(args ...any) {
	if !p.Verbosity() {
		return
	}
	p.writeLevel(verboseLevel, args...)
}

// Info prints informational text to the UI stream.
func (p *Printer) Info(args ...any) {
	p.writeLevel(infoLevel, args...)
}

// Success prints a success message to the UI stream.
func (p *Printer) Success(args ...any) {
	p.writeLevel(successLevel, args...)
}

// Warn prints a warning message to the UI stream.
func (p *Printer) Warn(args ...any) {
	p.writeLevel(warnLevel, args...)
}

// Error prints an error message to the UI stream.
func (p *Printer) Error(args ...any) {
	p.writeLevel(errorLevel, args...)
}

// Table prints args to the Data stream, comma-separated with a single
// trailing newline. It is intended for simple, script-friendly tabular
// output.
func (p *Printer) Table(args ...any) {
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, arg)
	}
	b.WriteString("\n")
	out := b.String()

	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.data, out)
}

// Data prints s to the Data stream followed by a newline, for tools that
// emit a single blob of structured output (e.g. one JSON line) rather than
// row-oriented Table data.
func (p *Printer) Data(s string) {
	out := s + "\n"
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.data, out)
}

// formatLevel builds the complete line for a leveled message: args are
// joined exactly as fmt.Sprintln joins them (space-separated), then styled
// (symbol + color) when styling is enabled; otherwise the line is plain
// text, byte-identical to fmt.Fprintln(p.ui, args...). It does not touch
// p.ui or hold p.mu across the formatting work, so it's safe to call with
// arguments whose String/Error methods might themselves log through this
// same Printer.
func (p *Printer) formatLevel(level levelStyle, args ...any) string {
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	return p.formatMessage(level, msg)
}

// formatMessage is formatLevel split out for callers (ErrorDetail) that
// already have a formatted message rather than a Sprintln-style arg list.
func (p *Printer) formatMessage(level levelStyle, msg string) string {
	if !p.colorEnabled() {
		return msg + "\n"
	}
	return styleMessage(level, msg) + "\n"
}

// writeLevel formats and writes one leveled line to the UI stream. The
// message is fully built (including any Stringer/Error calls in args)
// before p.mu is ever taken, and the lock is held only around a single
// Write, so a String/Error method that logs through this same Printer
// cannot deadlock against it.
func (p *Printer) writeLevel(level levelStyle, args ...any) {
	out := p.formatLevel(level, args...)
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.ui, out)
}
