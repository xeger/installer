package archive

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func writeTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tool.tar.gz")
	f, err := os.Create(path) //nolint:gosec // test
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []interface{ Close() error }{tw, gz, f} {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestExtractFile(t *testing.T) {
	src := writeTarGz(t, map[string]string{
		"README.md":          "docs",
		"dist/linux/foo":     "binary",
		"../../escape/other": "nope",
	})
	dest := filepath.Join(t.TempDir(), "foo")

	if err := ExtractFile(src, "foo", dest); err != nil {
		t.Fatalf("ExtractFile() error = %v", err)
	}
	got, err := os.ReadFile(dest) //nolint:gosec // test
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "binary" {
		t.Errorf("extracted %q, want %q", got, "binary")
	}
	if info, _ := os.Stat(dest); info.Mode().Perm()&0o100 == 0 && os.PathSeparator == '/' {
		t.Errorf("extracted mode %v is not executable", info.Mode())
	}
	if _, err := os.Stat(dest + ".tmp"); !os.IsNotExist(err) {
		t.Error("temporary file left behind")
	}
}

func TestExtractFileMissing(t *testing.T) {
	src := writeTarGz(t, map[string]string{"bar": "x"})
	dest := filepath.Join(t.TempDir(), "foo")
	if err := ExtractFile(src, "foo", dest); err == nil {
		t.Fatal("ExtractFile() of missing entry succeeded")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("dest created despite failure")
	}
}
