package clikit

import (
	"os"
	"runtime/debug"
	"testing"
)

// withBuildInfo temporarily replaces versionReadBuildInfo for the duration
// of the test.
func withBuildInfo(t *testing.T, info *debug.BuildInfo, ok bool) {
	t.Helper()
	orig := versionReadBuildInfo
	versionReadBuildInfo = func() (*debug.BuildInfo, bool) { return info, ok }
	t.Cleanup(func() { versionReadBuildInfo = orig })
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name        string
		ldflags     string
		buildInfo   *debug.BuildInfo
		buildInfoOK bool
		want        string
	}{
		{
			name:        "ldflags value wins",
			ldflags:     "v1.2.3",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}},
			buildInfoOK: true,
			want:        "v1.2.3",
		},
		{
			name:        "falls back to module version",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "v0.5.0"}},
			buildInfoOK: true,
			want:        "v0.5.0",
		},
		{
			name:        "devel module version is ignored",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			buildInfoOK: true,
			want:        "dev",
		},
		{
			name:        "empty module version is ignored",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: ""}},
			buildInfoOK: true,
			want:        "dev",
		},
		{
			name:        "no build info",
			buildInfoOK: false,
			want:        "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := version
			version = tt.ldflags
			t.Cleanup(func() { version = orig })
			withBuildInfo(t, tt.buildInfo, tt.buildInfoOK)

			if got := Version(); got != tt.want {
				t.Errorf("Version() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommit(t *testing.T) {
	tests := []struct {
		name        string
		buildInfo   *debug.BuildInfo
		buildInfoOK bool
		want        string
	}{
		{
			name:        "revision present",
			buildInfoOK: true,
			buildInfo: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
			}},
			want: "abc123",
		},
		{
			name:        "no revision setting",
			buildInfoOK: true,
			buildInfo:   &debug.BuildInfo{},
			want:        "",
		},
		{
			name:        "no build info",
			buildInfoOK: false,
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withBuildInfo(t, tt.buildInfo, tt.buildInfoOK)
			if got := Commit(); got != tt.want {
				t.Errorf("Commit() = %q, want %q", got, tt.want)
			}
		})
	}
}

// These tests replace package-level state (version, os.Args, the build info
// seam); do not mark them t.Parallel.

func TestSetVersion(t *testing.T) {
	origVersion := version
	t.Cleanup(func() { version = origVersion })
	withBuildInfo(t, &debug.BuildInfo{Main: debug.Module{Version: "v0.0.1"}}, true)

	version = ""
	SetVersion("")
	if got := Version(); got != "v0.0.1" {
		t.Errorf("after SetVersion(\"\"), Version() = %q, want build info fallback %q", got, "v0.0.1")
	}
	SetVersion("v2.0.0")
	if got := Version(); got != "v2.0.0" {
		t.Errorf("after SetVersion(%q), Version() = %q", "v2.0.0", got)
	}
}

func TestVersionInfo(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Run("with commit", func(t *testing.T) {
		os.Args = []string{"/usr/local/bin/mytool", "arg"}
		origVersion := version
		version = "v1.2.3"
		t.Cleanup(func() { version = origVersion })
		withBuildInfo(t, &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "deadbeef"},
		}}, true)

		want := "mytool v1.2.3 (deadbeef)"
		if got := VersionInfo(); got != want {
			t.Errorf("VersionInfo() = %q, want %q", got, want)
		}
	})

	t.Run("without commit omits parenthetical", func(t *testing.T) {
		os.Args = []string{"/usr/local/bin/mytool"}
		origVersion := version
		version = "v1.2.3"
		t.Cleanup(func() { version = origVersion })
		withBuildInfo(t, &debug.BuildInfo{}, true)

		want := "mytool v1.2.3"
		if got := VersionInfo(); got != want {
			t.Errorf("VersionInfo() = %q, want %q", got, want)
		}
	})

	t.Run("strips .exe suffix", func(t *testing.T) {
		os.Args = []string{"/opt/tools/mytool.exe"}
		origVersion := version
		version = "v1.2.3"
		t.Cleanup(func() { version = origVersion })
		withBuildInfo(t, &debug.BuildInfo{}, true)

		if got, want := VersionInfo(), "mytool v1.2.3"; got != want {
			t.Errorf("VersionInfo() = %q, want %q", got, want)
		}
	})
}
