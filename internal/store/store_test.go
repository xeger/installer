package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/xeger/installer/internal/platform"
)

func withDataHome(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{"foo", "foo-cli", "a.b_c-1", "9lives"} {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
	for _, name := range []string{"", "-rf", ".hidden", "..", "Foo", "a/b", `a\b`, "a b", "../etc"} {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
}

func TestBinary(t *testing.T) {
	withDataHome(t)
	got, err := Binary("foo")
	if err != nil {
		t.Fatal(err)
	}
	root, err := platform.DataDir("tools")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "foo", ExeName("foo"))
	if got != want {
		t.Errorf("Binary() = %q, want %q", got, want)
	}
	if _, err := Binary("../foo"); err == nil {
		t.Error("Binary(../foo) should fail")
	}
}

func TestReleaseRoundTrip(t *testing.T) {
	withDataHome(t)
	if _, err := ReadRelease("foo"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadRelease() of missing tool = %v, want fs.ErrNotExist", err)
	}
	dir, _ := Dir("foo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := Release{Tag: "v1.2.3", Repository: "acme/foo", CheckedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}
	if err := WriteRelease("foo", want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadRelease("foo")
	if err != nil {
		t.Fatal(err)
	}
	if !got.CheckedAt.Equal(want.CheckedAt) || got.Tag != want.Tag || got.Repository != want.Repository {
		t.Errorf("ReadRelease() = %+v, want %+v", got, want)
	}
}

func TestList(t *testing.T) {
	withDataHome(t)
	if tools, err := List(); err != nil || tools != nil {
		t.Fatalf("List() with nothing installed = %v, %v; want nil, nil", tools, err)
	}

	install := func(name string, withBinary bool, tag string) {
		t.Helper()
		root, err := platform.DataDir("tools")
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if withBinary {
			if err := os.WriteFile(filepath.Join(dir, ExeName(name)), nil, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if tag != "" {
			if err := WriteRelease(name, Release{Tag: tag}); err != nil {
				t.Fatal(err)
			}
		}
	}
	install("zeta", true, "v2.0.0")
	install("alpha", true, "")
	install("nobinary", false, "v1.0.0")
	install("Invalid", true, "")

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := []Tool{{Name: "alpha"}, {Name: "zeta", Tag: "v2.0.0"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}
