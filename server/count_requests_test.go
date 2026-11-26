package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nighttroop.t1d.nz/utils/metrics"
)

// Each request is counted in the metrics by the router its host is for, with its status, and never its site
func TestRequestsAreCountedByRouter(t *testing.T) {
	teapot := countRequests([]string{"nighttroop.test"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	for _, host := range []string{"control.nighttroop.test", "www.nighttroop.test", "yherqdif.nighttroop.test",
		"yherqdif.ns.nighttroop.test", "wwcteswd.ns.nighttroop.test", "nighttroop.example"} {
		teapot.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://"+host+"/api/v1/entries", nil))
	}

	recorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
	body, _ := io.ReadAll(recorder.Body)
	for _, want := range []string{
		`nighttroop_http_requests_total{code="418",router="control"} 1`,
		`nighttroop_http_requests_total{code="418",router="public"} 1`,
		`nighttroop_http_requests_total{code="418",router="site"} 1`,
		`nighttroop_http_requests_total{code="418",router="nightscout"} 2`,
		`nighttroop_http_requests_total{code="418",router="unknown"} 1`,
		`nighttroop_http_request_duration_seconds_count{router="nightscout"} `,
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("metrics have no %q", want)
		}
	}
	if strings.Contains(string(body), "yherqdif") {
		t.Error("a site's ID is in the metrics")
	}
}
