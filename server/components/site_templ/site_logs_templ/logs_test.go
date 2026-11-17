package site_logs_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
)

// TestSiteLogsShowNoSiteOrControlLinks checks a site's logs pages never show a Site column or detail, and show request
// and job IDs as plain text, never linking to control's pages
func TestSiteLogsShowNoSiteOrControlLinks(t *testing.T) {
	request, job, batch, finished := "request-1", "job-1", "batch-1", int64(2000)
	event := &sitemodels.SiteEvent{ID: 1, Name: "site_backup", Batch: &batch, StartedAt: 1000, FinishedAt: &finished, RequestID: &request, JobID: &job}
	siteError := &sitemodels.SiteError{ID: 2, Message: "failed", CreatedAt: 1000, RequestID: &request, JobID: &job}
	log := &sitemodels.SiteLog{ID: 3, Level: "INFO", Message: "logged", CreatedAt: 1000, RequestID: &request, JobID: &job}
	tests := []struct {
		name      string
		component templ.Component
		ids       bool
	}{
		{name: "events", component: EventsTable([]*sitemodels.SiteEvent{event}, 0, 100, nil, nil)},
		{name: "event", component: EventModal(event), ids: true},
		{name: "errors", component: ErrorsTable([]*sitemodels.SiteError{siteError}, 0, 100, nil, nil)},
		{name: "error", component: ErrorModal(siteError), ids: true},
		{name: "logs", component: LogsTable([]*sitemodels.SiteLog{log}, 0, 100, nil)},
		{name: "log", component: LogModal(log), ids: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tt.component.Render(context.Background(), &buf))
			html := buf.String()
			require.NotContains(t, html, ">Site<", html)
			require.NotContains(t, html, "control--", html)
			require.NotContains(t, html, "request_id=", html)
			require.NotContains(t, html, "job_id=", html)
			if tt.ids {
				require.Contains(t, html, `<dd id="site--`+tt.name+`--request-id" class="break-all">request-1</dd>`)
				require.Contains(t, html, `<dd id="site--`+tt.name+`--job-id" class="break-all">job-1</dd>`)
			}
		})
	}
}
