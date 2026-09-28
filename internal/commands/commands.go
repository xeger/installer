// Package commands implements the launcher's built-in behaviors:
// installing, upgrading, and keeping tools current.
package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/xeger/installer/internal/archive"
	"github.com/xeger/installer/internal/github"
	"github.com/xeger/installer/internal/platform"
	"github.com/xeger/installer/internal/store"
	"github.com/xeger/installer/internal/ui"
	"github.com/xeger/installer/internal/updates"
)

var now = time.Now

var validOwner = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)

// parseTool splits a tool argument into the tool's name and, when the
// argument names a repository ("owner/repo"), that repository. A tool from
// an explicit repository is named after the repository.
func parseTool(arg string) (tool, repo string, err error) {
	owner, name, explicit := strings.Cut(arg, "/")
	if !explicit {
		if err := store.ValidateName(arg); err != nil {
			return "", "", err
		}
		return arg, "", nil
	}
	if !validOwner.MatchString(owner) || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid repository %q: use <owner>/<tool>", arg)
	}
	if err := store.ValidateName(name); err != nil {
		return "", "", err
	}
	return name, arg, nil
}

// Ensure makes the tool named by arg (a tool name or "owner/repo") ready to
// run and returns its binary: it installs the tool if missing and, at most
// once per updates.Interval, upgrades it. A failed update check only warns,
// so an installed tool still runs offline.
func Ensure(ctx context.Context, arg string) (string, error) {
	tool, repo, err := parseTool(arg)
	if err != nil {
		return "", err
	}
	bin, err := store.Binary(tool)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(bin); errors.Is(err, fs.ErrNotExist) {
		ui.Default.Info(tool, "is not installed yet; installing it")
		return bin, installFrom(ctx, tool, repo)
	}
	if r, err := store.ReadRelease(tool); err == nil && !updates.Due(r.CheckedAt, now()) {
		return bin, nil
	}
	if err := upgrade(ctx, tool, false); err != nil {
		ui.Default.Warn("Could not check for", tool, "updates:", firstLine(err))
	}
	return bin, nil
}

// Install installs the latest release of the tool named by arg (a tool
// name or "owner/repo"), replacing any installed copy.
func Install(ctx context.Context, arg string) error {
	tool, repo, err := parseTool(arg)
	if err != nil {
		return err
	}
	return installFrom(ctx, tool, repo)
}

// Upgrade upgrades the named tools, or every installed tool if none are named.
func Upgrade(ctx context.Context, args []string) error {
	if len(args) == 0 {
		installed, err := store.List()
		if err != nil {
			return err
		}
		if len(installed) == 0 {
			ui.Default.Info("No tools installed yet.")
			return nil
		}
		for _, t := range installed {
			args = append(args, t.Name)
		}
	}
	var errs []error
	for _, arg := range args {
		tool, _, err := parseTool(arg)
		if err == nil {
			err = upgrade(ctx, tool, true)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", arg, err))
		}
	}
	return errors.Join(errs...)
}

// upgrade installs tool's latest release if it is newer than the installed
// one, installing tool from scratch if it has no release record.
func upgrade(ctx context.Context, tool string, verbose bool) error {
	cur, err := store.ReadRelease(tool)
	if err != nil {
		return installFrom(ctx, tool, "")
	}
	repo, rel, err := resolve(ctx, tool, cur.Repository)
	if err != nil {
		return err
	}
	if !updates.Newer(cur.Tag, rel.Tag) {
		if verbose {
			ui.Default.Info(tool, cur.Tag, "is up to date")
		}
		cur.CheckedAt = now()
		return store.WriteRelease(tool, cur)
	}
	ui.Default.Info("Upgrading", tool, cur.Tag, "→", rel.Tag)
	return install(ctx, tool, repo, rel)
}

func installFrom(ctx context.Context, tool, repo string) error {
	repo, rel, err := resolve(ctx, tool, repo)
	if err != nil {
		return err
	}
	return install(ctx, tool, repo, rel)
}

// resolve finds the repository publishing tool and its latest release. If
// repo is non-empty, only that repository is consulted; otherwise
// platform.Candidates supplies the repositories to try.
func resolve(ctx context.Context, tool, repo string) (string, *github.Release, error) {
	repos := []string{repo}
	if repo == "" {
		repos = platform.Candidates(tool)
	}
	if len(repos) == 0 {
		return "", nil, fmt.Errorf("%[1]s doesn't know where %[2]s is published; install it once with '%[1]s install <owner>/%[2]s'",
			platform.Name, tool)
	}
	if err := github.Check(ctx); err != nil {
		return "", nil, err
	}
	for _, r := range repos {
		if rel, err := github.Latest(ctx, r); err == nil {
			return r, rel, nil
		}
	}
	return "", nil, notFound(ctx, tool, repos)
}

func notFound(ctx context.Context, tool string, repos []string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "no release of %s found; looked in:\n", tool)
	for _, r := range repos {
		fmt.Fprintf(&b, "  %s\n", r)
	}
	who := "your GitHub account"
	if user, err := github.User(ctx); err == nil && user != "" {
		who = "your GitHub account (" + user + ")"
	}
	fmt.Fprintf(&b, "If %s is in a private repository, make sure %s can access it.", tool, who)
	return errors.New(b.String())
}

func install(ctx context.Context, tool, repo string, rel *github.Release) error {
	asset := github.AssetName(tool, rel.Tag, runtime.GOOS, runtime.GOARCH)
	if !rel.Has(asset) {
		return fmt.Errorf("%s %s has no build for %s/%s (expected asset %s)",
			repo, rel.Tag, runtime.GOOS, runtime.GOARCH, asset)
	}

	tmp, err := os.MkdirTemp("", platform.Name+"-"+tool+"-")
	if err != nil {
		return fmt.Errorf("create download directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }() // best effort; the OS cleans temp eventually

	ui.Default.Info("Downloading", tool, rel.Tag, "from", repo)
	if err := github.Download(ctx, repo, rel.Tag, asset, tmp); err != nil {
		return err
	}

	dir, err := store.Dir(tool)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	bin := filepath.Join(dir, store.ExeName(tool))
	if err := archive.ExtractFile(filepath.Join(tmp, asset), store.ExeName(tool), bin); err != nil {
		return err
	}
	if err := store.WriteRelease(tool, store.Release{Tag: rel.Tag, Repository: repo, CheckedAt: now()}); err != nil {
		return err
	}
	ui.Default.Success("Installed", tool, rel.Tag)
	return nil
}

func firstLine(err error) string {
	msg, _, _ := strings.Cut(err.Error(), "\n")
	return msg
}
