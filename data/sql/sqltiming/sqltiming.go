// Package sqltiming times every query each DB runs, by wrapping the SQLite driver's connections, and logs any slower
// than a threshold with the DB and the sqlc query's name, so the slow ones can be found and worked on. A query's
// values are never logged, as they hold the site's data.
package sqltiming

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log/slog"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
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

// record logs a query that took at least the slow threshold, with how many rows it gave, or -1 for one that gives none,
// and keeps a sample of every query's times
func record(ctx context.Context, db, name string, start time.Time, rows int) {
	took := time.Since(start)
	sample(db, name, took)
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

// samplePercent is how many in a hundred queries are kept in the stats
var samplePercent atomic.Int64

// SetSamplePercent sets how many in a hundred queries are kept in the stats; 0 keeps none
func SetSamplePercent(percent int) {
	samplePercent.Store(int64(percent))
}

// bucketsMs are the upper bounds, in milliseconds, of the times a query's sampled runs are counted in, for its p95
var bucketsMs = [...]float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, math.Inf(1)}

// queryStats is a query's sampled runs: how many, their total and longest time, and how many took up to each bucket
type queryStats struct {
	count      int64
	total, max time.Duration
	buckets    [len(bucketsMs)]int64
}

// stats is each DB's queries' sampled runs since the server started, by DB and query name
var stats = struct {
	sync.Mutex
	byQuery map[[2]string]*queryStats
}{byQuery: map[[2]string]*queryStats{}}

// sample keeps one in a hundred queries' times for every samplePercent
func sample(db, name string, took time.Duration) {
	if percent := samplePercent.Load(); percent <= 0 || rand.Int64N(100) >= percent {
		return
	}
	stats.Lock()
	defer stats.Unlock()
	s := stats.byQuery[[2]string{db, name}]
	if s == nil {
		s = &queryStats{}
		stats.byQuery[[2]string{db, name}] = s
	}
	s.count++
	s.total += took
	s.max = max(s.max, took)
	s.buckets[sort.SearchFloat64s(bucketsMs[:], ms(took))]++
}

// QueryStat is a query's sampled runs since the server started: how many, and their total, mean, 95th percentile (the
// upper bound of the times it is among, at most the longest) and longest time, in milliseconds
type QueryStat struct {
	DB      string  `json:"db"`
	Query   string  `json:"query"`
	Count   int64   `json:"count"`
	TotalMs float64 `json:"total_ms"`
	MeanMs  float64 `json:"mean_ms"`
	P95Ms   float64 `json:"p95_ms"`
	MaxMs   float64 `json:"max_ms"`
}

// Slowest are the limit queries whose sampled runs took the longest in all, longest first, as where time is spent
func Slowest(limit int) []QueryStat {
	stats.Lock()
	defer stats.Unlock()
	slowest := make([]QueryStat, 0, len(stats.byQuery))
	for key, s := range stats.byQuery {
		p95, need, seen := ms(s.max), (s.count*95+99)/100, int64(0)
		for i, n := range s.buckets {
			if seen += n; seen >= need {
				p95 = min(bucketsMs[i], p95)
				break
			}
		}
		slowest = append(slowest, QueryStat{DB: key[0], Query: key[1], Count: s.count, TotalMs: ms(s.total),
			MeanMs: ms(s.total) / float64(s.count), P95Ms: p95, MaxMs: ms(s.max)})
	}
	sort.Slice(slowest, func(i, j int) bool { return slowest[i].TotalMs > slowest[j].TotalMs })
	return slowest[:min(limit, len(slowest))]
}

func ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
