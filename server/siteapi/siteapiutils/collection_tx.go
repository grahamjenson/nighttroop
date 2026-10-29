package siteapiutils

import (
	"context"
	"database/sql"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// CollectionTx gives a store one transaction per database it writes to, started on first use.
// Each DB has a single connection, so while a transaction is open every other query on that
// DB waits for it: do network calls and plain Queries reads before the first Control() or Site().
// The two transactions commit separately, so they are atomic per DB, not across both.
type CollectionTx struct {
	ctx         context.Context
	siteContext *httputils.SiteContext
	controlTx   *sql.Tx
	siteTx      *sql.Tx
	// reportedChanges are the changes to the site's models made in the transaction, published once it commits
	reportedChanges []events.Event
}

func NewCollectionTx(ctx context.Context, siteContext *httputils.SiteContext) *CollectionTx {
	return &CollectionTx{ctx: ctx, siteContext: siteContext}
}

func (t *CollectionTx) Control() (*controldb.Queries, error) {
	if t.controlTx == nil {
		tx, err := t.siteContext.ControlDB.SQLiteDB.BeginTx(t.ctx, nil)
		if err != nil {
			return nil, err
		}
		t.controlTx = tx
	}
	return t.siteContext.ControlDB.Queries.WithTx(t.controlTx), nil
}

func (t *CollectionTx) Site() (*sitedb.Queries, error) {
	if t.siteTx == nil {
		tx, err := t.siteContext.SiteDB.SQLiteDB.BeginTx(t.ctx, nil)
		if err != nil {
			return nil, err
		}
		t.siteTx = tx
	}
	return t.siteContext.SiteDB.Queries.WithTx(t.siteTx), nil
}

// SiteQuery runs one query, such as (*sitedb.Queries).GetInsulinDose, with its params in the transaction's site DB
func SiteQuery[P, R any](ctx context.Context, t *CollectionTx, query func(*sitedb.Queries, context.Context, P) (R, error), params P) (R, error) {
	q, err := t.Site()
	if err != nil {
		var none R
		return none, err
	}
	return query(q, ctx, params)
}

// Live is the live DB's queries, which are not in the transaction, so a rollback does not undo them
func (t *CollectionTx) Live() *livedb.Queries {
	return t.siteContext.ControlDB.Live.Queries
}

// Finish commits on success, otherwise rolls back
func (t *CollectionTx) Finish(herr *apiutils.HTTPError) *apiutils.HTTPError {
	if herr != nil {
		t.Rollback()
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
		return apiutils.ErrInternalServerError
	}
	for _, change := range t.reportedChanges {
		t.siteContext.PublishModelChange(change.Action, change.Model)
	}
	return nil
}

// reportChange reports a change to one of the site's models, its sitedb row, to be published once committed
func (t *CollectionTx) ReportChange(action events.Action, row any) {
	t.reportedChanges = append(t.reportedChanges, events.Event{Action: action, Model: row})
}

// Rollback undoes anything not yet committed. It is safe to defer: after Finish it does nothing.
func (t *CollectionTx) Rollback() {
	for _, tx := range []*sql.Tx{t.siteTx, t.controlTx} {
		if tx != nil {
			_ = tx.Rollback()
		}
	}
}
