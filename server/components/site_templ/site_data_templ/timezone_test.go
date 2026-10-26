package site_data_templ

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestFormatReadingDateInTheSettingsTimezone checks times are shown in the identity's timezone, then the site's
func TestFormatReadingDateInTheSettingsTimezone(t *testing.T) {
	auckland, kathmandu := "Pacific/Auckland", "Asia/Kathmandu"
	summer := time.Date(2026, 1, 15, 0, 30, 0, 0, time.UTC).UnixMilli()
	winter := time.Date(2026, 7, 15, 0, 30, 0, 0, time.UTC).UnixMilli()
	tests := []struct {
		name     string
		site     *string
		identity *string
		ms       int64
		want     string
	}{
		{name: "neither set is UTC", ms: summer, want: "2026-01-15 00:30:00"},
		{name: "the site's, in daylight saving", site: &auckland, ms: summer, want: "2026-01-15 13:30:00"},
		{name: "the site's, out of daylight saving", site: &auckland, ms: winter, want: "2026-07-15 12:30:00"},
		{name: "the identity's over the site's", site: &auckland, identity: &kathmandu, ms: summer, want: "2026-01-15 06:15:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{Timezone: tt.site}, "site", "", 0, 0, "")
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Timezone: tt.identity})
			require.Equal(t, tt.want, formatReadingDate(ctx, tt.ms))
		})
	}
	// Without a site, as when rendered alone, times are in UTC
	require.Equal(t, "2026-01-15 00:30:00", formatReadingDate(context.Background(), summer))
}
