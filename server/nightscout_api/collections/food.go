package collections

import (
	"context"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

type FoodCollection struct{}

func (c FoodCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c FoodCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c FoodCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
