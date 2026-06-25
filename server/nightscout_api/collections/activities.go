package collections

import (
	"context"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ActivityCollection struct {
	converter func(c *user.Activity) nsutils.Record
}

func NewActivityCollection() ActivityCollection {
	return ActivityCollection{
		converter: DBModelToRecord(nsutils.Activity, nightscout.DBModelToActivity),
	}
}

func (c ActivityCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c ActivityCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
