package gen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func BuildSQLC(folder string) error {
	if folder == "" {
		panic("Folder is required")
	}

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		panic("Folder does not exist: " + folder)
	}

	if fileInfo, err := os.Stat(folder); err != nil || !fileInfo.IsDir() {
		panic("Invalid folder: " + folder)
	}

	// execute `sqlc generate` in the folder
	cmd := exec.Command("sqlc", "generate")
	cmd.Dir = folder
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute sqlc generate: %w, output: %s", err, string(output))
	}

	return joinTablesNotViews(folder)
}

// viewJoin is a join on one of the _vw views, which are each all of their table's rows
var viewJoin = regexp.MustCompile(`(JOIN \w+)_vw `)

// joinTablesNotViews makes the generated queries join the tables rather than their _vw views. The views are only there
// so sqlc types a LEFT JOIN's embedded columns as nullable (https://github.com/sqlc-dev/sqlc/issues/2348), but as each
// view is a join, SQLite cannot fold it into the query: it copies the view's whole table for every query that joins it.
func joinTablesNotViews(folder string) error {
	files, err := filepath.Glob(filepath.Join(folder, "*.sql.gen.go"))
	if err != nil {
		return err
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if out := viewJoin.ReplaceAll(src, []byte("$1 ")); !bytes.Equal(src, out) {
			if err := os.WriteFile(file, out, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
