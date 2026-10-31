package aslog

import (
	"bytes"
	"context"
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
