package site_settings_templ

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// A backfill's toast says what it fetches, and that it fetches the recent days again when it does
func TestBackfillQueuedSaysWhatItFetches(t *testing.T) {
	from := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC).UnixMilli()
	to, recent := from+24*time.Hour.Milliseconds(), from+5*24*time.Hour.Milliseconds()
	for _, test := range []struct {
		name       string
		recentFrom int64
		want       string
	}{
		{"one range", 0, "Fetching records from 1 Nov 10:00 to 2 Nov 10:00 UTC<"},
		{"and the recent days", recent, "Fetching records from 1 Nov 10:00 to 2 Nov 10:00 UTC, and again from 6 Nov 10:00 UTC to fill any gaps"},
	} {
		var buf bytes.Buffer
		require.NoError(t, BackfillQueued(consts.ConnectorTypeNightscout, from, to, test.recentFrom).Render(t.Context(), &buf))
		require.Contains(t, buf.String(), test.want, test.name)
	}
}
