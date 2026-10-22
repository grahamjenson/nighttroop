package siteapiutils

import (
	"context"
	"database/sql"
	"errors"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// RowStore is a store of one table in the site's database. Its methods read and write the table's sitedb rows (R)
// and return the database's errors; Rows makes it a store for the handlers, which shows rows as site models (M),
// answers errors, and publishes each change on the site's model events.
type RowStore[R, M any] interface {
	GetName() string
	ToModel(row *R, siteContext *httputils.SiteContext) *M
	GetRow(ctx context.Context, q *sitedb.Queries, id string) (*R, error)
	SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query) ([]*R, error)
	GetHTMXResponse(model *M) templ.Component
	SearchHTMXResponse(models []*M, query *apiutils.Query) templ.Component
}

// RowCreator, RowUpdater and RowDeleter are the writes a RowStore can do. An *apiutils.HTTPError they return,
// such as a value out of range, is answered as it is.
type RowCreator[R, M any] interface {
	CreateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, model *M) (*R, error)
	CreateHTMXResponse(model *M) templ.Component
}

type RowUpdater[R, M any] interface {
	UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *M, model *M) (*R, error)
	UpdateHTMXResponse(model *M) templ.Component
}

type RowDeleter[R, M any] interface {
	DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error
	DeleteHTMXResponse(model *M) templ.Component
}

// Rows is the handlers' store for a RowStore
func Rows[R, M any](store RowStore[R, M]) *RowsStore[R, M] {
	return &RowsStore[R, M]{store: store}
}

type RowsStore[R, M any] struct {
	store RowStore[R, M]
}

func (s *RowsStore[R, M]) GetName() string { return s.store.GetName() }

func (s *RowsStore[R, M]) toModel(siteContext *httputils.SiteContext) func(*R) *M {
	return func(row *R) *M { return s.store.ToModel(row, siteContext) }
}

func (s *RowsStore[R, M]) GetModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*M, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	row, err := s.store.GetRow(ctx, q, id)
	return apiutils.Found(row, err, s.GetName(), s.toModel(siteContext))
}

func (s *RowsStore[R, M]) SearchModels(ctx context.Context, tx *StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*M, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	rows, err := s.store.SearchRows(ctx, q, query)
	return apiutils.Searched(rows, err, s.GetName(), s.toModel(siteContext))
}

func (s *RowsStore[R, M]) CreateModel(ctx context.Context, tx *StoreTx, id *string, siteContext *httputils.SiteContext, model *M) (bool, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	row, err := s.store.(RowCreator[R, M]).CreateRow(ctx, q, siteContext, model)
	if herr := writeError(err, s.GetName()); herr != nil {
		return false, herr
	}
	tx.Created(row)
	*model = *s.store.ToModel(row, siteContext)
	return true, nil
}

func (s *RowsStore[R, M]) UpdateModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext, oldModel *M, model *M) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	row, err := s.store.(RowUpdater[R, M]).UpdateRow(ctx, q, siteContext, oldModel, model)
	if herr := writeError(err, s.GetName()); herr != nil {
		return herr
	}
	tx.Updated(row)
	*model = *s.store.ToModel(row, siteContext)
	return nil
}

func (s *RowsStore[R, M]) DeleteModel(ctx context.Context, tx *StoreTx, id string, model *M, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	// Read first, so the change can say what was deleted
	row, err := s.store.GetRow(ctx, q, id)
	if herr := writeError(err, s.GetName()); herr != nil {
		return herr
	}
	if herr := writeError(s.store.(RowDeleter[R, M]).DeleteRow(ctx, q, id, siteContext), s.GetName()); herr != nil {
		return herr
	}
	tx.Deleted(row)
	return nil
}

func (s *RowsStore[R, M]) CreateHTMXResponse(model *M) templ.Component {
	return s.store.(RowCreator[R, M]).CreateHTMXResponse(model)
}

func (s *RowsStore[R, M]) UpdateHTMXResponse(model *M) templ.Component {
	return s.store.(RowUpdater[R, M]).UpdateHTMXResponse(model)
}

func (s *RowsStore[R, M]) DeleteHTMXResponse(model *M) templ.Component {
	return s.store.(RowDeleter[R, M]).DeleteHTMXResponse(model)
}

func (s *RowsStore[R, M]) GetHTMXResponse(model *M) templ.Component { return s.store.GetHTMXResponse(model) }

func (s *RowsStore[R, M]) SearchHTMXResponse(models []*M, query *apiutils.Query) templ.Component {
	return s.store.SearchHTMXResponse(models, query)
}

// writeError is the answer to a write's error: the store's own, not found for a missing row, or the database's
func writeError(err error, what string) *apiutils.HTTPError {
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
