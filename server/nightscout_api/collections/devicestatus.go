package collections

import (
	"context"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

type DeviceStatusCollection struct{}

func (c DeviceStatusCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c DeviceStatusCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c DeviceStatusCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c DeviceStatusCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c DeviceStatusCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c DeviceStatusCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
