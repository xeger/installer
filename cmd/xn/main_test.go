package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xeger/installer/internal/platform"
)

func TestRun(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	tests := []struct {
		args     []string
		wantCode int
		wantOut  string
	}{
		{[]string{"help"}, 0, "Usage: " + platform.Name + " <tool>"},
		{[]string{"--help"}, 0, "No tools installed yet"},
		{[]string{"version"}, 0, platform.Name},
		{nil, 2, ""},
		{[]string{"--bogus"}, 2, ""},
		{[]string{"install"}, 2, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var out bytes.Buffer
			if code := run(context.Background(), tt.args, &out); code != tt.wantCode {
				t.Errorf("run(%q) = %d, want %d", tt.args, code, tt.wantCode)
			}
			if !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("run(%q) output %q should contain %q", tt.args, out.String(), tt.wantOut)
			}
		})
	}
}
