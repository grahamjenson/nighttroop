package control_rate_limits_templ

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// A rate limiter stands out amber once its most used has spent half its burst, and red while any would refuse, and so
// does its group's title, from the highest of the group's rate limiters
func TestRateLimitersStandOutNearlyFullOrRefusing(t *testing.T) {
	quiet := config.RateLimiterStats{Name: "identity_get", Per: "identity", Setting: "site_get", Scale: 1, FullestPercent: 10}
	tests := []struct {
		name     string
		stats    config.RateLimiterStats
		level    int
		row, bar string
		status   string
	}{
		{name: "in use", stats: config.RateLimiterStats{FullestPercent: 49}, bar: "bg-primary"},
		{name: "nearly full", stats: config.RateLimiterStats{FullestPercent: 50}, level: 1, row: "bg-amber-50", bar: "bg-amber-500", status: "Nearly full"},
		{name: "refusing", stats: config.RateLimiterStats{FullestPercent: 100, Limited: 1}, level: 2, row: "bg-red-50", bar: "bg-red-600", status: "Refusing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			busy := tt.stats
			busy.Name, busy.Per, busy.Setting, busy.Scale = "site_get", "site", "site_get", config.SiteIdentitiesShare
			stats := []config.RateLimiterStats{busy, quiet}
			require.Equal(t, tt.level, level(busy))
			require.Equal(t, tt.level, groupLevel(stats, []string{"site_get", "site_post"}))
			require.Zero(t, groupLevel(stats, []string{"site_post"}), "another group's rate limiters do not mark it")

			var page bytes.Buffer
			settings := []config.RateSetting{{Name: "site_get", Env: "SITE_GET_PER_MINUTE", Default: config.Rate{PerMinute: 60, Burst: 120}, Rate: config.Rate{PerMinute: 60, Burst: 120}}}
			require.NoError(t, RateSettings(settings, stats, true).Render(context.Background(), &page))
			html := page.String()
			require.Contains(t, html, `id="control--rate-limits--site_get--bar" class="h-full min-w-[2px] rounded-full `+tt.bar+`"`)
			status := regexp.MustCompile(`id="control--rate-limits--group--site_data--status" data-rate-limits-live>(<span[^>]*>([^<]*)</span>)?</span>`).FindStringSubmatch(html)
			require.NotNil(t, status)
			require.Equal(t, tt.status, status[2], "the group's title is marked")
			require.Contains(t, html, `id="control--rate-limits--site_get" data-rate-limits-live class="`+strings.TrimSpace("border-b text-gray-600 "+tt.row)+`"`)
		})
	}
}
