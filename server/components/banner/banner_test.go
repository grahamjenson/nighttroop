package banner

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
)

func TestStatusBanner(t *testing.T) {
	tests := []struct {
		name   string
		status config.ServerStatus
		banner bool
		s3     bool
		sentry bool
		pager  bool
		sites  string
	}{
		{name: "all working", status: config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true}},
		{name: "s3 missing", status: config.ServerStatus{SentryWorking: true, PagerDutyWorking: true}, banner: true, s3: true},
		{name: "sentry missing", status: config.ServerStatus{S3Working: true, PagerDutyWorking: true}, banner: true, sentry: true},
		{name: "pagerduty missing", status: config.ServerStatus{S3Working: true, SentryWorking: true}, banner: true, pager: true},
		{name: "all missing", status: config.ServerStatus{}, banner: true, s3: true, sentry: true, pager: true},
		{name: "sites under 90% of the limit", status: config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true, Sites: 89, MaxSites: 100}},
		{name: "sites at 90% of the limit", status: config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true, Sites: 90, MaxSites: 100}, banner: true, sites: "Sites near their limit"},
		{name: "sites at the limit", status: config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true, Sites: 100, MaxSites: 100}, banner: true, sites: "Sites at their limit"},
		{name: "sites over the limit", status: config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true, Sites: 101, MaxSites: 100}, banner: true, sites: "Sites at their limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, Strip("control--status-banner", ForControl(context.Background(), tt.status), false).Render(context.Background(), &buf))
			html := buf.String()
			require.Equal(t, tt.banner, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner"`)), html)
			require.Equal(t, tt.s3, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--s3"`)), html)
			require.Equal(t, tt.sentry, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--sentry"`)), html)
			require.Equal(t, tt.pager, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--pagerduty"`)), html)
			require.Equal(t, tt.sites != "", bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--sites"`)), html)
			if tt.sites != "" {
				require.Contains(t, html, tt.sites)
			}
		})
	}
}

// The banner says how much memory is used and disk is free in bytes, where their totals are known, and the CPU's seconds
func TestResourceWarning(t *testing.T) {
	total := float64(954 << 30)
	tests := []struct {
		level config.ResourceLevel
		want  string
	}{
		{config.ResourceLevel{Resource: "disk", Value: 87, Limit: 60, Total: &total}, "124.0 GB free of 954.0 GB."},
		{config.ResourceLevel{Resource: "memory", Value: 75, Limit: 60, Total: new(float64(2 << 30))}, "1.5 GB used of 2.0 GB."},
		{config.ResourceLevel{Resource: "memory", Value: 75, Limit: 60}, "75% used; it warns from 60%."},
		{config.ResourceLevel{Resource: "cpu", Value: 25, Limit: 20}, "at 100% for 25 seconds; it warns from 20 seconds."},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, resourceWarning(context.Background(), tt.level))
	}
}

// The unavailable sites are named in order with why, at most ten, then how many more
func TestForControlUnavailable(t *testing.T) {
	sites := map[string]string{}
	for _, letter := range "abcdefghijkl" {
		sites["site"+string(letter)+"zzz"] = db.UnavailableMissing
	}
	banners := ForControl(context.Background(), config.ServerStatus{S3Working: true, SentryWorking: true, PagerDutyWorking: true, UnavailableSites: sites})
	require.Len(t, banners, 1)
	require.Equal(t, Error, banners[0].Level)
	require.Len(t, banners[0].Links, 10)
	require.Equal(t, Link{ID: "control--status-banner--unavailable--siteazzz", Href: "/sites/siteazzz", Text: "siteazzz", After: " (DB missing),"}, banners[0].Links[0])
	require.NotContains(t, banners[0].Links[9].After, ",")
	require.Contains(t, banners[0].After, "2")
}

