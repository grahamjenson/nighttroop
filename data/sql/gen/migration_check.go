package gen

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// CheckMigrationsWithData migrates db up through each of the migrations in turn. At each version it adds rows with
// every column at the largest and the smallest values the CHECKs allow, see SeedLimits, then migrates down a version
// and up again, and adds them again; the first version, which makes the whole schema, is not migrated down until last.
// So a migration that cannot take what an earlier version allowed, such as one narrowing a CHECK, or a
// down migration that cannot take what its version allowed, fails here rather than on a deploy. Last it migrates down
// to nothing, with all the rows in. db should be opened as the server opens it, with foreign keys on.
func CheckMigrationsWithData(migrations fs.FS, db *sql.DB) error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return err
	}
	defer m.Close()

	first, err := source.First()
	for version := first; ; version, err = source.Next(version) {
		if errors.Is(err, fs.ErrNotExist) {
			break
		}
		if err != nil {
			return err
		}
		steps := []struct {
			n    int
			what string
		}{{1, "up to"}, {0, "adding rows at their limits at"}, {-1, "down from"}, {1, "up again to"},
			// Again, as migrating down may have dropped tables with their rows, so the next version starts with them
			{0, "adding rows at their limits again at"}}
		// The first migration makes the whole schema, which its down migration leaves, as nothing is kept without it
		if version == first {
			steps = steps[:2]
		}
		for _, step := range steps {
			if step.n == 0 {
				err = SeedLimits(db)
			} else {
				err = m.Steps(step.n)
			}
			if err != nil {
				return fmt.Errorf("migrating %s version %d: %w", step.what, version, err)
			}
		}
	}
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating down with rows at their limits: %w", err)
	}
	return nil
}

var (
	checkStart = regexp.MustCompile(`(?i)\bCHECK\s*\(`)
	// a column's limits, each matched with %[1]s as the column's name
	lengthAtMost  = `length\(\s*%[1]s\s*\)\s*<=\s*(\d+)`
	lengthExactly = `length\(\s*%[1]s\s*\)\s*=\s*(\d+)`
	lengthOver    = `length\(\s*%[1]s\s*\)\s*>\s*(\d+)`
	between       = `(?i)\b%[1]s\s+BETWEEN\s+(-?[\d.]+)\s+AND\s+(-?[\d.]+)`
	in            = `(?i)\b%[1]s\s+IN\s*\(([^)]*)\)`
	equals        = `(?:^|[^\w.])%[1]s\s*=\s*(-?\d+)`
	atLeast       = `(?:^|[^\w.])%[1]s\s*(>=?)\s*(-?[\d.]+)`
	atMost        = `(?:^|[^\w.])%[1]s\s*(<=?)\s*(-?[\d.]+)`
	jsonValid     = `json_valid\(\s*%[1]s\s*\)`
)

// column is a table's column and the limits its CHECKs set
type column struct {
	name, kind          string
	rowid               bool // an INTEGER PRIMARY KEY SQLite numbers
	minLen, maxLen, len int
	low, high           string
	in                  []string
	equals              string
	json                bool
	parent, parentKey   string // what a foreign key references
}

// SeedLimits adds two rows to each table: one with each column at the largest value its CHECKs allow, the longest
// text, the highest number, the last of a list, and one at the smallest. A text with no limit is short, and a number
// with none is far from 0. A foreign key is a row of its parent at the same end, so parents are added first. A row a
// table already has the key of, such as a table of one row, is left out.
func SeedLimits(db *sql.DB) error {
	tables, err := seedOrder(db)
	if err != nil {
		return err
	}
	added := map[string]map[bool]map[string]any{} // each table's rows added, at each end, by column
	for _, table := range tables {
		columns, err := tableColumns(db, table)
		if err != nil {
			return err
		}
		added[table] = map[bool]map[string]any{}
		for _, largest := range []bool{true, false} {
			row, names, values := map[string]any{}, []string{}, []any{}
			for _, c := range columns {
				if c.rowid {
					continue
				}
				value := c.limit(largest)
				if c.parent != "" {
					if value, err = parentKey(db, added, c, largest); err != nil {
						return err
					}
				}
				row[c.name] = value
				names = append(names, `"`+c.name+`"`)
				values = append(values, value)
			}
			marks := strings.TrimSuffix(strings.Repeat("?, ", len(values)), ", ")
			insert := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, table, strings.Join(names, ", "), marks)
			// A table of only a row ID has no column to give
			if len(names) == 0 {
				insert = fmt.Sprintf(`INSERT INTO "%s" DEFAULT VALUES`, table)
			}
			result, err := db.Exec(insert, values...)
			if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			if err != nil {
				return fmt.Errorf("cannot add a row at its limits to %s: %w; %v", table, err, shortened(row))
			}
			for _, c := range columns {
				if c.rowid {
					row[c.name], _ = result.LastInsertId()
				}
			}
			added[table][largest] = row
		}
	}
	return nil
}

