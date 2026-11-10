package site_reports_templ

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// viewedFrom is a context of a site in one timezone, viewed by an identity whose own settings are in another
func viewedFrom(siteTimezone, viewerTimezone string) context.Context {
	var action httputils.HTTPAction
	var accept httputils.AcceptType
	var content httputils.ContentType
	ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{ID: "site", Timezone: &siteTimezone}, "site", "", action, accept, content)
	return httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "viewer", Timezone: &viewerTimezone})
}

// A report's days are the site's, as its summaries are built from midnight in the site's timezone, whatever the
// timezone of whoever views them
func TestReportsInSiteTimezone(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	start := time.Date(2026, 11, 2, 0, 0, 0, 0, auckland)
	week := &sitemodels.SiteSummaryPeriod{Period: "week", Date: "2026-11-02", Start: start.UnixMilli(), End: start.AddDate(0, 0, 7).UnixMilli()}
	ctx := viewedFrom("Pacific/Auckland", "America/Los_Angeles")

	first, _ := days(ctx, week)
	require.Equal(t, "2026-11-02", first.Format(time.DateOnly))
	require.Equal(t, "2 Nov – 8 Nov 2026", title(ctx, week))

	weeks := profileWeeks(ctx, week, []*sitemodels.SiteSummaryDaily{{Date: "2026-11-02"}}, reportLocation(ctx))
	require.Equal(t, "2026-11-02", weeks[0][0].Date.Format(time.DateOnly))
	require.NotNil(t, weeks[0][0].Day, "the site's first day is found by its date")
}
