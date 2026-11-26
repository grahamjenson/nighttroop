package metrics

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// scrape is the metrics as Prometheus reads them
func scrape(t *testing.T) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
	body, _ := io.ReadAll(recorder.Body)
	return string(body)
}

func TestMetricsAreServed(t *testing.T) {
	Request("site", 201, 30*time.Millisecond, false)
	// A Socket.IO connection is counted, but its minutes open are not a request's time
	Request("nightscout", 101, time.Minute, true)
	Refused("Global rate limit exceeded", 429)
	RateLimited()
	WriteRefused("disk_full")
	JobRun("CheckResources", time.Now(), nil)
	JobRun("connector_sync", time.Now(), errors.New("down"))
	Sites(3, map[string]string{"aaaaaaaa": "missing", "bbbbbbbb": "missing", "cccccccc": "cannot_open"})
	DiskFree(1 << 30)
	SiteDBs(2048)

	body := scrape(t)
	for _, want := range []string{
		`nighttroop_http_requests_total{code="201",router="site"} 1`,
		`nighttroop_http_requests_total{code="101",router="nightscout"} 1`,
		`nighttroop_http_request_duration_seconds_count{router="site"} 1`,
		`nighttroop_requests_refused_total{code="429",reason="Global rate limit exceeded"} 1`,
		`nighttroop_rate_limited_total 1`,
		`nighttroop_site_writes_refused_total{why="disk_full"} 1`,
		`nighttroop_job_runs_total{job="CheckResources",result="ok"} 1`,
		`nighttroop_job_runs_total{job="connector_sync",result="failed"} 1`,
		`nighttroop_job_duration_seconds_count{job="connector_sync"} 1`,
		`nighttroop_sites_open 3`,
		`nighttroop_sites_unavailable{reason="missing"} 2`,
		`nighttroop_sites_unavailable{reason="cannot_open"} 1`,
		`nighttroop_disk_free_bytes 1.073741824e+09`,
		`nighttroop_site_dbs_bytes 2048`,
		`go_goroutines `,
		`go_memstats_heap_alloc_bytes `,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics have no %q", want)
		}
	}
	if strings.Contains(body, `nighttroop_http_request_duration_seconds_count{router="nightscout"}`) {
		t.Error("a long lived request is timed")
	}

	// A site restored is no longer unavailable
	Sites(4, map[string]string{"cccccccc": "cannot_open"})
	if body := scrape(t); strings.Contains(body, `reason="missing"`) || !strings.Contains(body, `nighttroop_sites_open 4`) {
		t.Error("the sites unavailable are not the last check's")
	}
}
