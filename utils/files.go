package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

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
