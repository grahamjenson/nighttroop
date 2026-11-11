package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func MkDirs(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}
	return nil
}

func MkParentDirs(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}
	return nil
}

// FileExists is whether path can be found; a path that cannot be checked, e.g. for permissions, does not exist
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DeleteFile removes the file at path, if there is one
func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("cannot delete file %s: %w", path, err)
	}
	return nil
}

// DeleteFiles removes each file there is, stopping at the first that cannot be removed
func DeleteFiles(paths []string) error {
	for _, path := range paths {
		if err := DeleteFile(path); err != nil {
			return err
		}
	}
	return nil
}

type CountingWriter struct {
	W     io.Writer
	Bytes int64
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.W.Write(p)
	c.Bytes += int64(n)
	return n, err
}

type CountingReader struct {
	R     io.Reader
	Bytes int64
}

func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.R.Read(p)
	c.Bytes += int64(n)
	return n, err
}

func FindProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}
