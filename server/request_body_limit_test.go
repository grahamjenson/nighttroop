package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// A request body larger than MAX_REQUEST_BODY_BYTES is never read in full: refused with 413 when it says how large it
// is, and cut off, failing its decode, when it does not
func TestRequestBodyLimit(t *testing.T) {
	testConfig, controlDB, site := testutils.NewControlDBWithSite(t)
	config := *testConfig
	config.MaxRequestBodyBytes = 1000
	handler := GetMainHandler(controlDB, &config)

	entries := func(count int) string {
		entry := `{"type": "sgv", "sgv": 120, "date": %d, "device": "test"}`
		list := []string{}
		for i := range count {
			list = append(list, fmt.Sprintf(entry, 1700000000000+int64(i)*300000))
		}
		return "[" + strings.Join(list, ",") + "]"
	}
	small, large := entries(2), entries(50)
	require.Less(t, len(small), 1000)
	require.Greater(t, len(large), 1000)

	send := func(body string, sayLength bool) int {
		var reader io.Reader = strings.NewReader(body)
		if !sayLength {
			// Hides the length, as a chunked upload does
			reader = io.MultiReader(reader)
		}
		r := httptest.NewRequest("POST", "http://"+site.SiteDB.Subdomain+".ns.localhost/api/v1/entries?token="+site.PwdNSToken, reader)
		if !sayLength {
			r.ContentLength = -1
		}
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}

	tests := []struct {
		name      string
		body      string
		sayLength bool
		status    int
	}{
		{name: "small", body: small, sayLength: true, status: http.StatusOK},
		{name: "small without a length", body: small, sayLength: false, status: http.StatusOK},
		{name: "large", body: large, sayLength: true, status: http.StatusRequestEntityTooLarge},
		{name: "large without a length", body: large, sayLength: false, status: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.status, send(tt.body, tt.sayLength))
		})
	}
}
