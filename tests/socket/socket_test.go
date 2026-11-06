package socket_test

import (
	"context"
	"embed"
	"io/fs"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/flow"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/**
var embeddedFiles embed.FS

// TestSocketIO runs each .sio file against the whole server, with its own connection
func TestSocketIO(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	router := server.GetMainHandler(controlDB, config)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, config, s3utils.NewS3Client(config))
	testutils.DumpOnFailure(t, controlDB, "test_socket")

	err := fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".sio" {
			return err
		}
		t.Run(strings.TrimPrefix(path, "nightscout/"), func(t *testing.T) {
			socketFlow, err := flow.Parse(embeddedFiles, path)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			runner := &flow.Runner{
				Client:            flow.Client(ts),
				RunBackgroundJobs: func() { backgroundJobs.RunAllNow(context.Background()) },
			}
			runner.Run(t, socketFlow, map[string]any{
				"default_site":  site.SiteDB.Subdomain + ".ns.localhost",
				"default_token": site.PwdNSToken,
			})
		})
		return nil
	})
	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}
}
