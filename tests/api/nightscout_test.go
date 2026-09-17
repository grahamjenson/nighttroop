package api_test

import (
	"embed"
	_ "embed"
	"io/fs"
	"path/filepath"
	"testing"

	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/**
var embeddedFiles embed.FS

func TestNS(t *testing.T) {

	config, controlDB, siteInformation, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	defaultJwt, _, err := config.Security.GenerateUserJWTToken(siteInformation.SiteDB.Subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

	testutils.DumpOnFailure(t, controlDB, "test_ns")

	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".http" {
			return nil
		}

		RunTheRequestFileAsATest(t, "nightscout", path, router, siteInformation, defaultJwt)

		return nil
	})

	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}

}
