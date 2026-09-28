//go:build windows

package clikit

import (
	"path/filepath"
	"testing"
)

// TestWindowsDefaults covers the Windows default base directories: the
// platform environment variable (APPDATA/LOCALAPPDATA) when set to an
// absolute path, and the AppData fallback under the home directory when
// it is unset or relative.
func TestWindowsDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)

	tests := []struct {
		name   string
		xdgEnv string
		winEnv string
		fn     func() (string, error)
		rel    []string
	}{
		{"config", "XDG_CONFIG_HOME", "APPDATA", ConfigHome, []string{"AppData", "Roaming"}},
		{"data", "XDG_DATA_HOME", "APPDATA", DataHome, []string{"AppData", "Roaming"}},
		{"state", "XDG_STATE_HOME", "LOCALAPPDATA", StateHome, []string{"AppData", "Local"}},
		{"cache", "XDG_CACHE_HOME", "LOCALAPPDATA", CacheHome, []string{"AppData", "Local"}},
	}

	for _, tt := range tests {
		fallback := filepath.Join(append([]string{home}, tt.rel...)...)

		t.Run(tt.name+"/platform env set", func(t *testing.T) {
			t.Setenv(tt.xdgEnv, "")
			winDir := filepath.Join(t.TempDir(), "platform-"+tt.name)
			t.Setenv(tt.winEnv, winDir)

			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != winDir {
				t.Errorf("got %q, want %q", got, winDir)
			}
		})

		t.Run(tt.name+"/platform env relative is ignored", func(t *testing.T) {
			t.Setenv(tt.xdgEnv, "")
			t.Setenv(tt.winEnv, `relative\path`)

			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != fallback {
				t.Errorf("got %q, want %q", got, fallback)
			}
		})

		t.Run(tt.name+"/platform env unset falls back to home", func(t *testing.T) {
			t.Setenv(tt.xdgEnv, "")
			t.Setenv(tt.winEnv, "")

			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != fallback {
				t.Errorf("got %q, want %q", got, fallback)
			}
		})

		t.Run(tt.name+"/XDG absolute override wins over platform env", func(t *testing.T) {
			xdgDir := filepath.Join(t.TempDir(), "xdg-"+tt.name)
			t.Setenv(tt.xdgEnv, xdgDir)
			t.Setenv(tt.winEnv, filepath.Join(t.TempDir(), "platform-"+tt.name))

			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != xdgDir {
				t.Errorf("got %q, want %q", got, xdgDir)
			}
		})
	}
}

// TestWindowsHomeMissing verifies that when no home-directory source and
// no XDG or platform environment variable is available, the *Home
// functions return a wrapped error rather than an empty string.
func TestWindowsHomeMissing(t *testing.T) {
	t.Setenv("USERPROFILE", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("APPDATA", "")

	got, err := ConfigHome()
	if err == nil {
		t.Fatalf("expected error when no home directory source is available, got %q", got)
	}
}

// TestWindowsXDGAbsoluteness pins filepath.IsAbs semantics on Windows: only
// values with a volume and a rooted path count as absolute; anything else
// is ignored and the platform default applies.
func TestWindowsXDGAbsoluteness(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", "")
	fallback := filepath.Join(home, "AppData", "Roaming")

	tests := []struct {
		xdg  string
		want string
	}{
		{`C:\xdg`, `C:\xdg`},
		{`C:/xdg`, `C:\xdg`},
		{`\\server\share\xdg`, `\\server\share\xdg`},
		{`C:xdg`, fallback},
		{`\xdg`, fallback},
		{`xdg`, fallback},
	}

	for _, tt := range tests {
		t.Run(tt.xdg, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.xdg)
			got, err := ConfigHome()
			if err != nil {
				t.Fatalf("ConfigHome() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ConfigHome() = %q, want %q", got, tt.want)
			}
		})
	}
}