// A site's banners are its storage first, then what control admins wrote for it or for every site, never another site's
func TestOfSite(t *testing.T) {
	kb := func(n int64) *int64 { return &n }
	written := []*controldb.SiteBanner{
		{ID: 1, Message: "For every site", Level: "info", SiteIdsJson: "[]"},
		{ID: 2, Message: "For alice <b>", Level: "warning", SiteIdsJson: `["alicezzz"]`},
		{ID: 3, Message: "For bob", Level: "warning", SiteIdsJson: `["bobzzzzz"]`},
		{ID: 4, Message: "Not read", Level: "info", SiteIdsJson: "{"},
	}
	tests := []struct {
		name          string
		site          controldb.Site
		writesRefused string
		want          []Banner
	}{
		{name: "nothing to say", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(10), MaxDbSizeKbLimit: kb(20)}},
		{name: "size not known", site: controldb.Site{ID: "davezzzz"}},
		{name: "under 90% of its storage", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(921), MaxDbSizeKbLimit: kb(1024)}},
		{name: "at 90% of its storage", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(922), MaxDbSizeKbLimit: kb(1024)}, want: []Banner{
			{ID: "site--banner--near-storage", Level: Info, Title: "Near the storage limit:", Text: "This site's data is 922.0 KB of its limit of 1.0 MB. Delete old data to make space."},
		}},
		{name: "at its storage limit, not over it", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(1024), MaxDbSizeKbLimit: kb(1024)}, want: []Banner{
			{ID: "site--banner--near-storage", Level: Info, Title: "Near the storage limit:", Text: "This site's data is 1.0 MB of its limit of 1.0 MB. Delete old data to make space."},
		}},
		{name: "over its storage", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(2048), MaxDbSizeKbLimit: kb(1024)}, want: []Banner{
			{ID: "site--banner--over-storage", Level: Warning, Title: "Over the storage limit:", Text: "This site's data is 2.0 MB, over its limit of 1.0 MB. Delete old data, or the site stops storing more."},
		}},
		{name: "writes refused, said once", site: controldb.Site{ID: "davezzzz", EstimatedDbSizeKb: kb(9000), MaxDbSizeKbLimit: kb(1024)}, writesRefused: "Nothing more is stored", want: []Banner{
			{ID: "site--banner--storage-full", Level: Error, Title: "Storage full:", Text: "Nothing more is stored"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, OfSite(context.Background(), &tt.site, tt.writesRefused, nil))
		})
	}
	require.Equal(t, []Banner{
		{ID: "site--banner--1", Level: Info, Text: "For every site"},
		{ID: "site--banner--2", Level: Warning, Text: "For alice <b>"},
	}, OfSite(context.Background(), &controldb.Site{ID: "alicezzz"}, "", written))

	// What an admin wrote is text on the page, never HTML
	var buf bytes.Buffer
	require.NoError(t, Strip("site--banners", OfSite(context.Background(), &controldb.Site{ID: "alicezzz"}, "", written), true).Render(context.Background(), &buf))
	require.Contains(t, buf.String(), "For alice &lt;b&gt;")
	// Only the information can be hidden, not the warning, and nothing on a page that hides none
	require.Contains(t, buf.String(), `id="site--banner--1--hide"`)
	require.NotContains(t, buf.String(), `id="site--banner--2--hide"`)
	buf.Reset()
	require.NoError(t, Strip("site--banners", OfSite(context.Background(), &controldb.Site{ID: "alicezzz"}, "", written), false).Render(context.Background(), &buf))
	require.NotContains(t, buf.String(), "--hide")
	buf.Reset()
	require.NoError(t, Strip("site--banners", nil, true).Render(context.Background(), &buf))
	require.Empty(t, buf.String())
}

// A warning and an error stay: only a level under them can be hidden
func TestCanHide(t *testing.T) {
	require.True(t, Info.CanHide())
	require.False(t, Warning.CanHide())
	require.False(t, Error.CanHide())
	require.False(t, Level("other").CanHide())
}
