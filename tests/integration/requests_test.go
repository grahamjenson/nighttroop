package integration_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/user_treatments.http
var userTreatmentsHTTP string

func TestUserTreatments(t *testing.T) {
	requestContext := map[string]any{}
	requestFile, err := ParseRequestFile(userTreatmentsHTTP, requestContext)
	if err != nil {
		t.Fatalf("failed to parse request file: %v", err)
	}

	config, controlDB, err := testutils.GetConfigControlDBForTest()
	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

	DumpOnFailure(t, controlDB, "user_treatments_test_failure_dump.sqlite")

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
			realRequest.PrintRequest()

			statusCode, output, headers, newContext := realRequest.ExecuteRequest(t, router, requestContext)
			requestContext = newContext

			fmt.Printf("Request: %s %s\n", realRequest.Method, realRequest.URL)
			fmt.Printf("Status Code: %d\n", statusCode)
			prettyOutput, _ := json.MarshalIndent(output, "", "  ")
			fmt.Printf("Output: %+v\n", string(prettyOutput))
			fmt.Printf("Headers: %+v\n", headers)

		}

	}

}

// DumpOnFailure registers a cleanup hook that saves the in-memory DB
// to disk ONLY if the test fails.
func DumpOnFailure(t *testing.T, controlDB *db.ControlDB, filePrefix string) {
	t.Helper()

	dumpFolder := filepath.Join(utils.FindProjectRoot(), "tmp/test_db_dumps/", filePrefix+"_"+t.Name())
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
