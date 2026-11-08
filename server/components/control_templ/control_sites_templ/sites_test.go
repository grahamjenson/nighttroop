package control_sites_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/consts"
)

// TestSiteActionsTruncate checks the Truncate button shows only for a site over its DB size limit, which the browser
// tests cannot make as their DBs are in memory
func TestSiteActionsTruncate(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		name     string
		sizeKB   *int64
		truncate bool
	}{
		{"over its limit", size(4000), true},
		{"within its limit", size(1000), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			site := &controlmodels.Site{ID: "site", State: "live", EstimatedDbSizeKb: test.sizeKB, MaxDbSizeKbLimit: size(3000)}
			require.NoError(t, siteActions(site).Render(context.Background(), &buf))
			require.Equal(t, test.truncate, bytes.Contains(buf.Bytes(), []byte(`id="control--site--truncate"`)), buf.String())
		})
	}
}

// A site's DB size, kept in KB, is shown in MB
func TestSizeMB(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		kb   *int64
		want string
	}{
		{nil, "-"},
		{size(0), "0.00"},
		{size(40), "0.04"},
		{size(1024), "1.00"},
		{size(3000), "2.93"},
		{size(512000), "500.00"},
	} {
		require.Equal(t, test.want, sizeMB(test.kb))
	}
}

// The DB size bar is as wide as the share of its limit the site uses, at most all of it, in the theme's colour, amber
// nearing the limit and red over it; a size not measured has no bar
func TestDBSizeBar(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		name   string
		sizeKB *int64
		want   []string
	}{
		{"half", size(1500), []string{`style="width: 50%;"`, "bg-primary", "50% of its limit used"}},
		{"nearing it", size(2700), []string{`style="width: 90%;"`, "bg-amber-500", "90% of its limit used"}},
		{"over it", size(4500), []string{`style="width: 100%;"`, "bg-red-500", "150% of its limit used"}},
		{"not measured", nil, []string{"Not measured yet"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			site := &controlmodels.Site{ID: "site", State: "live", EstimatedDbSizeKb: test.sizeKB, MaxDbSizeKbLimit: size(3000)}
			require.NoError(t, dbSize(site).Render(context.Background(), &buf))
			for _, want := range test.want {
				require.Contains(t, buf.String(), want)
			}
			require.Equal(t, test.sizeKB != nil, bytes.Contains(buf.Bytes(), []byte(`id="control--site--db-used"`)))
		})
	}
}

// A site's identities are counted as people, admins and users, and devices
func TestIdentityTypes(t *testing.T) {
	identities := func(roles ...consts.IdentityRole) []*controlmodels.SiteIdentity {
		list := []*controlmodels.SiteIdentity{}
		for _, role := range roles {
			list = append(list, &controlmodels.SiteIdentity{Role: role})
		}
		return list
	}
	require.Equal(t, "1 person, 2 devices", identityTypes(identities(consts.IdentityRoleAdmin, consts.IdentityRoleDeviceRead, consts.IdentityRoleDevicePWDPhone)))
	require.Equal(t, "2 people, 1 device", identityTypes(identities(consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleDeviceRead)))
	require.Equal(t, "0 people, 0 devices", identityTypes(nil))
}
