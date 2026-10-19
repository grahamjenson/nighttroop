package utils

import (
	"fmt"
	"io"
	"log/slog"
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

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if !os.IsNotExist(err) {
		return false
	}
	return false
}

func DeleteFile(path string) error {
	if FileExists(path) {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("cannot delete file %s: %w", path, err)
		}
	}
	return nil
}

func DeleteFiles(paths []string) error {
	for _, path := range paths {
		err := DeleteFile(path)
		if err != nil {
			return fmt.Errorf("cannot delete file %s: %w", path, err)
		}
	}
	return nil
}

func RelPathBetween(newBasePath string, targetPath string) string {

	absFile, _ := filepath.Abs(targetPath)
	absBase, _ := filepath.Abs(newBasePath)

	relPath, err := filepath.Rel(absBase, absFile)
	if err != nil {
		slog.Error("Error calculating relative path", "base", absBase, "file", absFile, "error", err)
		return absFile
	}

	return relPath
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
