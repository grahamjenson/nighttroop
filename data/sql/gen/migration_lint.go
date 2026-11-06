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

	additive   = regexp.MustCompile(`(?m)^-- additive$`)
	adds       = regexp.MustCompile(`(?is)^\s*(CREATE\s+(TABLE|INDEX|VIEW)\s|ALTER\s+TABLE\s+\S+\s+ADD\s)`)
	references = regexp.MustCompile(`(?i)\bREFERENCES\b`)
	onDelete   = regexp.MustCompile(`(?i)\bON\s+DELETE\s+(CASCADE|SET\s+NULL)\b`)
)

// LintMigrations fails if a migration runs a PRAGMA other than defer_foreign_keys. Each migration runs in one
// transaction, where some do nothing without an error, such as foreign_keys = OFF and journal_mode, and the rest, such
// as ignore_check_constraints, last past the migration on its connection, which the server goes on to use.
// defer_foreign_keys ends with the transaction.
//
// It also fails if a migration marked additive, which a build without it goes on using the DB after, does more than
// add tables, indexes, views and columns, or references a table without ON DELETE CASCADE or SET NULL, which would
// stop that build deleting the rows referenced.
func LintMigrations(migrations fs.FS) error {
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	problems, changes := []string{}, []string{}
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
		if additive.Match(text) && strings.HasSuffix(file, ".up.sql") {
			changes = append(changes, notAdditive(path.Base(file), comments.ReplaceAllString(string(text), ""))...)
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s; in a migration's transaction a PRAGMA does nothing, or lasts past it on its connection, "+
			"so only PRAGMA defer_foreign_keys is allowed, see docs/migrations.md", strings.Join(problems, ", "))
	}
	if len(changes) > 0 {
		return fmt.Errorf("%s; a build without an additive migration goes on using the DB after it, so it may only add "+
			"tables, indexes, views and columns, and reference a table ON DELETE CASCADE or SET NULL, see docs/migrations.md",
			strings.Join(changes, ", "))
	}
	return nil
}

// notAdditive is what the additive migration file, without its comments, does other than add to the schema
func notAdditive(file, sql string) []string {
	changes := []string{}
	for _, stmt := range strings.Split(sql, ";") {
		if strings.TrimSpace(stmt) != "" && !adds.MatchString(stmt) {
			changes = append(changes, fmt.Sprintf("%s runs %q", file, strings.Join(strings.Fields(stmt), " ")))
		}
	}
	if len(references.FindAllString(sql, -1)) > len(onDelete.FindAllString(sql, -1)) {
		changes = append(changes, file+" references a table without ON DELETE CASCADE or SET NULL")
	}
	return changes
}
