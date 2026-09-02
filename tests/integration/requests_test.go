package integration_test

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/**
var embeddedFiles embed.FS

func TestRequests(t *testing.T) {

	config, controlDB, siteInformation, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	defaultJwt, _, err := config.Security.GenerateUserJWTToken(siteInformation.SiteDB.Subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

	testutils.DumpOnFailure(t, controlDB, "user_treatments_test_failure_dump.sqlite")

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

		runTheRequestFileAsATest(t, path, router, siteInformation, defaultJwt)

		return nil
	})

	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}

}

func runTheRequestFileAsATest(t *testing.T, path string, router http.Handler, siteInformation *db.SiteCreateResult, defaultJwt string) {
	removeNightscout := strings.TrimPrefix(path, "nightscout/")
	t.Run(removeNightscout, func(t *testing.T) {
		requestFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read request file: %v", err)
		}
		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running request file: %s\n", path)
		fmt.Println("===============================================")

		requestContext := map[string]any{
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		}
		requestFile, err := ParseRequestFile(string(requestFileBytes), requestContext, path)
		if err != nil {
			t.Fatalf("failed to parse request file: %v", err)
		}
		for _, request := range requestFile.Requests {
			realRequest, err := request.BuildRealRequest(requestContext)
			if err != nil {
				t.Fatalf("failed to build real request: %v", err)
			}

			if request.IsFunction {
				requestContext, err = realRequest.ExecuteFunction(requestContext)
				if err != nil {
					t.Fatalf("failed to execute function: %v", err)
				}

			} else {
				fmt.Println()
				fmt.Println("=============== " + request.Name + "  ================")
				fmt.Println()
				realRequest.PrintRequest()
				fmt.Println()
				fmt.Println("- - - - - - - - - - - - - - - - ")
				fmt.Println()
				// default information

				statusCode, output, headers, newContext, err := realRequest.ExecuteRequest(router, requestContext)
				if err != nil {
					t.Fatalf("failed to execute request: %v, method: %s, url: %s", err, realRequest.Method, realRequest.URL)
				}
				requestContext = newContext
				fmt.Println()
				fmt.Println("- - - - - - - - - - - - - - - - ")
				fmt.Println()
				fmt.Printf("STATUS CODE: %d\n", statusCode)
				var buf bytes.Buffer
				encoder := json.NewEncoder(&buf)
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					fmt.Printf("RESPONSE: %s and failed to format response: %v\n", output, err)
				} else {
					fmt.Printf("RESPONSE:\n%s", buf.String())
				}
				fmt.Println("===============================")

				realRequest.ExecuteAsserts(t, statusCode, output, headers, requestContext)

			}
		}
	})
}
