package api_test

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
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/** nighttroop/**
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

func RunTheRequestFileAsATest(t *testing.T, root string, path string, router http.Handler, siteInformation *db.SiteCreateResult, defaultJwt string) {
	removeNightscout := strings.TrimPrefix(path, root+"/")
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
		requestFile, err := jsonapi.ParseRequestFile(string(requestFileBytes), requestContext, path)
		if err != nil {
			t.Fatalf("failed to parse request file: %v", err)
		}
		for _, request := range requestFile.Requests {

			if err != nil {
				t.Fatalf("failed to build real request: %v", err)
			}

			fmt.Println()
			fmt.Println("=============== " + request.Name + "  ================")
			fmt.Println()
			request.PrintRequest()
			fmt.Println()
			fmt.Println("- - - - - - - - - - - - - - - - ")
			fmt.Println()
			// default information

			newContext, err := request.ExecuteRequest(t, router, requestContext)
			if err != nil {
				t.Fatalf("failed to execute request: %v, method: %s, url: %s", err, request.Method, request.URL)
			}

			output := newContext[request.Name]
			statusCode := newContext[request.Name+"_status"]

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

		}
	})
}
