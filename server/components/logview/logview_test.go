package logview

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// col is a table column's own element, not its colgroup
var col = regexp.MustCompile(`<col[ >/]`)

// The logs table keeps its columns' widths, shows each level as a badge, and, on a View that Links, gives each log a
// button at its right narrowing the logs to its request, or else its job, which does not open the log's details. A
// log with neither has no button, and a View without links, as a site's own, has no buttons or their column.
func TestLogsTable(t *testing.T) {
	request, job := "host/abc-000001", "3587779f1fb8d8dd"
	rows := []Row{
		{ID: 1, Name: "info", Message: "200: OK", RequestID: &request, JobID: &job},
		{ID: 2, Name: "error", Message: "Cannot sync", JobID: &job},
		{ID: 3, Name: "warn", Message: "Slow query"},
	}
	render := func(v View) string {
		var html bytes.Buffer
		require.NoError(t, LogsTable(v, "logs", rows).Render(context.Background(), &html))
		return html.String()
	}

	html := render(View{Prefix: "control", Links: true, About: "Site", Date: func(context.Context, int64) string { return "" }})
	require.Contains(t, html, "table-fixed")
	require.Len(t, col.FindAllString(html, -1), 5, "level, message, site, when and the button")
	require.Contains(t, html, `id="control--log--1--request-id"`)
	require.Contains(t, html, `href="/logs?request_id=host%2Fabc-000001"`)
	require.NotContains(t, html, `id="control--log--1--job-id"`, "a log's request is preferred to its job")
	require.Contains(t, html, `id="control--log--2--job-id"`)
	require.Contains(t, html, `href="/logs?job_id=3587779f1fb8d8dd"`)
	require.NotContains(t, html, `id="control--log--3--`)
	require.Equal(t, 2, bytes.Count([]byte(html), []byte("event.stopPropagation()")), "the buttons do not open the log's details")
	require.Contains(t, html, "bg-red-")

	site := render(View{Prefix: "site", Path: "/logs", Date: func(context.Context, int64) string { return "" }})
	require.Contains(t, site, "table-fixed")
	require.NotContains(t, site, "--request-id")
	require.NotContains(t, site, "--job-id")
	require.Len(t, col.FindAllString(site, -1), 3, "level, message and when")
}
