// Package ui prints the launcher's status messages: short, leveled lines on
// stderr for people, never data for scripts. When color is enabled each
// line starts with a colored symbol; otherwise output is plain text, byte
// for byte what a script or log would expect.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/xeger/installer/internal/platform"
)

// level describes how one kind of message is decorated when color is on.
type level struct {
	symbol string // single-width glyph that starts the line
	sgr    string // SGR parameters for the symbol
	text   string // SGR parameters for the message text; "" leaves it unstyled
}

// Info and success text stays at the terminal's default color so it reads
// well on any background; warnings and errors are colored to draw the eye.
var (
	infoLevel    = level{symbol: "›", sgr: "1;34"}
	successLevel = level{symbol: "✓", sgr: "1;32"}
	warnLevel    = level{symbol: "!", sgr: "1;33", text: "33"}
	errorLevel   = level{symbol: "✗", sgr: "1;31", text: "31"}
)

// Printer writes status messages to a single stream. It is safe for
// concurrent use; each message is written whole with one Write.
type Printer struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

// New returns a Printer that writes to w, decorating messages when color
// is true.
func New(w io.Writer, color bool) *Printer {
	return &Printer{w: w, color: color}
}

// Default writes to stderr, with color decided by platform.ColorEnabled.
// Tests may replace it.
var Default = New(os.Stderr, platform.ColorEnabled(os.Stderr))

// Info reports progress or a neutral fact.
func (p *Printer) Info(args ...any) { p.write(p.format(infoLevel, args...)) }

// Success reports that something finished.
func (p *Printer) Success(args ...any) { p.write(p.format(successLevel, args...)) }

// Warn reports a problem the launcher worked around.
func (p *Printer) Warn(args ...any) { p.write(p.format(warnLevel, args...)) }

// Error reports a failure.
func (p *Printer) Error(args ...any) { p.write(p.format(errorLevel, args...)) }

// ErrorDetail reports a failure: an Error line built from prefaceContext
// (omitted when empty), then err as an indented tree. Errors that wrap
// several others (errors.Join, fmt.Errorf with multiple %w) list each
// cause beneath the parent's own message.
func (p *Printer) ErrorDetail(err error, prefaceContext ...any) {
	var b strings.Builder
	switch {
	case err == nil && len(prefaceContext) == 0:
		b.WriteString("<nil>\n")
	case err == nil:
		b.WriteString(p.format(errorLevel, prefaceContext...))
	default:
		if len(prefaceContext) > 0 {
			b.WriteString(p.format(errorLevel, prefaceContext...))
		}
		explain(&b, 1, err)
	}
	p.write(b.String())
}

// format renders one message, joining args like fmt.Sprintln. Formatting
// happens before any lock is taken, so a Stringer or error that itself
// logs through the same Printer cannot deadlock.
func (p *Printer) format(l level, args ...any) string {
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	if !p.color {
		return msg + "\n"
	}
	return decorate(l, msg) + "\n"
}

func (p *Printer) write(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, _ = io.WriteString(p.w, s)
}

// decorate styles msg for l: the colored symbol starts the first line and
// continuation lines are indented to align under the text. Each line is
// styled separately, so no escape sequence spans a newline.
func decorate(l level, msg string) string {
	var b strings.Builder
	for i, line := range strings.Split(msg, "\n") {
		switch {
		case i == 0:
			b.WriteString(sgr(l.sgr, l.symbol))
			if line != "" {
				b.WriteString(" " + sgr(l.text, line))
			}
		case line == "":
			b.WriteString("\n")
		default:
			b.WriteString("\n  " + sgr(l.text, line))
		}
	}
	return b.String()
}

// sgr wraps s in an ANSI Select Graphic Rendition sequence; empty params
// leave s unstyled.
func sgr(params, s string) string {
	if params == "" {
		return s
	}
	return "\x1b[" + params + "m" + s + "\x1b[m"
}

// multiError is implemented by errors that wrap several others.
type multiError interface{ Unwrap() []error }

// explain writes err as a tree indented by nesting levels.
func explain(b *strings.Builder, nesting int, err error) {
	indent := strings.Repeat("  ", nesting)
	m, ok := err.(multiError) //nolint:errorlint // inspecting this error's own shape, not its chain
	if !ok {
		leaf(b, indent, err)
		return
	}
	children := m.Unwrap()
	switch len(children) {
	case 0:
		leaf(b, indent, err)
		return
	case 1:
		explain(b, nesting, children[0])
		return
	}
	// errors.Join's message is just its children's; anything else (such as
	// "syncing foo: a; b" from fmt.Errorf) carries context worth keeping.
	if err.Error() == joined(children) {
		b.WriteString(indent + "multiple errors:\n")
	} else {
		leaf(b, indent, err)
	}
	for _, c := range children {
		explain(b, nesting+1, c)
	}
}

func leaf(b *strings.Builder, indent string, err error) {
	b.WriteString(indent + strings.ReplaceAll(err.Error(), "\n", "\n"+indent) + "\n")
}

func joined(errs []error) string {
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "\n")
}
