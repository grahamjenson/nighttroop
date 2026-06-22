package collections

import (
	"context"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

type ActivityCollection struct{}

func (c ActivityCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c ActivityCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c ActivityCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
