// Package stmtcache prepares each query the first time it runs on a DB and keeps it for the next time, so opening a
// DB does not spend its time preparing every query, most of which it may never run.
package stmtcache

import (
	"context"
	"database/sql"
	"sync"
)

// DB runs sqlc's queries as statements prepared on first use. Only use it for queries whose text never changes, as
// each text is kept; a sqlc.slice query has a text for every length of its list.
type DB struct {
	db      *sql.DB
	stmts   *sync.Map // query text to *sql.Stmt
	tx      *sql.Tx
	txStmts *sync.Map // query text to *sql.Stmt prepared in tx, closed when it ends
}

func New(db *sql.DB) *DB {
	return &DB{db: db, stmts: &sync.Map{}}
}

// WithTx runs the queries in tx, using the statements already prepared on the DB
func (c *DB) WithTx(tx *sql.Tx) *DB {
	return &DB{db: c.db, stmts: c.stmts, tx: tx, txStmts: &sync.Map{}}
}

func (c *DB) stmt(ctx context.Context, query string) (*sql.Stmt, error) {
	if c.tx != nil {
		if cached, ok := c.stmts.Load(query); ok {
			return c.tx.StmtContext(ctx, cached.(*sql.Stmt)), nil
		}
		// The transaction may hold the DB's only connection, so a query the DB has not prepared is prepared in it
		return prepareOnce(c.txStmts, query, func() (*sql.Stmt, error) { return c.tx.PrepareContext(ctx, query) })
	}
	return prepareOnce(c.stmts, query, func() (*sql.Stmt, error) { return c.db.PrepareContext(ctx, query) })
}

func prepareOnce(stmts *sync.Map, query string, prepare func() (*sql.Stmt, error)) (*sql.Stmt, error) {
	if cached, ok := stmts.Load(query); ok {
		return cached.(*sql.Stmt), nil
	}
	stmt, err := prepare()
	if err != nil {
		return nil, err
	}
	// Another query may have prepared it first, so only one is kept
	cached, loaded := stmts.LoadOrStore(query, stmt)
	if loaded {
		stmt.Close()
	}
	return cached.(*sql.Stmt), nil
}

func (c *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	stmt, err := c.stmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

// PrepareContext is not kept, as its caller closes it
func (c *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if c.tx != nil {
		return c.tx.PrepareContext(ctx, query)
	}
	return c.db.PrepareContext(ctx, query)
}

func (c *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	stmt, err := c.stmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (c *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	stmt, err := c.stmt(ctx, query)
	if err != nil {
		// A Row cannot be made with an error, so running it unprepared gives the Row the same error
		if c.tx != nil {
			return c.tx.QueryRowContext(ctx, query, args...)
		}
		return c.db.QueryRowContext(ctx, query, args...)
	}
	return stmt.QueryRowContext(ctx, args...)
}
