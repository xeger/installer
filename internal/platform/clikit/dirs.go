package clikit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidAppPath is returned (wrapped) by App's directory methods when
// Vendor, Name, or elem would produce a path outside the app's directory,
// e.g. an empty Name, a Name containing a path separator, or ".." in elem.
var ErrInvalidAppPath = errors.New("invalid app path")

// dirKind describes one of the four XDG base directory categories: the
// environment variable that overrides it on every OS, plus enough
// OS-specific metadata (consulted only by the platform's own
// xdgDefaultBase, in dirs_unix.go / dirs_windows.go) to compute its
// default location when that variable is unset.
type dirKind struct {
	// name is the human-readable kind ("config", "state", ...) for errors.
	name string
	// xdgEnv is the XDG_*_HOME environment variable name for this kind.
	// Per the XDG Base Directory Specification, it is honored on every OS
	// -- including macOS and Windows -- but only when set to an absolute
	// path; a relative value is treated as unset.
	xdgEnv string
	// unixRel is the path, relative to the user's home directory, used as
	// the Unix default (this includes macOS: we deliberately use the XDG
	// layout there too, unlike os.UserConfigDir).
	unixRel []string
	// winEnv is the Windows environment variable (APPDATA or
	// LOCALAPPDATA) consulted as the platform default.
	winEnv string
	// winRel is the fallback path, relative to the user's home directory,
	// used when winEnv is unset or not absolute.
	winRel []string
}

var (
	dirConfigKind = dirKind{
		name:    "config",
		xdgEnv:  "XDG_CONFIG_HOME",
		unixRel: []string{".config"},
		winEnv:  "APPDATA",
		winRel:  []string{"AppData", "Roaming"},
	}
	dirStateKind = dirKind{
		name:    "state",
		xdgEnv:  "XDG_STATE_HOME",
		unixRel: []string{".local", "state"},
		winEnv:  "LOCALAPPDATA",
		winRel:  []string{"AppData", "Local"},
	}
	dirCacheKind = dirKind{
		name:    "cache",
		xdgEnv:  "XDG_CACHE_HOME",
		unixRel: []string{".cache"},
		winEnv:  "LOCALAPPDATA",
		winRel:  []string{"AppData", "Local"},
	}
	dirDataKind = dirKind{
		name:    "data",
		xdgEnv:  "XDG_DATA_HOME",
		unixRel: []string{".local", "share"},
		winEnv:  "APPDATA",
		winRel:  []string{"AppData", "Roaming"},
	}
)

// dirResolve is the single resolver shared by ConfigHome, StateHome,
// CacheHome and DataHome: if kind's XDG environment variable is set to an
// absolute path, that value (cleaned) wins on every OS; otherwise it falls
// back to the platform default from xdgDefaultBase.
func dirResolve(kind dirKind) (string, error) {
	if v := os.Getenv(kind.xdgEnv); v != "" && filepath.IsAbs(v) {
		return filepath.Clean(v), nil
	}
	base, err := xdgDefaultBase(kind)
	if err != nil {
		return "", fmt.Errorf("clikit: resolve %s directory: %w", kind.name, err)
	}
	return base, nil
}

// dirUserHome returns the user's home directory, which must be absolute:
// a relative $HOME would silently yield paths relative to the working
// directory.
func dirUserHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("home directory %q is not an absolute path", home)
	}
	return home, nil
}

// ConfigHome returns the base directory for per-user configuration files.
//
// If XDG_CONFIG_HOME is set to an absolute path, it is used as-is on every
// OS, including macOS and Windows; a relative value is ignored, per the
// XDG Base Directory Specification. Otherwise the OS default applies:
// ~/.config on Unix and macOS (we deliberately use the XDG layout on macOS
// rather than ~/Library, unlike os.UserConfigDir), or %AppData% (falling
// back to <home>\AppData\Roaming) on Windows.
func ConfigHome() (string, error) {
	return dirResolve(dirConfigKind)
}

// StateHome returns the base directory for per-user state that should
// persist between runs but is not as important as data (logs, history,
// recently-used files, and similar).
//
// If XDG_STATE_HOME is set to an absolute path, it is used as-is on every
// OS, including macOS and Windows; a relative value is ignored, per the
// XDG Base Directory Specification. Otherwise the OS default applies:
// ~/.local/state on Unix and macOS (deliberately the XDG layout, not
// ~/Library), or %LocalAppData% (falling back to
// <home>\AppData\Local) on Windows.
func StateHome() (string, error) {
	return dirResolve(dirStateKind)
}

