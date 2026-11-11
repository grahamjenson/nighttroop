package aslog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextHandler(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(ContextHandler{slog.NewTextHandler(&logs, nil)})
	ctx := NewContext(context.Background(), slog.String("request_id", "r1"))

	// What is added to a context made from it, as a request's is after sign in, is logged with the first too
	signedIn := context.WithValue(ctx, struct{}{}, "signed in")
	AddAttrs(signedIn, slog.Int64("identity_id", 7))
	AddAttrs(signedIn, slog.Int64("identity_id", 8))
	logger.InfoContext(ctx, "outer")
	require.Contains(t, logs.String(), "msg=outer request_id=r1 identity_id=8")
	require.Equal(t, 1, strings.Count(logs.String(), "identity_id"), "an attribute added again replaces the first")

	// A logger's own attributes are kept beside them
	logs.Reset()
	logger.With("job", "sync").ErrorContext(ctx, "failed")
	require.Contains(t, logs.String(), "job=sync request_id=r1 identity_id=8")

	// A context not made with NewContext logs nothing more, and cannot be added to
	logs.Reset()
	AddAttrs(context.Background(), slog.String("site_id", "ignored"))
	logger.InfoContext(context.Background(), "plain")
	logger.Info("no context")
	require.NotContains(t, logs.String(), "site_id")
	require.NotContains(t, logs.String(), "request_id")
}

// A job's context logs the job and its own ID, and work for a site within it the site too, without changing what the
// job's own context logs
func TestJobAndSiteContexts(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(ContextHandler{slog.NewTextHandler(&logs, nil)})

	job := NewJob(context.Background(), "BackupSiteJob", slog.String("batch", "ab"))
	alice, bob := WithSite(job, "alice"), WithSite(job, "bob")
	logger.InfoContext(alice, "for alice")
	logger.InfoContext(bob, "for bob")
	logger.InfoContext(job, "for the job")
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Len(t, lines, 3)
	require.Regexp(t, `job_id=[0-9a-f]{16} job=BackupSiteJob batch=ab site_id=alice$`, lines[0])
	require.Regexp(t, `job=BackupSiteJob batch=ab site_id=bob$`, lines[1])
	require.Regexp(t, `job=BackupSiteJob batch=ab$`, lines[2])
	jobID := func(line string) string { return strings.Fields(line[strings.Index(line, "job_id="):])[0] }
	require.Equal(t, jobID(lines[0]), jobID(lines[2]), "one run's logs share its ID")

	// Each run has its own ID, and a site given again replaces the first
	logs.Reset()
	logger.InfoContext(WithSite(alice, "carol"), "again")
	logger.InfoContext(NewJob(context.Background(), "BackupSiteJob"), "next run")
	lines = strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Equal(t, 1, strings.Count(lines[0], "site_id="))
	require.Contains(t, lines[0], "site_id=carol")
	require.NotEqual(t, jobID(lines[0]), jobID(lines[1]))
}

// An error that is its context being cancelled or timing out is logged as a warning, so Sentry is not sent one for
// each layer of a request a client left; any other error, and a cancellation at a lower level, are left as they are
func TestCancellationIsAWarning(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(ContextHandler{slog.NewTextHandler(&logs, nil)})
	tests := []struct {
		name  string
		log   func()
		level string
	}{
		{"cancelled", func() { logger.Error("Error writing glucose", "error", context.Canceled) }, "WARN"},
		{"wrapped timeout", func() {
			logger.Error("failed to search the database", "error", fmt.Errorf("searching: %w", context.DeadlineExceeded))
		}, "WARN"},
		{"other error", func() { logger.Error("failed to map data", "error", errors.New("unknown event type")) }, "ERROR"},
		{"no error", func() { logger.Error("something failed") }, "ERROR"},
		{"cancelled info", func() { logger.Info("stopped", "error", context.Canceled) }, "INFO"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logs.Reset()
			test.log()
			require.Contains(t, logs.String(), "level="+test.level+" ")
		})
	}
}
