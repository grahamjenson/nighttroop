package gen

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

var (
	pragma   = regexp.MustCompile(`(?i)\bPRAGMA\s+(?:\w+\.)?(\w+)`)
	comments = regexp.MustCompile(`(?s)--[^\n]*|/\*.*?\*/`)
)

// LintMigrations fails if a migration runs a PRAGMA other than defer_foreign_keys. Each migration runs in one
// transaction, where some do nothing without an error, such as foreign_keys = OFF and journal_mode, and the rest, such
// as ignore_check_constraints, last past the migration on its connection, which the server goes on to use.
// defer_foreign_keys ends with the transaction.
func LintMigrations(migrations fs.FS) error {
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	problems := []string{}
	for _, file := range files {
		text, err := fs.ReadFile(migrations, file)
		if err != nil {
			return err
		}
		for _, m := range pragma.FindAllStringSubmatch(comments.ReplaceAllString(string(text), ""), -1) {
			if !strings.EqualFold(m[1], "defer_foreign_keys") {
				problems = append(problems, fmt.Sprintf("%s runs PRAGMA %s", path.Base(file), m[1]))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s; in a migration's transaction a PRAGMA does nothing, or lasts past it on its connection, "+
			"so only PRAGMA defer_foreign_keys is allowed, see docs/migrations.md", strings.Join(problems, ", "))
	}
	return nil
}
