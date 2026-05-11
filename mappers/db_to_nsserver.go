package mappers

import (
	"fmt"

	"nighttroop.t1d.nz/data/sql/user/entry"
	"nighttroop.t1d.nz/server/nightscout"
)

func DBEntryToNSEntry(dbEntry entry.IEntry) *nightscout.NSEntry {
	return &nightscout.NSEntry{
		ID:        fmt.Sprintf("%d", dbEntry.Id()),
		Date:      dbEntry.CreatedAt(),
		SGV:       int(dbEntry.Value()),
		Type:      dbEntry.CreatedByType(),
		Direction: "",
	}
}
