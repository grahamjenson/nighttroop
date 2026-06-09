package collections

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct{}

func (c EntriesCollection) Create(ctx context.Context, data map[string]any) (model.Record, bool, error) {
	return model.Record{}, false, nil
}

func (c EntriesCollection) Get(ctx context.Context, identifier string) (model.Record, error) {
	return model.Record{}, nil
}

func (c EntriesCollection) Search(ctx context.Context, query di.Query) ([]model.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entries, err := userContext.UserDB.Queries.AllEntries(ctx)
	if err != nil {
		return nil, err
	}

	return CreateRecordSlice(entries, SGVToRecord), nil
}
func (c EntriesCollection) Replace(ctx context.Context, collection, identifier string, data map[string]any) (model.Record, bool, error) {
	return model.Record{}, false, nil
}

func (c EntriesCollection) Patch(ctx context.Context, collection, identifier string, patch map[string]any) (model.Record, error) {
	return model.Record{}, nil
}

func (c EntriesCollection) Delete(ctx context.Context, collection, identifier string, permanent bool) error {
	return nil
}

func (c EntriesCollection) DeleteMatching(ctx context.Context, collection string, query di.Query, permanent bool) (int, error) {
	return 0, nil
}

func (c EntriesCollection) History(ctx context.Context, collection string, since int64, limit int) ([]model.Record, error) {
	return nil, nil
}

func (c EntriesCollection) LastModified(ctx context.Context) (int64, error) {
	return 0, nil
}

////
// Models
////

type Entry struct {
	Identifier string          `json:"identifier,omitempty"`
	Type       string          `json:"type,omitempty"`
	SGV        FlexibleFloat64 `json:"sgv,omitempty"`
	MBG        FlexibleFloat64 `json:"mbg,omitempty"`
	Date       int64           `json:"date,omitempty"`
	DateString string          `json:"dateString,omitempty"`
	Device     string          `json:"device,omitempty"`
	Direction  string          `json:"direction,omitempty"`
	Payload    map[string]any  `json:"payload,omitempty"`
}

////
// Mappers
////

func SGVToRecord(sgv *user.Sgv) model.Record {
	if sgv == nil {
		return model.Record{}
	}
	// This id needs to be unique across time and different records
	id := fmt.Sprintf("%d_%d", sgv.ID, sgv.CreatedAt)
	entryModel := Entry{
		Identifier: id,
		Type:       "sgv",
		SGV:        FlexibleFloat64(sgv.Value),
		Date:       sgv.CreatedAt,
		DateString: dateString(sgv.CreatedAt),
		Device:     fmt.Sprintf("%d", sgv.CreatedByIdentityID),
	}

	return model.Record{
		Collection:  "entries",
		ID:          id,
		Data:        StructToMap(entryModel),
		SrvCreated:  sgv.CreatedAt,
		SrvModified: sgv.CreatedAt,
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
