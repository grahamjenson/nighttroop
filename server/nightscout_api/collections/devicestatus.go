package collections

import (
	"context"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

type DeviceStatusCollection struct{}

func (c DeviceStatusCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, bool, error) {
	return nil, false, nil
}

func (c DeviceStatusCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, error) {
	return nil, nil
}

func (c DeviceStatusCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, error) {
	return nil, nil
}

func (c DeviceStatusCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, error) {
	return nil, nil
}

func (c DeviceStatusCollection) Delete(ctx context.Context, identifier string, permanent bool) error {
	return nil
}

func (c DeviceStatusCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, error) {
	return 0, nil
}

////
// Models
////

type DeviceStatus struct {
	Identifier string         `json:"identifier,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
	Date       int64          `json:"date,omitempty"`
	UTCOffset  int            `json:"utcOffset,omitempty"`
	Device     string         `json:"device,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}
