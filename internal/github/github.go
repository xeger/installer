// Package github queries and downloads GitHub releases through the GitHub
// CLI (gh), which supplies authentication for private repositories.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/xeger/installer/internal/platform"
)

// Release is a GitHub release and its downloadable assets.
type Release struct {
	Tag    string  `json:"tagName"`
	Assets []Asset `json:"assets"`
}

// Asset is a file attached to a release.
type Asset struct {
	Name string `json:"name"`
}

// Has reports whether r has an asset named name.
func (r *Release) Has(name string) bool {
	for _, a := range r.Assets {
		if a.Name == name {
			return true
		}
	}
	return false
}

// AssetName returns the conventional release asset name for tool.
func AssetName(tool, tag, goos, goarch string) string {
	return fmt.Sprintf("%s_%s_%s_%s.tar.gz", tool, tag, goos, goarch)
}

// Check verifies that gh is installed and signed in to github.com. Its
// error explains how to fix the problem.
func Check(ctx context.Context) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("%s downloads tools with the GitHub CLI (gh), which is not installed.\n"+
			"Install it:\n  %s\nthen sign in:\n  gh auth login", platform.Name, installHint(runtime.GOOS))
	}
	if _, err := gh(ctx, "auth", "status", "--hostname", "github.com"); err != nil {
		return errors.New("the GitHub CLI (gh) is not signed in to github.com.\nSign in:\n  gh auth login")
	}
	return nil
}

func installHint(goos string) string {
	switch goos {
	case "darwin":
		return "brew install gh"
	case "windows":
		return "winget install --id GitHub.cli"
	default:
		return "see https://github.com/cli/cli#installation"
	}
}

// User returns the login of the account gh is signed in as.
func User(ctx context.Context) (string, error) {
	out, err := gh(ctx, "api", "user", "--jq", ".login")
	return strings.TrimSpace(string(out)), err
}

// Latest returns repo's latest release.
func Latest(ctx context.Context, repo string) (*Release, error) {
	out, err := gh(ctx, "release", "view", "--repo", repo, "--json", "tagName,assets")
	if err != nil {
		return nil, err
	}
	var r Release
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("parse %s release: %w", repo, err)
	}
	return &r, nil
}

// Download saves asset from repo's release tag into dir.
func Download(ctx context.Context, repo, tag, asset, dir string) error {
	_, err := gh(ctx, "release", "download", tag, "--repo", repo, "--pattern", asset, "--dir", dir)
	return err
}

// gh runs the GitHub CLI and returns its stdout. Errors include gh's
// stderr, which is where it explains what went wrong.
func gh(ctx context.Context, args ...string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, fmt.Errorf("gh %s: %s", args[0], msg)
		}
		return nil, fmt.Errorf("gh %s: %w", args[0], err)
	}
	return out, nil
}
