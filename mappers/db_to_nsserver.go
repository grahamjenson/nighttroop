package mappers

import (
	"fmt"
	"math"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
)

func DBEntryToNSEntry(dbEntry user.Entry) *model.Entry {
	return &model.Entry{
		ID:   fmt.Sprintf("%d", dbEntry.ID),
		Date: dbEntry.CreatedAtDate(),
		SGV:  int(math.Round(dbEntry.Value)),
	}
}
