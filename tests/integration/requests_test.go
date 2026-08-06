package integration_test

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/**
var embeddedFiles embed.FS

func TestRequests(t *testing.T) {

	config, controlDB, err := testutils.GetConfigControlDBForTest()
	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

	DumpOnFailure(t, controlDB, "user_treatments_test_failure_dump.sqlite")

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

		runTheRequestFileAsATest(t, path, router)

		return nil
	})

	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}

}

func runTheRequestFileAsATest(t *testing.T, requestFilePath string, router http.Handler) {
	removeNightscout := strings.TrimPrefix(requestFilePath, "nightscout/")
	t.Run(removeNightscout, func(t *testing.T) {
		requestFileBytes, err := embeddedFiles.ReadFile(requestFilePath)
		if err != nil {
			t.Fatalf("failed to read request file: %v", err)
		}
		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running request file: %s\n", requestFilePath)
		fmt.Println("===============================================")

		requestContext := map[string]any{}
		requestFile, err := ParseRequestFile(string(requestFileBytes), requestContext, requestFilePath)
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
				fmt.Println("===============================")
				realRequest.PrintRequest()
				fmt.Println()
				fmt.Println("- - - - - - - - - - - - - - - - ")
				fmt.Println()
				statusCode, output, _, newContext := realRequest.ExecuteRequest(router, requestContext)
				requestContext = newContext
				fmt.Println()
				fmt.Println("- - - - - - - - - - - - - - - - ")
				fmt.Println()
				fmt.Printf("STATUS CODE: %d\n", statusCode)
				prettyOutput, _ := json.MarshalIndent(output, "", "  ")
				fmt.Printf("RESPONSE: %+v\n", string(prettyOutput))
				fmt.Println("===============================")
				realRequest.ExecuteAsserts(t, statusCode, output, requestContext)

			}

		}
	})
}

// DumpOnFailure registers a cleanup hook that saves the in-memory DB
// to disk ONLY if the test fails.
func DumpOnFailure(t *testing.T, controlDB *db.ControlDB, filePrefix string) {
	t.Helper()

	dumpFolder := filepath.Join(utils.FindProjectRoot(), ".test/dumps/", filePrefix+"_"+t.Name())
	// delete folder if it exists
	_ = os.RemoveAll(dumpFolder)

	t.Cleanup(func() {

		// Only run if the test actually failed
		if !t.Failed() {
			return
		}

		// Ensure output directory exists (e.g., ./debug_dumps)

		utils.MkDirs(dumpFolder)
		backupPath := filepath.Join(dumpFolder, "control.sqlite")
		err := utils.SqliteBackup(context.TODO(), controlDB.SQLiteDB, backupPath)
		if err != nil {
			t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved: %v", err)
			return
		}

		for _, site := range controlDB.AllSites() {
			if site.Subdomain == "" {
				continue
			}
			_, backupPath := utils.SiteDBPaths(site.Subdomain)
			backupPath = filepath.Join(dumpFolder, site.Subdomain+".sqlite")
			err := utils.SqliteBackup(context.TODO(), site.SQLiteDB, backupPath)
			if err != nil {
				t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved for site %s: %v", site.Subdomain, err)
				return
			}
		}
		t.Logf("🛑 Test failed! In-memory DB snapshot saved to: %s", backupPath)
	})
}
