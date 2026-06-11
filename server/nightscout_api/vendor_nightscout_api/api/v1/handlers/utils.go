package handlers

import (
	"time"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// ensureV1Timestamp fills in date/created_at when missing. Nightscout v1
// clients (Trio, xDrip, etc.) routinely POST treatments, devicestatus, and
// food without a timestamp and expect the server to default to now.
func ensureV1Timestamp(doc map[string]any) {
	if doc == nil {
		return
	}
	hasDate := false
	if _, ok := doc["date"]; ok && doc["date"] != nil {
		hasDate = true
	}
	if value, ok := nsutils.StringField(doc, "created_at"); ok && value != "" {
		hasDate = true
	}
	if hasDate {
		return
	}
	now := time.Now().UTC()
	doc["created_at"] = now.Format("2006-01-02T15:04:05.000Z")
	doc["date"] = now.UnixMilli()
}
