package main

import (
	"io/fs"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/flow"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/testutils"
)

// new_site.flow makes a site with its readings, treatments and phone status, as it would on staging
func TestNewSiteScript(t *testing.T) {
	config, controlDB, _ := testutils.NewControlDBWithSite(t)
	ts := httptest.NewServer(server.GetMainHandler(controlDB, config))
	t.Cleanup(ts.Close)
	run(t, os.DirFS("../../scripts/staging"), "new_site.flow", jsonapi.Client(ts), settings{scheme: "http", host: "localhost"})
}

// Every staging script parses, and none needs a test server's background jobs
func TestStagingScriptsParse(t *testing.T) {
	scripts := os.DirFS("../../scripts/staging")
	paths, err := fs.Glob(scripts, "*.flow")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		script, err := flow.Parse(scripts, path)
		require.NoError(t, err, path)
		require.NoError(t, refuseBackgroundJobs(script), path)
	}
}

// A script with RUN_BACKGROUND_JOBS, as an HTTP or a browser step, is refused before it runs
func TestRefusesBackgroundJobs(t *testing.T) {
	scripts := fstest.MapFS{
		"http.flow":   {Data: []byte("NAME health\nGET http://localhost/health\nRUN_BACKGROUND_JOBS\n")},
		"browse.flow": {Data: []byte("BROWSE\nGOTO {{ server_url }}/\nRUN_BACKGROUND_JOBS\n")},
	}
	for _, path := range []string{"http.flow", "browse.flow"} {
		script, err := flow.Parse(scripts, path)
		require.NoError(t, err)
		require.ErrorContains(t, refuseBackgroundJobs(script), path+": RUN_BACKGROUND_JOBS")
	}
}
