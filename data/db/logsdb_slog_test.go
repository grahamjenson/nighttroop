package db_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// slog logs are written to the logs table in a batch, with their site, request and job, and so are events and errors
// recorded in a request or job
func TestSlogLogsAreWrittenWithTheirRequestAndJob(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	logs, q := controlDB.Logs, controlDB.Logs.Queries
	buffer := &aslog.Buffer{Full: make(chan struct{}, 1)}
	logger := slog.New(aslog.ContextHandler{Handler: aslog.BufferHandler{Buffer: buffer}})
	site := "slogsite"
	request := aslog.NewContext(context.Background(), slog.String("request_id", "req-42"), slog.String("site_id", site))
	job := aslog.NewJob(context.Background(), "sync")
	_, jobID := aslog.IDs(job)
	require.NotEmpty(t, jobID)

	logger.ErrorContext(request, "cannot sync", "error", errors.New("refused"))
	logger.InfoContext(job, "synced", "readings", 12)
	logs.FlushFrom(buffer)

	found, err := q.SearchLogs(t.Context(), logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, found, 2)
	require.Equal(t, "synced", found[0].Message)
	require.Nil(t, found[0].RequestID)
	require.Equal(t, jobID, *found[0].JobID)
	require.Equal(t, "cannot sync", found[1].Message)
	require.Equal(t, "error", found[1].Level)
	require.Equal(t, site, *found[1].SiteID)
	require.Equal(t, "req-42", *found[1].RequestID)

	logs.SyncError(request, site, consts.SyncSourceTypeDexcom, errors.New("refused"))
	logs.SiteCreatedEvent(job, site)
	errs, err := q.SearchErrors(t.Context(), logsdb.SearchErrorsParams{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, "req-42", *errs[0].RequestID)
	events, err := q.SearchEvents(t.Context(), logsdb.SearchEventsParams{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, jobID, *events[0].JobID)
}
