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

//go:embed nightscout/site_v1_treatments.http
var siteV1TreatmentsHTTP string

func TestUserTreatments(t *testing.T) {
	requestContext := map[string]any{}
	requestFile, err := ParseRequestFile(siteV1TreatmentsHTTP, requestContext)
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
