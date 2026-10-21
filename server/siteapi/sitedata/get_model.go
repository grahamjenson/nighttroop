package sitedata

import (
	"context"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// getModel runs a site Get query in the store's transaction and converts the row, nil if not found
func getModel[P, R, M any](ctx context.Context, tx *siteapiutils.StoreTx, get func(*sitedb.Queries, context.Context, P) (*R, error), params P, what string, toModel func(*R) *M) (*M, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	row, err := get(q, ctx, params)
	return apiutils.Found(row, err, what, toModel)
}
