package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func JSONRequest[T any](t *testing.T, handler http.Handler, method, path_or_url string, body any) (int, T, http.Header) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode json: %v", err)
		}
	}

	// If PATH then host is exmaple.com, if URL then host is from URL
	req := httptest.NewRequest(method, path_or_url, &buf)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var out T
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode json: %v\nbody: %s", err, rr.Body.String())
	}

	return rr.Code, out, rr.Header()
}
