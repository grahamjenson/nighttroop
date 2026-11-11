package sitedb

import (
	"context"
	"database/sql"
	"sync"
)

// stmtCache runs each of a site's queries as a statement prepared the first time it is run, then reused, so SQLite does
// not parse it again on every run. Only the queries a site runs are prepared, and database/sql prepares a statement on
// each connection that runs it and finalizes it when that connection closes, so they only cost memory while it is busy.
// Its PrepareContext is the DB's, uncached, as the caller closes what it prepares.
type stmtCache struct {
	*sql.DB
	// prepareInTx is whether a query first run in a transaction can be prepared on another connection, unlike in memory
	prepareInTx bool
	stmts       sync.Map // query to *sql.Stmt
}

// stmt is query's statement, prepared on its first run. One that fails to prepare is not kept, so it is tried again.
func (c *stmtCache) stmt(ctx context.Context, query string) (*sql.Stmt, error) {
	if s, ok := c.stmts.Load(query); ok {
		return s.(*sql.Stmt), nil
	}
	s, err := c.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if first, loaded := c.stmts.LoadOrStore(query, s); loaded {
		s.Close() // another run prepared it at the same time
		return first.(*sql.Stmt), nil
	}
	return s, nil
}

func (c *stmtCache) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	s, err := c.stmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return s.ExecContext(ctx, args...)
}

func (c *stmtCache) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	s, err := c.stmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return s.QueryContext(ctx, args...)
}

func (c *stmtCache) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	s, err := c.stmt(ctx, query)
	if err != nil {
		return c.DB.QueryRowContext(ctx, query, args...) // a sql.Row cannot be made with err, so running it gives it
	}
	return s.QueryRowContext(ctx, args...)
}

// Close closes every statement prepared
func (c *stmtCache) Close() {
	if c == nil {
		return
	}
	c.stmts.Range(func(query, s any) bool {
		s.(*sql.Stmt).Close()
		c.stmts.Delete(query)
		return true
	})
}

// txStmts runs a site's queries in a transaction, each with its cached statement on the transaction's connection
type txStmts struct {
	*sql.Tx
	cache *stmtCache
}

// stmt is query's statement on the transaction's connection, or nil to run it unprepared, giving any error preparing it
func (t *txStmts) stmt(ctx context.Context, query string) *sql.Stmt {
	if s, ok := t.cache.stmts.Load(query); ok {
		return t.Tx.StmtContext(ctx, s.(*sql.Stmt))
	}
	if !t.cache.prepareInTx {
		return nil
	}
	// Prepared on another of the DB's connections, as the transaction holds its own (a file's DB opens as many as it
	// needs), then again on the transaction's by StmtContext, which keeps it for that connection's later transactions
	s, err := t.cache.stmt(ctx, query)
	if err != nil {
		return nil
	}
	return t.Tx.StmtContext(ctx, s)
}

func (t *txStmts) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	s := t.stmt(ctx, query)
	if s == nil {
		return t.Tx.ExecContext(ctx, query, args...)
	}
	return s.ExecContext(ctx, args...)
}

func (t *txStmts) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	s := t.stmt(ctx, query)
	if s == nil {
		return t.Tx.QueryContext(ctx, query, args...)
	}
	return s.QueryContext(ctx, args...)
}

func (t *txStmts) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	s := t.stmt(ctx, query)
	if s == nil {
		return t.Tx.QueryRowContext(ctx, query, args...)
	}
	return s.QueryRowContext(ctx, args...)
}
