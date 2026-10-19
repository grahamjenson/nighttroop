package siteapiutils

import (
	"context"
	"database/sql"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/httputils"
)

// StoreTx gives a store one transaction per database it writes to, started on first use.
// Each DB has a single connection, so while a transaction is open every other query on that
// DB waits for it: do network calls and plain Queries reads before the first Control() or Site().
// The two transactions commit separately, so they are atomic per DB, not across both.
type StoreTx struct {
	ctx         context.Context
	siteContext *httputils.SiteContext
	controlTx   *sql.Tx
	siteTx      *sql.Tx
	afterCommit []func(ctx context.Context) *HTTPError
}

func NewStoreTx(ctx context.Context, siteContext *httputils.SiteContext) *StoreTx {
	return &StoreTx{ctx: ctx, siteContext: siteContext}
}

func (t *StoreTx) Control() (*controldb.Queries, error) {
	if t.controlTx == nil {
		tx, err := t.siteContext.ControlDB.SQLiteDB.BeginTx(t.ctx, nil)
		if err != nil {
			return nil, err
		}
		t.controlTx = tx
	}
	return t.siteContext.ControlDB.Queries.WithTx(t.controlTx), nil
}

func (t *StoreTx) Site() (*sitedb.Queries, error) {
	if t.siteTx == nil {
		tx, err := t.siteContext.SiteDB.SQLiteDB.BeginTx(t.ctx, nil)
		if err != nil {
			return nil, err
		}
		t.siteTx = tx
	}
	return t.siteContext.SiteDB.Queries.WithTx(t.siteTx), nil
}

// AfterCommit runs fn once both transactions have committed, for work that uses the DBs directly
func (t *StoreTx) AfterCommit(fn func(ctx context.Context) *HTTPError) {
	t.afterCommit = append(t.afterCommit, fn)
}

// Finish commits on success, otherwise rolls back, then runs the AfterCommit hooks
func (t *StoreTx) Finish(herr *HTTPError) *HTTPError {
	if herr != nil {
		t.rollback()
		return herr
	}
	var err error
	if t.siteTx != nil {
		err = t.siteTx.Commit()
	}
	if t.controlTx != nil {
		if err != nil {
			_ = t.controlTx.Rollback()
		} else {
			err = t.controlTx.Commit()
		}
	}
	if err != nil {
		return ErrInternalServerError
	}
	for _, fn := range t.afterCommit {
		if herr := fn(t.ctx); herr != nil {
			return herr
		}
	}
	return nil
}

func (t *StoreTx) rollback() {
	for _, tx := range []*sql.Tx{t.siteTx, t.controlTx} {
		if tx != nil {
			_ = tx.Rollback()
		}
	}
}
