package api_test

import (
	"context"
	"embed"
	_ "embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/** nighttroop/** steps/**
var embeddedFiles embed.FS

func TestNS(t *testing.T) {
	// A new site can create 5 identities at once, so identities_max.http reaches the identity limit within that burst
	t.Setenv("MAX_IDENTITIES_PER_USER", "8")
	// Every file signs in as its site's admin from the same IP
	t.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "100")

	config, controlDB, siteInformation := testutils.NewControlDBWithSite(t)

	defaultJwt, _, err := config.Security.GenerateUserJWTToken(siteInformation.SiteDB.Subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)
	// Background jobs are run after every request instead of on their timers
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, config, s3utils.NewS3Client(config))

	testutils.DumpOnFailure(t, controlDB, "test_ns")

	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".http" {
			return err
		}
		t.Run(strings.TrimPrefix(path, "nightscout/"), func(t *testing.T) {
			requestFile, err := jsonapi.ParseRequestFile(embeddedFiles, path)
			if err != nil {
				t.Fatalf("failed to parse request file: %v", err)
			}
			requestFile.Run(t, router, map[string]any{
				"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
				"default_token": siteInformation.PwdNSToken,
				"default_jwt":   defaultJwt,
			}, func() { backgroundJobs.RunAllNow(context.Background()) })
		})
		return nil
	})
	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}
}
