package flow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/tests/dsl/browser"
)

// A .http file's NAME blocks are API steps, an INCLUDE's SET goes on its last step, and browser lines in between
// are browser steps, each with the asserts after it
func TestParseMixes(t *testing.T) {
	fsys := fstest.MapFS{
		"steps/site.steps": {Data: []byte("NAME create_site\nPOST http://localhost/site\nASSERT_STATUS 200\nSET token {{ .create_site.token }}\n")},
		"api/list.http":    {Data: []byte("INCLUDE steps/site.steps\nSET site {{ .create_site.subdomain }}\n\nNAME list\nGET http://localhost/list?token={{ token }}\nASSERT_STATUS 200\n")},
		"api/missing.http": {Data: []byte("NAME list\nGET http://localhost/list\nINCLUDE steps/missing.steps\n")},
		"site/mixed.flow": {Data: []byte(`INCLUDE steps/site.steps
GOTO {{ server_url }}/
ASSERT_VISIBLE #site--login
POST /settings/push-registrations {"app": "loop"}
NAME connect
CONNECT http://{{ site }}/socket.io/?EIO=3
NAME updated
WAIT dataUpdate
CLICK #site--nav--home
ASSERT_PATH /
`)},
	}

	list, err := Parse(fsys, "api/list.http")
	require.NoError(t, err)
	require.Len(t, list.Steps, 2)
	require.Equal(t, "create_site", list.Steps[0].API.Name)
	require.Len(t, list.Steps[0].API.Sets, 2)
	require.Equal(t, "list", list.Steps[1].API.Name)
	require.Equal(t, "steps/site.steps", list.Steps[0].API.Path)

	mixed, err := Parse(fsys, "site/mixed.flow")
	require.NoError(t, err)
	kinds := []string{}
	for _, step := range mixed.Steps {
		if step.API != nil {
			kinds = append(kinds, step.API.Verb)
		} else {
			kinds = append(kinds, step.Browser.ToFlatString())
		}
	}
	require.Equal(t, []string{"POST", "GOTO {{ server_url }}/", `POST /settings/push-registrations {"app": "loop"}`, "CONNECT", "WAIT",
		"CLICK #site--nav--home"}, kinds)
	require.IsType(t, &browser.Goto{}, mixed.Steps[1].Browser)
	require.Equal(t, []string{"dataUpdate"}, mixed.API.Events())

	_, err = Parse(fsys, "api/missing.http")
	require.ErrorContains(t, err, "api/missing.http:3: INCLUDE steps/missing.steps")
}

// Each step sees what those before it SET, a RUN_BACKGROUND_JOBS runs the jobs, and a request is sent as written:
// to the test server whatever its host, with its redirect not followed
func TestRun(t *testing.T) {
	fsys := fstest.MapFS{"api/echo.http": {Data: []byte(`NAME first
GET http://localhost/echo?say=hello
ASSERT_STATUS 200
SET said {{ .first.said }}

RUN_BACKGROUND_JOBS

NAME second
GET http://abcdefgh.ns.localhost/echo?say={{ .said }}-again
ASSERT_STATUS 302
ASSERT_EQUALS "/moved" "#Location.0"
ASSERT_EQUALS "hello-again" "@said"
ASSERT_EQUALS "abcdefgh.ns.localhost" "@host"
`)}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("Accept-Encoding"))
		if r.Host != "localhost" {
			w.Header().Set("Location", "/moved")
			w.WriteHeader(http.StatusFound)
		}
		json.NewEncoder(w).Encode(map[string]string{"said": r.URL.Query().Get("say"), "host": r.Host})
	}))
	t.Cleanup(ts.Close)

	echo, err := Parse(fsys, "api/echo.http")
	require.NoError(t, err)
	ran := 0
	ctx := map[string]any{}
	(&Runner{Server: ts, RunBackgroundJobs: func() { ran++ }}).Run(t, echo, ctx)
	require.Equal(t, 1, ran)
	require.Equal(t, "hello-again", ctx["second"].(map[string]any)["said"])
}

// A STREAM's events are queued by name, their data as JSON, for a WAIT, with comments, as its pings are, skipped
func TestRunStream(t *testing.T) {
	fsys := fstest.MapFS{"api/stream.flow": {Data: []byte(`NAME readings
STREAM http://abcdefgh.localhost/data/glucose/events
H Authorization: Bearer token

NAME created
WAIT created
ASSERT_EQUALS 120 "@sgv"

NAME second
WAIT created
ASSERT_EQUALS 130 "@sgv"
`)}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(": ping\n\nevent: created\ndata: {\"sgv\": 120}\n\nevent: created\ndata: {\"sgv\": 130}\n\n"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(ts.Close)
	stream, err := Parse(fsys, "api/stream.flow")
	require.NoError(t, err)
	(&Runner{Server: ts}).Run(t, stream, map[string]any{})
}
