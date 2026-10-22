package siteapiutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events/sitemodels"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// Store is a site store. Its methods read and write rows (R) with the queries (Q) of the database they are in,
// and return the database's errors; Rows is the store the handlers use, which shows rows as site models (M),
// answers errors, and publishes each change to the site's own rows. A store without rows of its own, such as
// settings, uses its model as its row.
type Store[R, M, Q any] interface {
	GetName() string
	ToModel(row *R, siteContext *httputils.SiteContext) *M
	GetRow(ctx context.Context, q Q, id string, siteContext *httputils.SiteContext) (*R, error)
}

// Getter, Searcher, Creator, Updater and Deleter are what a Store can do, each with its HTMX response. An
// *apiutils.HTTPError their methods return, such as a value out of range, is answered as it is.
type Getter[M any] interface {
	GetHTMXResponse(model *M) templ.Component
}

type Searcher[R, M, Q any] interface {
	SearchRows(ctx context.Context, q Q, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*R, error)
	SearchHTMXResponse(models []*M, query *apiutils.Query) templ.Component
}

// Creator's id is the one chosen by an upsert, and nil for a create
type Creator[R, M, Q any] interface {
	CreateRow(ctx context.Context, q Q, id *string, siteContext *httputils.SiteContext, model *M) (*R, error)
	CreateHTMXResponse(model *M) templ.Component
}

type Updater[R, M, Q any] interface {
	UpdateRow(ctx context.Context, q Q, siteContext *httputils.SiteContext, oldModel *M, model *M) (*R, error)
	UpdateHTMXResponse(model *M) templ.Component
}

type Deleter[R, M, Q any] interface {
	DeleteRow(ctx context.Context, q Q, id string, siteContext *httputils.SiteContext) error
	DeleteHTMXResponse(model *M) templ.Component
}

// Rows is the store the handlers use
type Rows[R, M, Q any] struct {
	store          Store[R, M, Q]
	queries        func(*StoreTx) (Q, error)
	reportsChanges bool // changes to the site's own rows are reported, to be published on its model events
}

// SiteRows is a store of the site's own rows, whose changes are published
func SiteRows[R, M any](store Store[R, M, *sitedb.Queries]) *Rows[R, M, *sitedb.Queries] {
	return &Rows[R, M, *sitedb.Queries]{store: store, queries: (*StoreTx).Site, reportsChanges: true}
}

// ControlRows is a store of the site's rows in the control database, such as its identities
func ControlRows[R, M any](store Store[R, M, *controldb.Queries]) *Rows[R, M, *controldb.Queries] {
	return &Rows[R, M, *controldb.Queries]{store: store, queries: (*StoreTx).Control}
}

func (s *Rows[R, M, Q]) GetName() string { return s.store.GetName() }

// storeCan is the store as what a handler needs it to do; a store routed to a handler it cannot do fails at start up
func storeCan[I, R, M, Q any](s *Rows[R, M, Q], does string) I {
	i, ok := s.store.(I)
	if !ok {
		panic(fmt.Sprintf("the %s store cannot %s", s.GetName(), does))
	}
	return i
}

// get is the model, or nil when there is none
func (s *Rows[R, M, Q]) get(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*M, *R, *apiutils.HTTPError) {
	q, err := s.queries(tx)
	if err != nil {
		return nil, nil, apiutils.ErrInternalServerError
	}
	row, err := s.store.GetRow(ctx, q, id, siteContext)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil, nil
	}
	if herr := answer(err, s.GetName()); herr != nil {
		return nil, nil, herr
	}
	return s.store.ToModel(row, siteContext), row, nil
}

func (s *Rows[R, M, Q]) search(ctx context.Context, tx *StoreTx, searcher Searcher[R, M, Q], query *apiutils.Query, siteContext *httputils.SiteContext) ([]*M, *apiutils.HTTPError) {
	q, err := s.queries(tx)
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	rows, err := searcher.SearchRows(ctx, q, query, siteContext)
	if herr := answer(err, s.GetName()); herr != nil {
		return nil, herr
	}
	models := make([]*M, 0, len(rows))
	for _, row := range rows {
		models = append(models, s.store.ToModel(row, siteContext))
	}
	return models, nil
}

func (s *Rows[R, M, Q]) create(ctx context.Context, tx *StoreTx, creator Creator[R, M, Q], id *string, siteContext *httputils.SiteContext, model *M) *apiutils.HTTPError {
	q, err := s.queries(tx)
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	row, err := creator.CreateRow(ctx, q, id, siteContext, model)
	if herr := answer(err, s.GetName()); herr != nil {
		return herr
	}
	s.reportChange(tx, sitemodels.Created, row)
	*model = *s.store.ToModel(row, siteContext)
	return nil
}

func (s *Rows[R, M, Q]) update(ctx context.Context, tx *StoreTx, updater Updater[R, M, Q], siteContext *httputils.SiteContext, oldModel *M, model *M) *apiutils.HTTPError {
	q, err := s.queries(tx)
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	row, err := updater.UpdateRow(ctx, q, siteContext, oldModel, model)
	if herr := answer(err, s.GetName()); herr != nil {
		return herr
	}
	s.reportChange(tx, sitemodels.Updated, row)
	*model = *s.store.ToModel(row, siteContext)
	return nil
}

// delete deletes the row read as the model, so the change can say what was deleted
func (s *Rows[R, M, Q]) delete(ctx context.Context, tx *StoreTx, deleter Deleter[R, M, Q], id string, row *R, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := s.queries(tx)
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	if herr := answer(deleter.DeleteRow(ctx, q, id, siteContext), s.GetName()); herr != nil {
		return herr
	}
	s.reportChange(tx, sitemodels.Deleted, row)
	return nil
}

// reportChange reports a change to the site's own rows, to be published once the transaction commits
func (s *Rows[R, M, Q]) reportChange(tx *StoreTx, action sitemodels.Action, row *R) {
	if s.reportsChanges {
		tx.reportChange(action, row)
	}
}

// answer is the response to a store's error: its own, not found for a missing row, or the database's
func answer(err error, what string) *apiutils.HTTPError {
	var herr *apiutils.HTTPError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &herr):
		return herr
	case errors.Is(err, sql.ErrNoRows):
		return apiutils.ErrNotFound
	}
	return apiutils.WriteError(err, what)
}
