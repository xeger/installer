package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ExtractionOptions struct {
	DestDir       string
	FileFilter    func(string) bool
	StreamMode    bool
	TargetBinary  string
}

func ExtractTarGz(src string, opts ExtractionOptions) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	return ExtractTarGzStream(file, opts)
}

func ExtractTarGzStream(reader io.Reader, opts ExtractionOptions) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		filename := header.Name
		if opts.FileFilter != nil && !opts.FileFilter(filename) {
			continue
		}

		if opts.StreamMode && opts.TargetBinary != "" {
			baseName := filepath.Base(filename)
			if baseName == opts.TargetBinary {
				return extractToExecutable(tr, opts.TargetBinary, header.Mode)
			}
			continue
		}

		destPath := filepath.Join(opts.DestDir, filename)
		if err := extractFile(tr, destPath, header.Mode); err != nil {
			return fmt.Errorf("failed to extract %s: %w", filename, err)
		}
	}

	return nil
}

func extractFile(reader io.Reader, destPath string, mode int64) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func extractToExecutable(reader io.Reader, executablePath string, mode int64) error {
	tempPath := executablePath + ".tmp"
	
	if err := extractFile(reader, tempPath, mode); err != nil {
		return err
	}

	if err := os.Rename(tempPath, executablePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

func DefaultFileFilter(commandName string) func(string) bool {
	return func(filename string) bool {
		base := filepath.Base(filename)
		return base == commandName || strings.HasPrefix(base, commandName+"_")
	}
}