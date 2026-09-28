//go:build !windows

package clikit

import (
	"path/filepath"
	"testing"
)

// TestUnixDefaults covers the Unix (including macOS) default base
// directories, both when the XDG_*_HOME variable is unset and when it is
// set to a relative (and therefore invalid, per the XDG spec) path.
func TestUnixDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		env  string
		fn   func() (string, error)
		rel  []string
	}{
		{"config", "XDG_CONFIG_HOME", ConfigHome, []string{".config"}},
		{"state", "XDG_STATE_HOME", StateHome, []string{".local", "state"}},
		{"cache", "XDG_CACHE_HOME", CacheHome, []string{".cache"}},
		{"data", "XDG_DATA_HOME", DataHome, []string{".local", "share"}},
	}

	for _, tt := range tests {
		want := filepath.Join(append([]string{home}, tt.rel...)...)

		t.Run(tt.name+"/unset", func(t *testing.T) {
			t.Setenv(tt.env, "")
			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})

		t.Run(tt.name+"/relative is ignored", func(t *testing.T) {
			t.Setenv(tt.env, "relative/path")
			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// TestUnixHomeMissing verifies that when both HOME and the relevant XDG
// variable are unset, the *Home functions return a wrapped error rather
// than an empty string.
func TestUnixHomeMissing(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := ConfigHome()
	if err == nil {
		t.Fatalf("expected error when HOME and XDG_CONFIG_HOME are both unset, got %q", got)
	}
}

// TestUnixRelativeHome verifies that a relative $HOME is rejected rather
// than yielding paths relative to the working directory.
func TestUnixRelativeHome(t *testing.T) {
	t.Setenv("HOME", "rel")
	t.Setenv("XDG_DATA_HOME", "")

	if got, err := DataHome(); err == nil {
		t.Fatalf("DataHome() = %q, want error for relative HOME", got)
	}
}