// limit is the column's largest or smallest value
func (c column) limit(largest bool) any {
	switch {
	case c.equals != "":
		return literal(c.equals)
	case len(c.in) > 0 && largest:
		return literal(c.in[len(c.in)-1])
	case len(c.in) > 0:
		return literal(c.in[0])
	case c.high != "" && largest:
		return literal(c.high)
	case c.low != "":
		return literal(c.low)
	}
	switch c.kind {
	case "TEXT":
		length, char := max(c.minLen, 1), "a"
		if largest {
			length, char = max(c.maxLen, length), "z"
		}
		if c.len > 0 {
			length = c.len
		}
		if c.json {
			// A JSON string is valid JSON, and fills the length with its quotes
			return `"` + strings.Repeat(char, max(length-2, 1)) + `"`
		}
		return strings.Repeat(char, length)
	case "INTEGER":
		if largest {
			return int64(1<<53 - 1)
		}
		return int64(-(1<<53 - 1))
	case "REAL":
		if largest {
			return 1e15
		}
		return -1e15
	case "BLOB":
		if largest {
			return []byte{0xff}
		}
		return []byte{0}
	}
	return "a"
}

// bound is a CHECK's bound n, or for a strict one the whole number by past it
func bound(n string, strict bool, by int64) string {
	if i, err := strconv.ParseInt(n, 10, 64); err == nil && strict {
		return strconv.FormatInt(i+by, 10)
	}
	return n
}

// shortened is the row with each long text cut short, to say in an error
func shortened(row map[string]any) map[string]any {
	short := map[string]any{}
	for name, value := range row {
		if s, ok := value.(string); ok && len(s) > 20 {
			value = fmt.Sprintf("%s… (%d characters)", s[:10], len(s))
		}
		short[name] = value
	}
	return short
}

// literal is a number, or a string with its quotes taken off, as a CHECK writes it
func literal(s string) any {
	s = strings.TrimSpace(s)
	if unquoted, ok := strings.CutPrefix(s, "'"); ok {
		return strings.ReplaceAll(strings.TrimSuffix(unquoted, "'"), "''", "'")
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// parentKey is the key the column references in its parent's row at the same end, or any row of its parent
func parentKey(db *sql.DB, added map[string]map[bool]map[string]any, c column, largest bool) (any, error) {
	if row, ok := added[c.parent][largest]; ok {
		if value, ok := row[c.parentKey]; ok {
			return value, nil
		}
	}
	var value any
	err := db.QueryRow(fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 1`, c.parentKey, c.parent)).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return value, err
}

// tableColumns are the table's columns, but not those it generates, with the limits its CHECKs and foreign keys set
func tableColumns(db *sql.DB, table string) ([]column, error) {
	var createSQL string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&createSQL); err != nil {
		return nil, err
	}
	checks := checkExpressions(createSQL)

	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_xinfo("%s")`, table))
	if err != nil {
		return nil, err
	}
	columns := []column{}
	for rows.Next() {
		var cid, notNull, pk, hidden int
		var name, kind string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &pk, &hidden); err != nil {
			rows.Close()
			return nil, err
		}
		if hidden != 0 {
			continue
		}
		c := column{name: name, kind: strings.ToUpper(kind)}
		c.readChecks(checks)
		// An INTEGER PRIMARY KEY is numbered by SQLite, unless a CHECK limits it, as in a table of one row
		c.rowid = pk == 1 && c.kind == "INTEGER" && c.equals == "" && c.low == "" && len(c.in) == 0
		columns = append(columns, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	type foreignKey struct{ parent, from, to string }
	keys := []foreignKey{}
	keyRows, err := db.Query(fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, table))
	if err != nil {
		return nil, err
	}
	for keyRows.Next() {
		var id, seq int
		var parent, from, onUpdate, onDelete, match string
		var to sql.NullString
		if err := keyRows.Scan(&id, &seq, &parent, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			keyRows.Close()
			return nil, err
		}
		keys = append(keys, foreignKey{parent, from, to.String})
	}
	keyRows.Close()
	if err := keyRows.Err(); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if key.parent == table {
			continue
		}
		// A key naming no column references its parent's primary key
		if key.to == "" {
			if err := db.QueryRow(`SELECT name FROM pragma_table_info(?) WHERE pk = 1`, key.parent).Scan(&key.to); err != nil {
				return nil, err
			}
		}
		for i := range columns {
			if columns[i].name == key.from {
				columns[i].parent, columns[i].parentKey = key.parent, key.to
			}
		}
	}
	return columns, nil
}

