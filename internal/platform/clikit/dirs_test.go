package clikit

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestHomeFuncs_XDGAbsoluteOverride verifies that each XDG_*_HOME
// environment variable, when set to an absolute path, is honored as-is on
// every OS. This behavior does not depend on the platform default, so it
// is exercised here rather than in the OS-specific test files.
func TestHomeFuncs_XDGAbsoluteOverride(t *testing.T) {
	tests := []struct {
		name string
		env  string
		fn   func() (string, error)
	}{
		{"config", "XDG_CONFIG_HOME", ConfigHome},
		{"state", "XDG_STATE_HOME", StateHome},
		{"cache", "XDG_CACHE_HOME", CacheHome},
		{"data", "XDG_DATA_HOME", DataHome},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "xdg-"+tt.name)
			t.Setenv(tt.env, dir)

			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != dir {
				t.Errorf("got %q, want %q", got, dir)
			}
		})
	}
}

// TestApp_PathComposition covers App's path-joining rules: Vendor+Name+elem
// joined onto the resolved base, Vendor skipped when empty, and an empty
// Name treated as a programming error. It pins the base via an absolute
// XDG_CONFIG_HOME override so it needs no OS-specific default.
func TestApp_PathComposition(t *testing.T) {
	base := filepath.Join(t.TempDir(), "base")
	t.Setenv("XDG_CONFIG_HOME", base)

	t.Run("vendor and name", func(t *testing.T) {
		app := App{Vendor: "acme", Name: "cli"}
		got, err := app.ConfigDir("extra", "nested")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(base, "acme", "cli", "extra", "nested")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no elem", func(t *testing.T) {
		app := App{Vendor: "acme", Name: "cli"}
		got, err := app.ConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(base, "acme", "cli")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty vendor is skipped", func(t *testing.T) {
		app := App{Name: "foo"}
		got, err := app.ConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(base, "foo")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty name is an error", func(t *testing.T) {
		app := App{Vendor: "acme"}
		if _, err := app.ConfigDir(); err == nil {
			t.Error("expected error for empty Name, got nil")
		}
	})
}

// TestApp_AllDirKinds smoke-tests that each of StateDir, CacheDir and
// DataDir delegate to their respective *Home function and compose paths
// the same way as ConfigDir.
func TestApp_AllDirKinds(t *testing.T) {
	tests := []struct {
		name string
		env  string
		fn   func(App, ...string) (string, error)
	}{
		{"state", "XDG_STATE_HOME", App.StateDir},
		{"cache", "XDG_CACHE_HOME", App.CacheDir},
		{"data", "XDG_DATA_HOME", App.DataDir},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := filepath.Join(t.TempDir(), "base")
			t.Setenv(tt.env, base)

			app := App{Vendor: "acme", Name: "cli"}
			got, err := tt.fn(app, "sub")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := filepath.Join(base, "acme", "cli", "sub")
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// TestApp_RejectsEscapes guards against Vendor, Name, or elem (which often
// carries user input, e.g. `<launcher> install <name>`) resolving outside the app's
// directory.
func TestApp_RejectsEscapes(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	tests := []struct {
		name string
		app  App
		elem []string
	}{
		{"empty name", App{Vendor: "acme"}, nil},
		{"dot-dot name", App{Vendor: "acme", Name: ".."}, nil},
		{"dot name", App{Vendor: "acme", Name: "."}, nil},
		{"name with slash", App{Vendor: "acme", Name: "a/b"}, nil},
		{"name with backslash", App{Vendor: "acme", Name: `a\b`}, nil},
		{"dot-dot vendor", App{Vendor: "../..", Name: "cli"}, nil},
		{"vendor with slash", App{Vendor: "a/b", Name: "cli"}, nil},
		{"elem escapes", App{Vendor: "acme", Name: "cli"}, []string{"cmd", "../../../../etc"}},
		{"elem is parent", App{Vendor: "acme", Name: "cli"}, []string{".."}},
		{"elem absolute", App{Vendor: "acme", Name: "cli"}, []string{"/etc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.app.DataDir(tt.elem...)
			if !errors.Is(err, ErrInvalidAppPath) {
				t.Errorf("DataDir(%q) = %q, %v; want ErrInvalidAppPath", tt.elem, got, err)
			}
		})
	}
}

// TestApp_AllowsInnerDotDot checks that elem may use ".." as long as the
// result stays inside the app's directory.
func TestApp_AllowsInnerDotDot(t *testing.T) {
	base := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_DATA_HOME", base)

	got, err := App{Vendor: "acme", Name: "cli"}.DataDir("cmd", "..", "bin")
	if err != nil {
		t.Fatalf("DataDir() error = %v", err)
	}
	if want := filepath.Join(base, "acme", "cli", "bin"); got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

func TestHomeFuncs_XDGCleaned(t *testing.T) {
	base := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_DATA_HOME", base+string(filepath.Separator))

	got, err := DataHome()
	if err != nil {
		t.Fatalf("DataHome() error = %v", err)
	}
	if got != base {
		t.Errorf("DataHome() = %q, want cleaned %q", got, base)
	}
}
