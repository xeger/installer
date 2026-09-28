// Package archive unpacks tool release archives.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

// ExtractFile copies the regular file named base, found at any depth in the
// gzipped tarball src, to dest as an executable. dest is replaced
// atomically, so a failed extraction never leaves a truncated binary.
// Only base is extracted, so archive paths can never write outside dest.
func ExtractFile(src, base, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer func() { _ = f.Close() }() // read-only; nothing to flush

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}
	defer func() { _ = gz.Close() }() // read-only; nothing to flush

	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s not found in %s", base, path.Base(src))
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		if h.Typeflag == tar.TypeReg && path.Base(h.Name) == base {
			return writeExecutable(tr, dest)
		}
	}
}

func writeExecutable(r io.Reader, dest string) error {
	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", tmp, err)
	}
	// discard removes the partial file; the original error is what matters.
	discard := func() { _ = os.Remove(tmp) }
	if _, err := io.Copy(out, r); err != nil {
		_ = out.Close()
		discard()
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := out.Close(); err != nil {
		discard()
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		discard()
		return fmt.Errorf("install %s: %w", dest, err)
	}
	return nil
}
