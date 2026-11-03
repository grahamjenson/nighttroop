// Package sqltiming times every query each DB runs, by wrapping the SQLite driver's connections, and logs any slower
// than a threshold with the DB and the sqlc query's name, so the slow ones can be found and worked on. A query's
// values are never logged, as they hold the site's data.
package sqltiming

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"modernc.org/sqlite"
)

// slowNanos is how long a query runs before it is logged as slow; 0 logs none
var slowNanos atomic.Int64

// SetSlowThreshold sets how long a query runs before it is logged as slow; 0 logs none
func SetSlowThreshold(d time.Duration) {
	slowNanos.Store(int64(d))
}

// Open is sql.Open("sqlite", dsn), with each query of the DB named db timed
func Open(db string, dsn string) (*sql.DB, error) {
	c, err := sqlite.NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(connector{Connector: c, db: db}), nil
}

type connector struct {
	driver.Connector
	db string
}

// sqliteConn is what the driver's connection does, all of which database/sql uses
type sqliteConn interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
	driver.Pinger
	driver.SessionResetter
	driver.Validator
}

// sqliteStmt is what the driver's prepared statement does
type sqliteStmt interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if inner, ok := conn.(sqliteConn); ok && err == nil {
		return &timedConn{sqliteConn: inner, db: c.db}, nil
	}
	return conn, err
}

type timedConn struct {
	sqliteConn
	db string
}

// Unwrap is the driver's own connection, for what only it does, e.g. backing up the DB
func (c *timedConn) Unwrap() driver.Conn {
	return c.sqliteConn
}

func (c *timedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	stmt, err := c.sqliteConn.PrepareContext(ctx, query)
	if inner, ok := stmt.(sqliteStmt); ok && err == nil {
		return &timedStmt{sqliteStmt: inner, db: c.db, name: Name(query)}, nil
	}
	return stmt, err
}

func (c *timedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	result, err := c.sqliteConn.ExecContext(ctx, query, args)
	record(ctx, c.db, Name(query), start, -1)
	return result, err
}

func (c *timedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return timeRows(ctx, c.db, Name(query), func() (driver.Rows, error) { return c.sqliteConn.QueryContext(ctx, query, args) })
}

// BeginTx is timed as BEGIN, as a write transaction waits there for the write lock
func (c *timedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	start := time.Now()
	tx, err := c.sqliteConn.BeginTx(ctx, opts)
	record(ctx, c.db, "BEGIN", start, -1)
	return tx, err
}

type timedStmt struct {
	sqliteStmt
	db, name string
}

func (s *timedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	result, err := s.sqliteStmt.ExecContext(ctx, args)
	record(ctx, s.db, s.name, start, -1)
	return result, err
}

func (s *timedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return timeRows(ctx, s.db, s.name, func() (driver.Rows, error) { return s.sqliteStmt.QueryContext(ctx, args) })
}

// timeRows times a query until its rows are closed, as SQLite does most of a query's work as its rows are stepped
// through
func timeRows(ctx context.Context, db, name string, query func() (driver.Rows, error)) (driver.Rows, error) {
	start := time.Now()
	rows, err := query()
	if err != nil {
		record(ctx, db, name, start, -1)
		return rows, err
	}
	return &timedRows{Rows: rows, ctx: ctx, db: db, name: name, start: start}, nil
}

type timedRows struct {
	driver.Rows
	ctx      context.Context
	db, name string
	start    time.Time
	count    int
}

func (r *timedRows) Next(dest []driver.Value) error {
	err := r.Rows.Next(dest)
	if err == nil {
		r.count++
	}
	return err
}

func (r *timedRows) Close() error {
	err := r.Rows.Close()
	record(r.ctx, r.db, r.name, r.start, r.count)
	return err
}

// record logs a query that took at least the slow threshold, with how many rows it gave, or -1 for one that gives none
func record(ctx context.Context, db, name string, start time.Time, rows int) {
	took := time.Since(start)
	if slow := slowNanos.Load(); slow > 0 && took >= time.Duration(slow) {
		slog.WarnContext(ctx, "Slow query", "db", db, "query", name, "ms", took.Milliseconds(), "rows", rows)
	}
}

// Name is an sqlc query's name, from the "-- name: Name :many" it starts with, or else the first two words after any
// comments of one that is not, e.g. a migration's CREATE TABLE or a PRAGMA, never its values
func Name(query string) string {
	for {
		query = strings.TrimSpace(query)
		if rest, ok := strings.CutPrefix(query, "-- name: "); ok {
			name, _, _ := strings.Cut(rest, " ")
			return name
		}
		if !strings.HasPrefix(query, "--") {
			break
		}
		_, query, _ = strings.Cut(query, "\n")
	}
	words := strings.Fields(query)
	return strings.ToUpper(strings.Join(words[:min(2, len(words))], " "))
}
