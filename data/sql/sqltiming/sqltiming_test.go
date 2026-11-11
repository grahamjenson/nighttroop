package sqltiming

import (
	"bytes"
	"database/sql"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// logs is what the default logger writes while the test runs, with each query that took at least slow logged
func logs(t *testing.T, slow time.Duration) *bytes.Buffer {
	var buf bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	SetSlowThreshold(slow)
	t.Cleanup(func() { slog.SetDefault(before); SetSlowThreshold(0) })
	return &buf
}

// Every way a DB runs a query is timed, under its sqlc name and with how many rows it gave, and a write transaction's
// wait for the lock as BEGIN; the query's values never are
func TestSlowQueriesAreLogged(t *testing.T) {
	db, err := Open("testdb", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	ctx := t.Context()
	_, err = db.ExecContext(ctx, "-- name: CreateNotes :exec\nCREATE TABLE notes (secret TEXT)")
	require.NoError(t, err)

	buf := logs(t, time.Nanosecond)
	stmt, err := db.PrepareContext(ctx, "-- name: AddNote :exec\nINSERT INTO notes (secret) VALUES (?)")
	require.NoError(t, err)
	for _, note := range []string{"my secret", "another secret"} {
		_, err = stmt.ExecContext(ctx, note)
		require.NoError(t, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	rows, err := tx.QueryContext(ctx, "-- name: SearchNotes :many\nSELECT secret FROM notes WHERE secret LIKE ?", "%secret")
	require.NoError(t, err)
	for rows.Next() {
	}
	require.NoError(t, rows.Close())
	require.NoError(t, tx.Commit())
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM notes").Scan(&count))

	out := buf.String()
	require.Len(t, regexp.MustCompile(`msg="Slow query" db=testdb query=AddNote ms=\d+ rows=-1`).FindAllString(out, -1), 2, out)
	require.Regexp(t, `msg="Slow query" db=testdb query=BEGIN ms=\d+ rows=-1`, out)
	require.Regexp(t, `query=SearchNotes ms=\d+ rows=2`, out)
	require.Regexp(t, `query="SELECT COUNT\(\*\)" ms=\d+ rows=1`, out)
	require.NotContains(t, out, "secret", "a query's values are never logged")
}

// A query quicker than the threshold is not logged, nor is any with no threshold
func TestQuickQueriesAreNotLogged(t *testing.T) {
	db, err := Open("testdb", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	for _, slow := range []time.Duration{0, time.Hour} {
		buf := logs(t, slow)
		_, err = db.ExecContext(t.Context(), "-- name: Quick :exec\nSELECT 1")
		require.NoError(t, err)
		require.Empty(t, buf.String())
	}
}

func TestName(t *testing.T) {
	tests := map[string]string{
		"-- name: GetSite :one\nSELECT * FROM sites WHERE id = ?": "GetSite",
		"  -- name: CreateReading :exec\nINSERT INTO readings":    "CreateReading",
		"PRAGMA wal_checkpoint(TRUNCATE)":                         "PRAGMA WAL_CHECKPOINT(TRUNCATE)",
		"CREATE TABLE notes (secret TEXT)":                        "CREATE TABLE",
		"VACUUM":                                                  "VACUUM",
		"-- THE sites\n------\n\nCREATE INDEX idx ON t (a)":       "CREATE INDEX",
		"-- only a comment":                                       "",
		"":                                                        "",
	}
	for query, want := range tests {
		require.Equal(t, want, Name(query), query)
	}
}

// The timing's cost on a prepared query, against the same query without it
func BenchmarkPreparedQuery(b *testing.B) {
	for name, open := range map[string]func() (*sql.DB, error){
		"plain": func() (*sql.DB, error) { return sql.Open("sqlite", ":memory:") },
		"timed": func() (*sql.DB, error) { return Open("benchdb", ":memory:") },
	} {
		b.Run(name, func(b *testing.B) {
			db, err := open()
			require.NoError(b, err)
			defer db.Close()
			stmt, err := db.Prepare("-- name: One :one\nSELECT 1")
			require.NoError(b, err)
			var one int
			for b.Loop() {
				require.NoError(b, stmt.QueryRow().Scan(&one))
			}
		})
	}
}

// sampling is samplePercent for the test, with no stats from before it
func sampling(t *testing.T, percent int) {
	reset := func() {
		stats.Lock()
		stats.byQuery = map[[2]string]*queryStats{}
		stats.Unlock()
	}
	reset()
	SetSamplePercent(percent)
	t.Cleanup(func() { SetSamplePercent(0); reset() })
}

// The queries that took longest in all come first, with how many runs were sampled, their mean, 95th percentile (the
// bucket it is in, at most the longest) and longest
func TestSlowest(t *testing.T) {
	sampling(t, 100)
	for range 20 {
		sample("sitedb", "GetSite", time.Millisecond)
	}
	sample("sitedb", "GetSite", 300*time.Millisecond)
	sample("controldb", "SearchSites", 50*time.Millisecond)
	sample("controldb", "SearchSites", 50*time.Millisecond)

	slowest := Slowest(10)
	require.Equal(t, []QueryStat{
		{DB: "sitedb", Query: "GetSite", Count: 21, TotalMs: 320, MeanMs: 320.0 / 21, P95Ms: 1, MaxMs: 300},
		{DB: "controldb", Query: "SearchSites", Count: 2, TotalMs: 100, MeanMs: 50, P95Ms: 50, MaxMs: 50},
	}, slowest)
	require.Len(t, Slowest(1), 1)
}

// None are kept with sampling off, and every query a DB runs is with it at 100%
func TestSampling(t *testing.T) {
	db, err := Open("testdb", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	query := func() {
		var one int
		require.NoError(t, db.QueryRowContext(t.Context(), "-- name: One :one\nSELECT 1").Scan(&one))
	}

	sampling(t, 0)
	query()
	require.Empty(t, Slowest(10))

	sampling(t, 100)
	query()
	query()
	slowest := Slowest(10)
	require.Len(t, slowest, 1)
	require.Equal(t, "testdb", slowest[0].DB)
	require.Equal(t, "One", slowest[0].Query)
	require.Equal(t, int64(2), slowest[0].Count)
}
