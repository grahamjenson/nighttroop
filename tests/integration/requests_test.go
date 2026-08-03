package integration_test

import (
	_ "embed"
	"fmt"
	"testing"

	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed nightscout/user_treatments.http
var userTreatmentsHTTP string

func TestUserTreatments(t *testing.T) {
	requestFile, err := ParseRequestFile(userTreatmentsHTTP, nil)
	if err != nil {
		t.Fatalf("failed to parse request file: %v", err)
	}

	config, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()
	if err != nil {
		fmt.Println("djsoigjdsoigjoidsgjdsoig")
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)

	for _, requestFile := range requestFile.Requests {
		t.Run(requestFile.Name, func(t *testing.T) {
			statusCode, output, headers := testutils.JSONRequest[map[string]any](t, router, requestFile.Method, requestFile.Path, requestFile.Body)

			fmt.Printf("Request: %s %s\n", requestFile.Method, requestFile.Path)
			fmt.Printf("Status Code: %d\n", statusCode)
			fmt.Printf("Output: %+v\n", output)
			fmt.Printf("Headers: %+v\n", headers)
		})
	}

	fmt.Println("asd ", requestFile)
}
