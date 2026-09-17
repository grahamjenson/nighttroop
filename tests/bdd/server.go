package bdd

import (
	"net/http"
	"testing"

	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

func NewRouter(t *testing.T) http.Handler {
	config, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)
	return router
}
