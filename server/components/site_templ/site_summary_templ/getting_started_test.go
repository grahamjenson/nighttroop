package site_summary_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// Only a site with no glucose for over a month, or ever, is new enough to be shown the guide in place of its summary
func TestIsNew(t *testing.T) {
	inRange := 80.0
	require.True(t, isNew(&sitemodels.Summary{}))
	require.False(t, isNew(&sitemodels.Summary{GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{}}}))
	require.False(t, isNew(&sitemodels.Summary{DeviceStatuses: []*sitemodels.SiteDeviceStatus{{}}}))
	require.False(t, isNew(&sitemodels.Summary{LastMonth: sitemodels.SummaryDay{TimeInRange: &inRange}}), "readings last month")
}

// Admins are guided to get glucose in and add people and devices, which only they can do; everyone to settings and alarms
func TestGettingStartedSteps(t *testing.T) {
	for role, admin := range map[consts.IdentityRole]bool{consts.IdentityRoleAdmin: true, consts.IdentityRoleUser: false, consts.IdentityRoleUserReadOnly: false} {
		t.Run(string(role), func(t *testing.T) {
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, nil, "", "", httputils.HTTPActionRead, httputils.AcceptTypeHTML, "")
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: role})
			var buf bytes.Buffer
			require.NoError(t, GettingStarted().Render(ctx, &buf))
			html := buf.String()
			require.Equal(t, admin, bytes.Contains(buf.Bytes(), []byte(`id="site--getting-started--data"`)), html)
			require.Equal(t, admin, bytes.Contains(buf.Bytes(), []byte(`id="site--getting-started--identities"`)), html)
			require.Equal(t, !admin, bytes.Contains(buf.Bytes(), []byte("Your site&#39;s admin sets up")), html)
			require.Contains(t, html, `id="site--getting-started--settings"`)
			require.Contains(t, html, `id="site--getting-started--alarms"`)
		})
	}
}
