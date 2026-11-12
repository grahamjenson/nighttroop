package nsutils

import (
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/utils/types"
)

// Document is one stored record as the Nightscout API writes it, whatever collection it is from
type Document interface {
	// Identifier is its public ID
	Identifier() string
	// Modified is when the server last changed it, in milliseconds, for Last-Modified and API v3's lastModified
	Modified() int64
	// OccurredAt is when it happened, in milliseconds
	OccurredAt() int64
	// Fields is it as JSON fields: API v3's with the server's fields, or v1's with an _id
	Fields(v3 bool) map[string]any
	// JSON is Fields as JSON, with only the fields given less _id when there are any
	JSON(v3 bool, fields []string) json.RawMessage
}

// DateString is when the document happened, as Nightscout's dateString
func DateString(doc Document) string {
	return types.ISOString(time.UnixMilli(doc.OccurredAt()))
}
