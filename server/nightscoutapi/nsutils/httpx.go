package nsutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/utils/apiutils"
)

type RecordWriteType int

const (
	Created RecordWriteType = iota
	Replaced
	FoundDuplicate
	Found
)

func WriteRecordV3(w http.ResponseWriter, record Document, fields []string, createdType RecordWriteType) {
	status := http.StatusOK
	if createdType == Created {
		status = http.StatusCreated
	}
	body := map[string]any{"status": status, "result": record.JSON(true, fields)}
	if record.Identifier() != "" {
		body["identifier"] = record.Identifier()
		body["lastModified"] = record.Modified()
		if createdType == FoundDuplicate {
			body["isDeduplication"] = true
			body["deduplicatedIdentifier"] = record.Identifier()
		}
	}

	if record.Modified() > 0 {
		w.Header().Set("Last-Modified", time.UnixMilli(record.Modified()).UTC().Format(http.TimeFormat))
		w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, record.Modified()))
	}

	apiutils.WriteJSON(w, status, body)
}

func WriteRecordsV3(w http.ResponseWriter, records []Document, fields []string) {
	if len(records) == 0 {
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": []map[string]any{}})
		return
	}

	result := make([]json.RawMessage, 0, len(records))
	for _, record := range records {
		result = append(result, record.JSON(true, fields))
	}

	writeHeadersForRecords(w, records)

	apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": result})
}

func WriteRecordsV1(w http.ResponseWriter, records []Document, fields []string) {
	response := make([]json.RawMessage, 0, len(records))
	for _, record := range records {
		response = append(response, record.JSON(false, fields))
	}
	writeHeadersForRecords(w, records)
	apiutils.WriteJSON(w, http.StatusOK, response)
}

func WriteRecordV1(w http.ResponseWriter, record Document, fields []string) {
	response := record.JSON(false, fields)

	w.Header().Set("Last-Modified", time.UnixMilli(record.Modified()).UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, record.Modified()))

	apiutils.WriteJSON(w, http.StatusOK, response)
}

func writeHeadersForRecords(w http.ResponseWriter, records []Document) {
	// Adding Last-Modified and ETag headers for the newest record in the response
	if len(records) == 0 {
		return
	}
	// Adding Last-Modified and ETag headers for the newest record in the response
	var maxSrvModified int64
	for _, record := range records {
		if record.Modified() > maxSrvModified {
			maxSrvModified = record.Modified()
		}
	}

	w.Header().Set("Last-Modified", time.UnixMilli(maxSrvModified).UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, maxSrvModified))
}

// ParseIfUnmodifiedSince parses the If-Unmodified-Since header when present.
func ParseIfUnmodifiedSince(r *http.Request) (time.Time, bool) {
	value := strings.TrimSpace(r.Header.Get("If-Unmodified-Since"))
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(http.TimeFormat, value)
	return parsed, err == nil
}

type unmodifiedSinceKey struct{}

// WithIfUnmodifiedSince is the request's context with its If-Unmodified-Since, for a replace or patch to check against
// the record in the same transaction that changes it, so no change can come between
func WithIfUnmodifiedSince(r *http.Request) context.Context {
	if since, ok := ParseIfUnmodifiedSince(r); ok {
		return context.WithValue(r.Context(), unmodifiedSinceKey{}, since)
	}
	return r.Context()
}

// ModifiedSince is whether a record last modified at modified, in milliseconds, changed after the context's
// If-Unmodified-Since. It compares whole seconds, as the header has no finer, so a client sending back the
// Last-Modified it was given is not refused for the milliseconds it cannot send.
func ModifiedSince(ctx context.Context, modified int64) bool {
	since, ok := ctx.Value(unmodifiedSinceKey{}).(time.Time)
	return ok && modified/1000 > since.Unix()
}
