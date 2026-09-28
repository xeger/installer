// Package store manages the installed tools: where each binary lives and
// which release it came from.
//
// Layout, under the launcher's data directory (see platform.DataDir):
//
//	tools/<tool>/<tool>[.exe]
//	tools/<tool>/release.json
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"time"

	"github.com/xeger/installer/internal/platform"
)

const releaseFile = "release.json"

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// ValidateName reports whether name is acceptable as a tool name. Names
// come from the command line and become directory and repository names,
// so they are restricted to a conservative character set.
func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid tool name %q: use lowercase letters, digits, '.', '_' and '-'", name)
	}
	return nil
}

// ExeName returns the file name of tool's executable on this OS.
func ExeName(tool string) string {
	if runtime.GOOS == "windows" {
		return tool + ".exe"
	}
	return tool
}

// Dir returns the directory that holds tool's binary and release record.
func Dir(tool string) (string, error) {
	if err := ValidateName(tool); err != nil {
		return "", err
	}
	return platform.DataDir("tools", tool)
}

// Binary returns the path of tool's executable.
func Binary(tool string) (string, error) {
	dir, err := Dir(tool)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ExeName(tool)), nil
}

// Release records where an installed tool came from.
type Release struct {
	Tag        string    `json:"tag"`
	Repository string    `json:"repository"`
	CheckedAt  time.Time `json:"checked_at"` // last successful update check
}

// ReadRelease returns tool's release record. The error wraps fs.ErrNotExist
// when the tool has none.
func ReadRelease(tool string) (Release, error) {
	var r Release
	dir, err := Dir(tool)
	if err != nil {
		return r, err
	}
	data, err := os.ReadFile(filepath.Join(dir, releaseFile))
	if err != nil {
		return r, fmt.Errorf("read %s release: %w", tool, err)
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("parse %s release: %w", tool, err)
	}
	return r, nil
}

// WriteRelease saves tool's release record.
func WriteRelease(tool string, r Release) error {
	dir, err := Dir(tool)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s release: %w", tool, err)
	}
	if err := os.WriteFile(filepath.Join(dir, releaseFile), append(data, '\n'), 0o644); err != nil { //nolint:gosec // not secret
		return fmt.Errorf("write %s release: %w", tool, err)
	}
	return nil
}

// Tool is an installed tool. Tag is empty when its release is unknown.
type Tool struct {
	Name string
	Tag  string
}

// List returns the installed tools, sorted by name.
func List() ([]Tool, error) {
	root, err := platform.DataDir("tools")
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	var tools []Tool
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || ValidateName(name) != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, name, ExeName(name))); err != nil {
			continue
		}
		r, _ := ReadRelease(name)
		tools = append(tools, Tool{Name: name, Tag: r.Tag})
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools, nil
}
