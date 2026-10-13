package nsutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RecordWriteType int

const (
	Created RecordWriteType = iota
	Replaced
	FoundDuplicate
	Found
)

func WriteRecordV3(w http.ResponseWriter, record Record, fields []string, createdType RecordWriteType) {
	status := http.StatusOK
	if createdType == Created {
		status = http.StatusCreated
	}
	body := map[string]any{"status": status, "result": record.SelectFields(fields, true)}
	if record.Identifier() != "" {
		body["identifier"] = record.Identifier()
		body["lastModified"] = record.SrvModified
		if createdType == FoundDuplicate {
			body["isDeduplication"] = true
			body["deduplicatedIdentifier"] = record.Identifier()
		}
	}

	if record.SrvModified > 0 {
		w.Header().Set("Last-Modified", time.UnixMilli(record.SrvModified).UTC().Format(http.TimeFormat))
		w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, record.SrvModified))
	}

	WriteJSON(w, status, body)
}

func WriteRecordsV3(w http.ResponseWriter, records []Record, fields []string) {
	if len(records) == 0 {
		WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": []map[string]any{}})
		return
	}

	result := make([]map[string]any, 0, len(records))
	for _, record := range records {
		result = append(result, record.SelectFields(fields, true))
	}

	writeHeadersForRecords(w, records)

	WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": result})
}

func WriteRecordsV1(w http.ResponseWriter, records []Record, fields []string) {
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		response = append(response, record.SelectFields(fields, false))
	}
	writeHeadersForRecords(w, records)
	WriteJSON(w, http.StatusOK, response)
}

func WriteRecordV1(w http.ResponseWriter, record Record, fields []string) {
	response := record.SelectFields(fields, false)

	w.Header().Set("Last-Modified", time.UnixMilli(record.SrvModified).UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, record.SrvModified))

	WriteJSON(w, http.StatusOK, response)
}

// WriteJSON writes a JSON response with the provided HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(body)
}

func writeHeadersForRecords(w http.ResponseWriter, records []Record) {
	// Adding Last-Modified and ETag headers for the newest record in the response
	if len(records) == 0 {
		return
	}
	// Adding Last-Modified and ETag headers for the newest record in the response
	var maxSrvModified int64
	for _, record := range records {
		if record.SrvModified > maxSrvModified {
			maxSrvModified = record.SrvModified
		}
	}

	w.Header().Set("Last-Modified", time.UnixMilli(maxSrvModified).UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf(`W/"%d"`, maxSrvModified))
}

// ReadJSON decodes a request body as JSON while preserving number precision via
// json.Number.
func ReadJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	return decoder.Decode(dst)
}

// ParseIfModifiedSince parses the If-Modified-Since header when present.
func ParseIfModifiedSince(r *http.Request) (time.Time, bool) {
	value := strings.TrimSpace(r.Header.Get("If-Modified-Since"))
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(http.TimeFormat, value)
	return parsed, err == nil
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
