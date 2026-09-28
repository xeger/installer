package term

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func newPlain(verbose bool) (*Printer, *bytes.Buffer, *bytes.Buffer) {
	var data, ui bytes.Buffer
	p := New(Options{
		Data:    &data,
		UI:      &ui,
		Verbose: verbose,
		Color:   new(false),
	})
	return p, &data, &ui
}

func TestPlainLevels(t *testing.T) {
	tests := []struct {
		name string
		call func(p *Printer)
		want string
	}{
		{"Verbose", func(p *Printer) { p.SetVerbose(true); p.Verbose("digging", "in") }, "digging in\n"},
		{"Info", func(p *Printer) { p.Info("starting", "up") }, "starting up\n"},
		{"Success", func(p *Printer) { p.Success("all", "good") }, "all good\n"},
		{"Warn", func(p *Printer) { p.Warn("careful", "now") }, "careful now\n"},
		{"Error", func(p *Printer) { p.Error("broke", "badly") }, "broke badly\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _, ui := newPlain(false)
			tt.call(p)
			if got := ui.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVerboseSuppressedByDefault(t *testing.T) {
	p, _, ui := newPlain(false)
	p.Verbose("should", "not", "appear")
	if got := ui.String(); got != "" {
		t.Errorf("expected no output while verbose is disabled, got %q", got)
	}
	if p.Verbosity() {
		t.Error("Verbosity() = true, want false")
	}

	p.SetVerbose(true)
	if !p.Verbosity() {
		t.Error("Verbosity() = false after SetVerbose(true)")
	}
	p.Verbose("now", "visible")
	if got, want := ui.String(), "now visible\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTableAndData(t *testing.T) {
	p, data, ui := newPlain(false)

	p.Table("a", 1, "b", 2)
	if got, want := data.String(), "a, 1, b, 2\n"; got != want {
		t.Errorf("Table: got %q, want %q", got, want)
	}
	if ui.Len() != 0 {
		t.Errorf("Table wrote to UI stream: %q", ui.String())
	}

	data.Reset()
	p.Data(`{"k":"v"}`)
	if got, want := data.String(), "{\"k\":\"v\"}\n"; got != want {
		t.Errorf("Data: got %q, want %q", got, want)
	}
}

func TestErrorDetailPlainTree(t *testing.T) {
	p, data, ui := newPlain(false)

	leaf1 := errors.New("first line\nsecond line")
	leaf2 := errors.New("boom")
	joined := errors.Join(leaf1, leaf2)
	wrapped := errors.Join(joined) // single-error Unwrap: should pass through transparently

	p.ErrorDetail(wrapped, "operation failed:", "sync")

	want := "operation failed: sync\n" +
		"  multiple errors:\n" +
		"    first line\n" +
		"    second line\n" +
		"    boom\n"
	if got := ui.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if data.Len() != 0 {
		t.Errorf("ErrorDetail wrote to Data stream: %q", data.String())
	}
}

func TestErrorDetailSingleWrap(t *testing.T) {
	p, _, ui := newPlain(false)
	inner := errors.New("root cause")
	wrapped := errors.Join(inner) // Unwrap() []error of length 1
	p.ErrorDetail(wrapped)

	// No prefaceContext was passed, so no header line is printed (see
	// finding 8) -- just the tree, at nesting 1.
	want := "  root cause\n"
	if got := ui.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorDetailOwnMessagePreserved(t *testing.T) {
	p, _, ui := newPlain(false)

	a := errors.New("disk full")
	b := errors.New("network unreachable")
	wrapped := fmt.Errorf("syncing site foo: %w; %w", a, b)

	p.ErrorDetail(wrapped)

	want := "  syncing site foo: disk full; network unreachable\n" +
		"    disk full\n" +
		"    network unreachable\n"
	if got := ui.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestErrorDetailJoinStillUsesGenericLabel(t *testing.T) {
	p, _, ui := newPlain(false)

	a := errors.New("disk full")
	b := errors.New("network unreachable")
	wrapped := errors.Join(a, b)

	p.ErrorDetail(wrapped)

	want := "  multiple errors:\n" +
		"    disk full\n" +
		"    network unreachable\n"
	if got := ui.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestErrorDetailEmptyUnwrapIsLeaf(t *testing.T) {
	p, _, ui := newPlain(false)
	p.ErrorDetail(emptyUnwrapMultiError{}, "context:")

	want := "context:\n" +
		"  empty multi\n"
	if got := ui.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

type emptyUnwrapMultiError struct{}

func (emptyUnwrapMultiError) Error() string   { return "empty multi" }
func (emptyUnwrapMultiError) Unwrap() []error { return nil }

func TestErrorDetailNil(t *testing.T) {
	t.Run("no preface", func(t *testing.T) {
		p, _, ui := newPlain(false)
		p.ErrorDetail(nil)
		if got, want := ui.String(), "<nil>\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with preface", func(t *testing.T) {
		p, _, ui := newPlain(false)
		p.ErrorDetail(nil, "nothing to report:")
		if got, want := ui.String(), "nothing to report:\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestColorForcedOnEmitsEscapesAndSymbol(t *testing.T) {
	var ui bytes.Buffer
	p := New(Options{UI: &ui, Color: new(true)})
	p.Info("hello")
	got := ui.String()
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape sequence in colored output, got %q", got)
	}
	if !strings.Contains(got, infoSymbol) {
		t.Errorf("expected info symbol in colored output, got %q", got)
	}
}

func TestColorOffEmitsNoEscapesOrSymbol(t *testing.T) {
	var ui bytes.Buffer
	p := New(Options{UI: &ui, Color: new(false)})
	p.Info("hello")
	got := ui.String()
	if strings.Contains(got, "\x1b[") {
		t.Errorf("expected no ANSI escapes, got %q", got)
	}
	if strings.Contains(got, infoSymbol) {
		t.Errorf("expected no symbol, got %q", got)
	}
	if got != "hello\n" {
		t.Errorf("got %q, want %q", got, "hello\n")
	}
}

func TestAutoColorRespectsNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if got := autoColor(w); got {
		t.Error("autoColor() = true with NO_COLOR set, want false")
	}

	// Color left nil: auto-detection should also disable a Printer built
	// against a real file-like UI stream (not just a bytes.Buffer, which
	// would already be plain regardless of NO_COLOR).
	p := New(Options{UI: w})
	p.Info("hello")
	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "hello\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewDefaults(t *testing.T) {
	p := New(Options{})
	if p.data == nil || p.ui == nil {
		t.Fatal("New(Options{}) left data or ui nil")
	}
}

// loggingError's Error method logs through the same Printer that's
// printing it. Before the deadlock fix, ErrorDetail/writeLevel held p.mu
// while calling err.Error() (or while fmt.Sprintln invoked a Stringer),
// so this would deadlock.
type loggingError struct {
	p   *Printer
	msg string
}

func (e loggingError) Error() string {
	e.p.Verbose("explaining:", e.msg)
	return e.msg
}

func TestErrorMethodLoggingSameLevelDoesNotDeadlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		p, _, _ := newPlain(true)
		err := loggingError{p: p, msg: "boom"}
		p.ErrorDetail(err, "operation failed")
		p.Error(err)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out: likely deadlock in ErrorDetail/writeLevel while holding p.mu")
	}
}

// loggingStringer's String method logs through the Printer that's about to
// render it via fmt.Sprintln.
type loggingStringer struct {
	p *Printer
}

func (s loggingStringer) String() string {
	s.p.Info("formatting a value")
	return "stringer output"
}

func TestStringerLoggingSamePrinterDoesNotDeadlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		p, _, _ := newPlain(false)
		p.Info("value:", loggingStringer{p: p})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out: likely deadlock formatting a Stringer while holding p.mu")
	}
}

func TestStyledMultilineMessage(t *testing.T) {
	var ui bytes.Buffer
	p := New(Options{UI: &ui, Color: new(true)})
	p.Info("first line\n\tsecond line")

	got := ui.String()

	symbol := infoLevel.symbolStyle.Render(infoSymbol)
	line1 := infoLevel.textStyle.TabWidth(-1).Render("first line")
	line2 := infoLevel.textStyle.TabWidth(-1).Render("\tsecond line")
	want := symbol + " " + line1 + "\n" + "  " + line2 + "\n"

	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}

	// No line may carry trailing whitespace.
	for i, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if strings.HasSuffix(line, " ") {
			t.Errorf("line %d has trailing whitespace: %q", i, line)
		}
	}

	// The tab must survive untouched (no conversion to spaces).
	if !strings.Contains(got, "\tsecond line") {
		t.Errorf("expected literal tab in continuation line, got %q", got)
	}
}

func TestErrorWithNoArgsStyledHasNoTrailingSpace(t *testing.T) {
	var ui bytes.Buffer
	p := New(Options{UI: &ui, Color: new(true)})
	p.Error()

	got := ui.String()
	want := errorLevel.symbolStyle.Render(errorSymbol) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if strings.Contains(got, " \n") {
		t.Errorf("expected no trailing space before newline, got %q", got)
	}
}

func TestConcurrentUse(t *testing.T) {
	p, _, _ := newPlain(true)

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p.Info("info", i)
			p.Verbose("verbose", i)
			p.Success("success", i)
			p.Warn("warn", i)
			p.Error("error", i)
			p.Table("row", i)
			p.Data(fmt.Sprintf("data %d", i))
			p.ErrorDetail(fmt.Errorf("err %d", i), "context", i)
			p.SetVerbose(i%2 == 0)
			p.SetColor(i%2 == 0)
			_ = p.Verbosity()
			_ = p.colorEnabled()
		}(i)
	}
	wg.Wait()
}
