package sqltiming

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	createNotes = "-- name: CreateNotes :exec\nCREATE TABLE notes (secret TEXT)"
	addNote     = "-- name: AddNote :exec\nINSERT INTO notes (secret) VALUES (?)"
	searchNotes = "-- name: SearchNotes :many\nSELECT secret FROM notes ORDER BY secret"
)

// statements are the sqlc queries prepared on db's one connection
func statements(t *testing.T, db *sql.DB) map[string]*cachedStmt {
	conn, err := db.Conn(t.Context())
	require.NoError(t, err)
	defer conn.Close()
	var stmts map[string]*cachedStmt
	require.NoError(t, conn.Raw(func(dc any) error { stmts = dc.(*timedConn).stmts; return nil }))
	return stmts
}

// notes are the notes searchNotes gives, from q
func notes(t *testing.T, ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) []string {
	rows, err := q.QueryContext(ctx, searchNotes)
	require.NoError(t, err)
	defer rows.Close()
	found := []string{}
	for rows.Next() {
		var note string
		require.NoError(t, rows.Scan(&note))
		found = append(found, note)
	}
	require.NoError(t, rows.Err())
	return found
}

// An sqlc query is prepared on the connection that first runs it, in a transaction or not, then reused, so a
// transaction needs no other connection to prepare one; and it is finalized when the connection closes
func TestQueriesArePreparedOnTheirConnection(t *testing.T) {
	ctx := t.Context()
	db, err := Open("testdb", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1) // a transaction holds the only connection, so it cannot prepare on another
	_, err = db.ExecContext(ctx, createNotes)
	require.NoError(t, err)
	stmts := statements(t, db) // the map the connection keeps them in, read while it is not in a transaction

	// Prepared in a transaction, then reused outside it
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, addNote, "b")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	added := stmts[addNote]
	require.NotNil(t, added)
	_, err = db.ExecContext(ctx, addNote, "a")
	require.NoError(t, err)
	require.Same(t, added, stmts[addNote])

	// A query run again on the connection while its rows are open, as in a loop over them, runs unprepared, and
	// both give every row
	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	rows, err := tx.QueryContext(ctx, searchNotes)
	require.NoError(t, err)
	for rows.Next() {
		require.Equal(t, []string{"a", "b"}, notes(t, ctx, tx))
	}
	require.NoError(t, rows.Close())
	require.Equal(t, []string{"a", "b"}, notes(t, ctx, tx))
	require.NoError(t, tx.Commit())
	require.Len(t, stmts, 3, "the query run again is not prepared again")

	// Nor is a query that is not sqlc's, nor one that fails to prepare, which gives its error
	_, err = db.ExecContext(ctx, "INSERT INTO notes (secret) VALUES ('c')")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "-- name: Bad :exec\nINSERT INTO nowhere VALUES (1)")
	require.ErrorContains(t, err, "no such table")
	_, err = db.QueryContext(ctx, "-- name: Bad :many\nSELECT * FROM nowhere")
	require.ErrorContains(t, err, "no such table")
	require.Len(t, stmts, 3)

	// Closing the connection finalizes them
	require.NoError(t, db.Close())
	require.Empty(t, stmts)
}
