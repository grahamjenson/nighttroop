package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
)

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

// RequireRecord maps store lookup errors to the corresponding HTTP response.
func RequireRecord(storeErr error) (status int, body map[string]any) {
	switch {
	case errors.Is(storeErr, di.ErrGone):
		return http.StatusGone, map[string]any{"status": http.StatusGone}
	case errors.Is(storeErr, di.ErrNotFound):
		return http.StatusNotFound, map[string]any{"status": http.StatusNotFound}
	default:
		return http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": storeErr.Error()}
	}
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
