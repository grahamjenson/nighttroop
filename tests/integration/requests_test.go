package integration_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"nighttroop.t1d.nz/server"
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

	config, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()
	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

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