// CacheHome returns the base directory for per-user, non-essential cached
// data that can be safely deleted or regenerated.
//
// If XDG_CACHE_HOME is set to an absolute path, it is used as-is on every
// OS, including macOS and Windows; a relative value is ignored, per the
// XDG Base Directory Specification. Otherwise the OS default applies:
// ~/.cache on Unix and macOS (deliberately the XDG layout, not
// ~/Library), or %LocalAppData% (falling back to
// <home>\AppData\Local) on Windows.
func CacheHome() (string, error) {
	return dirResolve(dirCacheKind)
}

// DataHome returns the base directory for per-user application data that
// should persist (installed assets, databases, and similar).
//
// If XDG_DATA_HOME is set to an absolute path, it is used as-is on every
// OS, including macOS and Windows; a relative value is ignored, per the
// XDG Base Directory Specification. Otherwise the OS default applies:
// ~/.local/share on Unix and macOS (deliberately the XDG layout, not
// ~/Library), or %AppData% (falling back to <home>\AppData\Roaming) on
// Windows.
func DataHome() (string, error) {
	return dirResolve(dirDataKind)
}

// App identifies a CLI tool for the purpose of resolving its per-user
// directories, laid out as <base>/<Vendor>/<Name>.
type App struct {
	// Vendor is an optional namespacing segment, e.g. "acme". When
	// empty, it is omitted from the resulting path. When set, it must be a
	// single path component.
	Vendor string
	// Name is the tool's directory name, e.g. "cli". It must be a single,
	// non-empty path component.
	Name string
}

// path joins base (resolved by baseHome) with Vendor (if set), Name, and
// elem, as filepath.Join(base, [Vendor,] Name, elem...). It performs path
// resolution only -- it never creates directories. It is the shared
// implementation behind ConfigDir, StateDir, CacheDir and DataDir.
//
// Because elem often carries user input (e.g. a command name), the result
// is guaranteed to stay inside the app's directory.
func (a App) path(baseHome func() (string, error), elem ...string) (string, error) {
	if !dirIsComponent(a.Name) {
		return "", fmt.Errorf("clikit: App.Name %q: %w", a.Name, ErrInvalidAppPath)
	}
	if a.Vendor != "" && !dirIsComponent(a.Vendor) {
		return "", fmt.Errorf("clikit: App.Vendor %q: %w", a.Vendor, ErrInvalidAppPath)
	}
	if len(elem) > 0 {
		if rel := filepath.Join(elem...); !filepath.IsLocal(rel) {
			return "", fmt.Errorf("clikit: path %q escapes the app directory: %w", rel, ErrInvalidAppPath)
		}
	}
	base, err := baseHome()
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(elem)+3)
	parts = append(parts, base)
	if a.Vendor != "" {
		parts = append(parts, a.Vendor)
	}
	parts = append(parts, a.Name)
	parts = append(parts, elem...)
	return filepath.Join(parts...), nil
}

// dirIsComponent reports whether s is exactly one ordinary path component:
// non-empty, not "." or "..", and free of path separators.
func dirIsComponent(s string) bool {
	return s != "" && s != "." && s != ".." && !strings.ContainsAny(s, `/\`) && filepath.IsLocal(s)
}

// ConfigDir returns App's configuration directory, optionally joined with
// elem, under ConfigHome. It returns an error wrapping ErrInvalidAppPath if
// the App or elem is invalid (see App and ErrInvalidAppPath).
func (a App) ConfigDir(elem ...string) (string, error) {
	return a.path(ConfigHome, elem...)
}

// StateDir returns App's state directory, optionally joined with elem,
// under StateHome.
func (a App) StateDir(elem ...string) (string, error) {
	return a.path(StateHome, elem...)
}

// CacheDir returns App's cache directory, optionally joined with elem,
// under CacheHome.
func (a App) CacheDir(elem ...string) (string, error) {
	return a.path(CacheHome, elem...)
}

// DataDir returns App's data directory, optionally joined with elem,
// under DataHome.
func (a App) DataDir(elem ...string) (string, error) {
	return a.path(DataHome, elem...)
}
