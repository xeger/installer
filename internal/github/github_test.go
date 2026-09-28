package github

import "testing"

func TestAssetName(t *testing.T) {
	if got, want := AssetName("foo", "v1.2.3", "darwin", "arm64"), "foo_v1.2.3_darwin_arm64.tar.gz"; got != want {
		t.Errorf("AssetName() = %q, want %q", got, want)
	}
}

func TestHas(t *testing.T) {
	r := &Release{Assets: []Asset{{Name: "a"}, {Name: "b"}}}
	if !r.Has("b") || r.Has("c") {
		t.Errorf("Has() wrong for %+v", r.Assets)
	}
}

func TestInstallHint(t *testing.T) {
	for goos, want := range map[string]string{
		"darwin":  "brew install gh",
		"windows": "winget install --id GitHub.cli",
		"linux":   "see https://github.com/cli/cli#installation",
	} {
		if got := installHint(goos); got != want {
			t.Errorf("installHint(%q) = %q, want %q", goos, got, want)
		}
	}
}
