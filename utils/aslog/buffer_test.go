package aslog

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// A log is held with its site, request and job taken from its context, and the rest of its attributes as JSON, its
// groups as objects, less any IP address
func TestBufferHandlerHoldsLogsWithTheirIDs(t *testing.T) {
	buffer := &Buffer{Full: make(chan struct{}, 1)}
	logger := slog.New(ContextHandler{BufferHandler{Buffer: buffer}})
	ctx := NewJob(NewContext(context.Background(), slog.String("request_id", "req-1"), slog.String("site_id", "abcdefgh")), "sync")

	logger.With("component", "sync").WarnContext(ctx, "skipped", "count", 3, "error", errors.New("refused"),
		slog.Group("request", slog.String("path", "/api"), slog.String("ip", "10.0.0.1")), slog.String("ip", "10.0.0.2"))
	records, dropped := buffer.Take()
	require.Zero(t, dropped)
	require.Len(t, records, 1)
	record := records[0]
	require.Equal(t, "warn", record.Level)
	require.Equal(t, "skipped", record.Message)
	require.Equal(t, "abcdefgh", record.SiteID)
	require.Equal(t, "req-1", record.RequestID)
	require.Len(t, record.JobID, 16)
	require.JSONEq(t, `{"component":"sync","count":3,"error":"refused","job":"sync","request":{"path":"/api"}}`, record.Attributes)
}

// A record is cut short, and past the buffer's limit logs are dropped and counted rather than held, so a slow or
// failing write never grows memory
func TestBufferKeepsWithinItsLimits(t *testing.T) {
	buffer := &Buffer{Full: make(chan struct{}, 1)}
	logger := slog.New(BufferHandler{Buffer: buffer})

	logger.Info(strings.Repeat("m", 5000), "big", strings.Repeat("a", 5000))
	records, _ := buffer.Take()
	require.LessOrEqual(t, len(records[0].Message), maxMessageBytes)
	require.Equal(t, `{"truncated":true}`, records[0].Attributes)

	for range 3000 {
		logger.Info(strings.Repeat("m", 500))
	}
	require.Len(t, buffer.Full, 1, "a full batch is signalled to be written")
	taken, dropped := buffer.Take()
	require.Len(t, taken, FlushRecords)
	require.Positive(t, dropped)
	require.LessOrEqual(t, 500*(len(taken)+len(buffer.records)), maxBufferedBytes)
}
