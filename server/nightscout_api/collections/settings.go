package collections

import (
	"context"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

type SettingsCollection struct{}

func (c SettingsCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c SettingsCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c SettingsCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c SettingsCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c SettingsCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c SettingsCollection) DeleteType(ctx context.Context, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}
