package jsonapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestRunCarriesSetsAndRunsBackgroundJobs(t *testing.T) {
	fsys := fstest.MapFS{"api/echo.http": {Data: []byte(`NAME first
GET http://localhost/echo?say=hello
ASSERT_STATUS 200
SET said {{ .first.said }}

RUN_BACKGROUND_JOBS

NAME second
GET http://localhost/echo?say={{ .said }}-again
ASSERT_STATUS 200
ASSERT_EQUALS "hello-again" "@said"
`)}}
	// The handler answers with what it was asked to say
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"said": r.URL.Query().Get("say")})
	})

	file, err := ParseRequestFile(fsys, "api/echo.http")
	require.NoError(t, err)
	ran := 0
	ctx := map[string]any{}
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	file.Run(t, Client(ts), ctx, func() { ran++ })
	require.Equal(t, 1, ran)
	require.Equal(t, map[string]any{"said": "hello-again"}, ctx["second"])
}

// A request is sent over the network to the test server whatever its host, and its response is the one the server
// sent: a redirect is not followed, and the body is not compressed
func TestRunSendsAsWritten(t *testing.T) {
	fsys := fstest.MapFS{"api/redirect.http": {Data: []byte(`NAME moved
GET http://abcdefgh.ns.localhost/old
ASSERT_STATUS 302
ASSERT_EQUALS "/new" "#Location.0"
ASSERT_EQUALS "abcdefgh.ns.localhost" "@host"
`)}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("Accept-Encoding"))
		w.Header().Set("Location", "/new")
		w.WriteHeader(http.StatusFound)
		json.NewEncoder(w).Encode(map[string]string{"host": r.Host})
	}))
	t.Cleanup(ts.Close)
	file, err := ParseRequestFile(fsys, "api/redirect.http")
	require.NoError(t, err)
	file.Run(t, Client(ts), map[string]any{}, func() {})
}
