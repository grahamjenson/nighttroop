package integration_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		t.Run(request.Name, func(t *testing.T) {
			fmt.Println("THISSS", requestContext)
			realRequest, err := request.BuildRealRequest(requestContext)
			request.PrintRequest()
			if request.IsFunction {
				requestContext, err = realRequest.ExecuteFunction(requestContext)
				if err != nil {
					t.Fatalf("failed to execute function: %v", err)
				}

			} else {

				if err != nil {
					t.Fatalf("failed to build real request: %v", err)
				}
				statusCode, output, headers := JSONRequest(t, router, realRequest)

				fmt.Printf("Request: %s %s\n", realRequest.Method, realRequest.URL)
				fmt.Printf("Status Code: %d\n", statusCode)
				prettyOutput, _ := json.MarshalIndent(output, "", "  ")
				fmt.Printf("Output: %+v\n", string(prettyOutput))
				fmt.Printf("Headers: %+v\n", headers)

				requestContext[request.Name] = output
			}
		})
	}

}

func JSONRequest(t *testing.T, handler http.Handler, request Request) (int, map[string]any, http.Header) {

	var buf bytes.Buffer
	if request.Body != "" {
		if err := json.NewEncoder(&buf).Encode(request.Body); err != nil {
			t.Fatalf("failed to encode json: %v", err)
		}
	}

	req := httptest.NewRequest(request.Method, request.URL, &buf)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var out map[string]any

	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		if rr.Code == 200 || rr.Code == 201 {
			// It must pass if the status code is 200 or 201, otherwise it is an errors
			t.Fatalf("failed to decode json: %v\nbody: %s\nstatus: %d", err, rr.Body.String(), rr.Code)
		} else {
			fmt.Println("failed to decode json for", request.ToFlatString())
		}

	}

	return rr.Code, out, rr.Header()
}
