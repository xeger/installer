package clikit

import (
	"os"
	"testing"
)

// pipeWriter returns the write end of a pipe: an *os.File that is never a
// terminal.
func pipeWriter(t *testing.T) *os.File {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return w
}

func TestColorEnabled(t *testing.T) {
	tests := []struct {
		name    string
		noColor string
		force   string
		term    string
		want    bool
	}{
		{name: "pipe is not a terminal"},
		{name: "CLICOLOR_FORCE forces a pipe", force: "1", want: true},
		{name: "CLICOLOR_FORCE=0 does not force", force: "0"},
		{name: "NO_COLOR wins over CLICOLOR_FORCE", noColor: "1", force: "1"},
		{name: "any non-empty NO_COLOR disables", noColor: "yes", force: "1"},
		{name: "CLICOLOR_FORCE overrides TERM=dumb", force: "1", term: "dumb", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("CLICOLOR_FORCE", tt.force)
			t.Setenv("TERM", tt.term)
			if got := ColorEnabled(pipeWriter(t)); got != tt.want {
				t.Errorf("ColorEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorEnabledNilFile(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	if ColorEnabled(nil) {
		t.Error("ColorEnabled(nil) = true, want false")
	}
	t.Setenv("CLICOLOR_FORCE", "1")
	if !ColorEnabled(nil) {
		t.Error("ColorEnabled(nil) with CLICOLOR_FORCE = false, want true")
	}
}
