package ui

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPlainLevels(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false)
	p.Info("info", 1)
	p.Success("done")
	p.Warn("careful")
	p.Error("broke")
	p.Error()
	if got, want := buf.String(), "info 1\ndone\ncareful\nbroke\n\n"; got != want {
		t.Errorf("plain output = %q, want %q", got, want)
	}
}

func TestPlainMultilineIsUntouched(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, false).Info("first\n\tsecond")
	if got, want := buf.String(), "first\n\tsecond\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyledBytes(t *testing.T) {
	tests := []struct {
		name  string
		print func(*Printer)
		want  string
	}{
		{"info", func(p *Printer) { p.Info("hello") }, "\x1b[1;34m›\x1b[m hello\n"},
		{"success", func(p *Printer) { p.Success("done") }, "\x1b[1;32m✓\x1b[m done\n"},
		{"warn", func(p *Printer) { p.Warn("careful") }, "\x1b[1;33m!\x1b[m \x1b[33mcareful\x1b[m\n"},
		{"error", func(p *Printer) { p.Error("broke") }, "\x1b[1;31m✗\x1b[m \x1b[31mbroke\x1b[m\n"},
		{"no args has no trailing space", func(p *Printer) { p.Error() }, "\x1b[1;31m✗\x1b[m\n"},
		{
			"multi-line indents continuations, keeps tabs, no trailing space",
			func(p *Printer) { p.Warn("one\n\ttwo\n\nfour") },
			"\x1b[1;33m!\x1b[m \x1b[33mone\x1b[m\n  \x1b[33m\ttwo\x1b[m\n\n  \x1b[33mfour\x1b[m\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.print(New(&buf, true))
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorDetail(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		preface []any
		want    string
	}{
		{"nil without preface", nil, nil, "<nil>\n"},
		{"nil with preface", nil, []any{"Cannot run", "foo"}, "Cannot run foo\n"},
		{"leaf without preface", errors.New("boom"), nil, "  boom\n"},
		{"multi-line leaf is indented", errors.New("a\nb"), []any{"Failed"}, "Failed\n  a\n  b\n"},
		{
			"single wrap is followed", fmt.Errorf("ctx: %w", errors.New("x")), []any{"Failed"},
			"Failed\n  ctx: x\n",
		},
		{
			"errors.Join uses a generic label", errors.Join(errors.New("a"), errors.New("b")), []any{"Failed"},
			"Failed\n  multiple errors:\n    a\n    b\n",
		},
		{
			"multiple %w keeps its own context",
			fmt.Errorf("syncing site foo: %w; %w", errors.New("disk full"), errors.New("network unreachable")),
			[]any{"Failed"},
			"Failed\n  syncing site foo: disk full; network unreachable\n    disk full\n    network unreachable\n",
		},
		{"empty Unwrap is a leaf", emptyMulti{}, nil, "  empty\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			New(&buf, false).ErrorDetail(tt.err, tt.preface...)
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

type emptyMulti struct{}

func (emptyMulti) Error() string   { return "empty" }
func (emptyMulti) Unwrap() []error { return nil }

// loopback logs through its Printer when formatted, which deadlocks if the
// Printer holds its lock while formatting.
type loopback struct{ p *Printer }

func (l loopback) Error() string  { l.p.Info("formatting"); return "loopback" }
func (l loopback) String() string { l.p.Info("formatting"); return "loopback" }

func TestReentrantFormattingDoesNotDeadlock(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false)
	done := make(chan struct{})
	go func() {
		p.ErrorDetail(loopback{p}, "Failed")
		p.Info(loopback{p})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: formatting logged through the same Printer")
	}
}

func TestConcurrentUseWritesWholeLines(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, true)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Info("line", i)
			p.ErrorDetail(errors.Join(errors.New("a"), errors.New("b")), "join", i)
		}()
	}
	wg.Wait()
	if n := strings.Count(buf.String(), "multiple errors:"); n != 20 {
		t.Errorf("found %d error trees, want 20", n)
	}
}
