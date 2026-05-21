package mappers

import (
	"fmt"
	"math"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout"
)

func DBEntryToNSEntry(dbEntry user.Entry) *nightscout.NSEntry {
	return &nightscout.NSEntry{
		ID:   fmt.Sprintf("%d", dbEntry.ID),
		Date: dbEntry.CreatedAtDate(),
		SGV:  int(math.Round(dbEntry.Value)),
	}
}
