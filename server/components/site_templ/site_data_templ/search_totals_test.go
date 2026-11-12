package site_data_templ

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/i18n"
)

// A day in the totals box is Today, Yesterday, its weekday and date this year, or its date and year before
func TestTotalsDate(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 15, 0, 0, time.UTC)
	tests := []struct {
		at   time.Time
		want string
	}{
		{at: now, want: "Today"},
		{at: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC), want: "Today"},
		{at: time.Date(2026, 11, 8, 23, 59, 0, 0, time.UTC), want: "Yesterday"},
		{at: time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC), want: "Sat 7 Nov"},
		{at: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), want: "Thu 1 Jan"},
		{at: time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC), want: "31 Dec 2025"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			require.Equal(t, test.want, totalsDate(context.Background(), test.at.UnixMilli(), now))
		})
	}
}

// The totals' noun is one of them for 1, and many for any other count
func TestTotalsNoun(t *testing.T) {
	ctx := context.Background()
	require.Equal(t, "temp targets", i18n.N(ctx, "site.data.temp_targets.noun", 0))
	require.Equal(t, "temp target", i18n.N(ctx, "site.data.temp_targets.noun", 1))
	require.Equal(t, "temp targets", i18n.N(ctx, "site.data.temp_targets.noun", 2))
}
