package gen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

	if err := checkLimitsCast(folder); err != nil {
		return err
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

// bareLimit is a LIMIT or OFFSET whose value is a bare bound parameter: sqlc.arg(...), sqlc.narg(...), ?, ?N or @name
var bareLimit = regexp.MustCompile(`(?i)\b(LIMIT|OFFSET)\s+\(*\s*(sqlc\.n?arg\s*\(|\?|@\w)`)

// checkLimitsCast fails if a query's LIMIT or OFFSET is a bare bound parameter rather than one wrapped in
// CAST(... AS INTEGER). SQLite reads a bare bound LIMIT/OFFSET value while planning, which marks the statement to be
// re-prepared whenever that parameter is bound, i.e. on every run: a search took ~8x longer. A CAST is not read while
// planning, so the prepared statement is reused. Constant LIMITs (e.g. LIMIT 1) are fine.
func checkLimitsCast(folder string) error {
	files, err := filepath.Glob(filepath.Join(folder, "queries", "*.sql"))
	if err != nil {
		return err
	}
	var bare []string
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		for _, loc := range bareLimit.FindAllIndex(src, -1) {
			line := bytes.Count(src[:loc[0]], []byte("\n")) + 1
			bare = append(bare, fmt.Sprintf("%s:%d %s", file, line, strings.Join(strings.Fields(string(src[loc[0]:loc[1]])), " ")))
		}
	}
	if len(bare) > 0 {
		return fmt.Errorf("a bound LIMIT or OFFSET must be CAST(... AS INTEGER) so SQLite does not re-prepare the query on every run: %v", bare)
	}
	return nil
}
