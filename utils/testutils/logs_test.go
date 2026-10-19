package testutils_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestUseLogHandlerRestoresDefault(t *testing.T) {
	before := slog.Default()
	var logs bytes.Buffer

	t.Run("swapped", func(t *testing.T) {
		testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))
		slog.Info("inside")
		assert.NotSame(t, before, slog.Default())
	})

	assert.Contains(t, logs.String(), "inside")
	assert.Same(t, before, slog.Default())
}
