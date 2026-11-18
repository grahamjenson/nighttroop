package replica

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// Only Litestream's warnings and errors are logged, with the attributes and groups they are given
func TestWarningsOnly(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(warnings{slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})}).With("system", "store").WithGroup("g")
	logger.Debug("synced")
	logger.Info("compacted")
	logger.Warn("slow", "n", 1)
	logger.Error("failed")
	require.NotContains(t, logs.String(), "synced")
	require.NotContains(t, logs.String(), "compacted")
	require.Contains(t, logs.String(), `level=WARN msg=slow system=store g.n=1`)
	require.Contains(t, logs.String(), `level=ERROR msg=failed system=store`)
}
