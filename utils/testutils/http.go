package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
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

// HostClient sends every request to the test server, whatever its host, so the site is chosen by the Host
// header as in production
func HostClient(ts *httptest.Server) *http.Client {
	addr := ts.Listener.Addr().String()
	return &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}}}
}
