package collections

import (
	"context"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type FoodCollection struct {
	converter func(c *user.Food) nsutils.Record
}

func NewFoodCollection() FoodCollection {
	return FoodCollection{
		converter: DBModelToRecord(nsutils.Food, nightscout.DBModelToFood),
	}
}

func (c FoodCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c FoodCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
