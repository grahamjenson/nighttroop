package nsutils

import "time"

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
}

// DateString is when the document happened, as Nightscout's dateString
func DateString(doc Document) string {
	return time.UnixMilli(doc.OccurredAt()).UTC().Format(time.RFC3339Nano)
}

// SelectFields projects a document to the requested fields.
func SelectFields(doc Document, fields []string, version3 bool) map[string]any {
	full := doc.Fields(version3)
	if len(fields) == 0 {
		return full
	}
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := full[field]; ok {
			out[field] = value
		}
	}
	if len(fields) > 0 {
		delete(out, "_id")
	}
	return out
}