// readChecks sets the column's limits from the table's CHECKs that name it
func (c *column) readChecks(checks []string) {
	name := regexp.QuoteMeta(c.name)
	match := func(pattern string, check string) []string {
		return regexp.MustCompile(fmt.Sprintf(pattern, name)).FindStringSubmatch(check)
	}
	for _, check := range checks {
		if m := match(lengthAtMost, check); m != nil {
			c.maxLen, _ = strconv.Atoi(m[1])
		}
		if m := match(lengthExactly, check); m != nil {
			c.len, _ = strconv.Atoi(m[1])
		}
		if m := match(lengthOver, check); m != nil {
			n, _ := strconv.Atoi(m[1])
			c.minLen = n + 1
		}
		if m := match(between, check); m != nil {
			c.low, c.high = m[1], m[2]
		}
		// A strict bound is a whole number past it, as these are on INTEGER columns
		if m := match(atLeast, check); m != nil {
			c.low = bound(m[2], m[1] == ">", 1)
		}
		if m := match(atMost, check); m != nil {
			c.high = bound(m[2], m[1] == "<", -1)
		}
		if m := match(in, check); m != nil {
			c.in = strings.Split(m[1], ",")
		}
		if m := match(equals, check); m != nil && match(lengthExactly, check) == nil {
			c.equals = m[1]
		}
		if match(jsonValid, check) != nil {
			c.json = true
		}
	}
}

// checkExpressions are what each CHECK in a CREATE TABLE checks
func checkExpressions(createSQL string) []string {
	checks := []string{}
	for _, loc := range checkStart.FindAllStringIndex(createSQL, -1) {
		depth := 1
		for i := loc[1]; i < len(createSQL); i++ {
			switch createSQL[i] {
			case '(':
				depth++
			case ')':
				depth--
			}
			if depth == 0 {
				checks = append(checks, createSQL[loc[1]:i])
				break
			}
		}
	}
	return checks
}

// seedOrder is every table, each after those its foreign keys reference
func seedOrder(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name <> 'schema_migrations' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	tables := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	parents := map[string][]string{}
	for _, table := range tables {
		keys, err := db.Query(fmt.Sprintf(`SELECT "table" FROM pragma_foreign_key_list('%s')`, table))
		if err != nil {
			return nil, err
		}
		for keys.Next() {
			var parent string
			if err := keys.Scan(&parent); err != nil {
				keys.Close()
				return nil, err
			}
			if parent != table {
				parents[table] = append(parents[table], parent)
			}
		}
		keys.Close()
	}
	ordered, seen := []string{}, map[string]bool{}
	var visit func(string)
	visit = func(table string) {
		if seen[table] {
			return
		}
		seen[table] = true
		for _, parent := range parents[table] {
			visit(parent)
		}
		ordered = append(ordered, table)
	}
	for _, table := range tables {
		visit(table)
	}
	return ordered, nil
}
