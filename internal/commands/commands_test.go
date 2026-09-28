package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xeger/installer/internal/platform"
	"github.com/xeger/installer/internal/store"
)

// These tests swap package-level state; do not mark them t.Parallel.

func withCandidates(t *testing.T, fn func(string) []string) {
	t.Helper()
	orig := platform.Candidates
	platform.Candidates = fn
	t.Cleanup(func() { platform.Candidates = orig })
}

func TestParseTool(t *testing.T) {
	tests := []struct {
		arg, tool, repo string
		wantErr         bool
	}{
		{arg: "foo", tool: "foo"},
		{arg: "acme/foo", tool: "foo", repo: "acme/foo"},
		{arg: "Acme-Corp/foo-cli", tool: "foo-cli", repo: "Acme-Corp/foo-cli"},
		{arg: "Foo", wantErr: true},
		{arg: "acme/Foo", wantErr: true},
		{arg: "acme/foo/bar", wantErr: true},
		{arg: "/foo", wantErr: true},
		{arg: "acme/", wantErr: true},
		{arg: "../foo", wantErr: true},
		{arg: "ac me/foo", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			tool, repo, err := parseTool(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTool(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
			}
			if tool != tt.tool || repo != tt.repo {
				t.Errorf("parseTool(%q) = %q, %q; want %q, %q", tt.arg, tool, repo, tt.tool, tt.repo)
			}
		})
	}
}

// A recently checked, installed tool must run without touching the network
// (and so without gh), which is what makes the launcher usable offline.
func TestEnsureRecentlyCheckedSkipsNetwork(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("PATH", "") // any attempt to run gh fails
	fixed := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })

	dir, err := store.Dir("foo")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, store.ExeName("foo"))
	if err := os.WriteFile(bin, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteRelease("foo", store.Release{Tag: "v1.0.0", CheckedAt: fixed.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}

	for _, arg := range []string{"foo", "acme/foo"} {
		got, err := Ensure(context.Background(), arg)
		if err != nil {
			t.Fatalf("Ensure(%q) error = %v", arg, err)
		}
		if got != bin {
			t.Errorf("Ensure(%q) = %q, want %q", arg, got, bin)
		}
	}
}

func TestEnsureRejectsBadName(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if _, err := Ensure(context.Background(), "../etc"); err == nil {
		t.Error("Ensure(../etc) succeeded")
	}
}

func TestInstallWithoutGH(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("PATH", "")
	withCandidates(t, func(tool string) []string { return []string{"acme/" + tool} })

	for _, arg := range []string{"foo", "acme/foo"} {
		err := Install(context.Background(), arg)
		if err == nil {
			t.Fatalf("Install(%q) without gh succeeded", arg)
		}
		if want := "gh auth login"; !strings.Contains(err.Error(), want) {
			t.Errorf("Install(%q) error %q should tell the user to run %q", arg, err, want)
		}
	}
}

func TestInstallUnknownTool(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("PATH", "") // must fail before needing gh
	withCandidates(t, func(string) []string { return nil })

	err := Install(context.Background(), "foo")
	if err == nil {
		t.Fatal("Install() of unknown tool succeeded")
	}
	want := platform.Name + " install <owner>/foo"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q should suggest %q", err, want)
	}
}
