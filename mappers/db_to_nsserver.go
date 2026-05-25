package mappers

import (
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
)

func DBEntryToNSEntry(dbEntry user.Sgv) *model.Record {
	return &model.Record{
		Collection:  "entries",
		ID:          string(dbEntry.ID),
		Data:        map[string]any{},
		SrvCreated:  dbEntry.CreatedAt,
		SrvModified: dbEntry.UpdatedAt,
		IsValid:     true,
	}
}
