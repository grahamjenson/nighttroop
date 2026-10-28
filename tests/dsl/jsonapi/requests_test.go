package jsonapi

import (
	"encoding/json"
	"net/http"
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
	file.Run(t, handler, ctx, func() { ran++ })
	require.Equal(t, 1, ran)
	require.Equal(t, map[string]any{"said": "hello-again"}, ctx["second"])
}
