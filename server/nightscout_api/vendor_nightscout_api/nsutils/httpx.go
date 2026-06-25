package nsutils

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func WriteRecordV3(w http.ResponseWriter, record Record, fields []string) {
	body := map[string]any{"status": http.StatusOK, "result": record.SelectFields(fields, true)}
	if record.Identifier() != "" {
		LastModifiedHeader(w, record.SrvModified)
		body["status"] = http.StatusOK
		body["identifier"] = record.Identifier()
		body["lastModified"] = record.SrvModified

	}
	WriteJSON(w, http.StatusOK, body)
}

func WriteRecordsV3(w http.ResponseWriter, records []Record, fields []string) {
	result := make([]map[string]any, 0, len(records))
	for _, record := range records {
		result = append(result, record.SelectFields(fields, true))
	}
	WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": result})
}

func WriteRecordsV1(w http.ResponseWriter, records []Record, fields []string) {
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		response = append(response, record.SelectFields(fields, false))
	}
	WriteJSON(w, http.StatusOK, response)
}

// WriteJSON writes a JSON response with the provided HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
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

// LastModifiedHeader writes a Last-Modified header for a Unix millisecond
// timestamp when the value is positive.
func LastModifiedHeader(w http.ResponseWriter, millis int64) {
	if millis <= 0 {
		return
	}
	w.Header().Set("Last-Modified", time.UnixMilli(millis).UTC().Format(http.TimeFormat))
}
