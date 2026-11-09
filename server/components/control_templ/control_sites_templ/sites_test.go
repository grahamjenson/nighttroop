package control_sites_templ

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestSiteActionsTruncate checks the Truncate button shows only for a site over its DB size limit, and otherwise when it
// would be offered, which the browser tests cannot make as their DBs are in memory
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
			admin := httputils.SetAdminControlContext(httputils.CreateControlContext(context.Background(), nil, "", 0),
				&controldb.ControlIdentity{Role: consts.ControlRoleAdmin}, config.ServerStatus{})
			require.NoError(t, siteActions(site).Render(admin, &buf))
			require.Equal(t, test.truncate, bytes.Contains(buf.Bytes(), []byte(`id="control--site--truncate"`)), buf.String())
			require.Equal(t, !test.truncate, bytes.Contains(buf.Bytes(), []byte("Offered while it is over its DB size limit.")), buf.String())
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
	tables := []*controlmodels.SiteTable{{Name: "glucose_readings", Bytes: 3 << 20}, {Name: "treatments", Bytes: 1 << 20}}
	for _, test := range []struct {
		name   string
		sizeKB *int64
		tables []*controlmodels.SiteTable
		want   []string
		bar    bool
	}{
		{"half", size(1500), tables, []string{`style="width: 50%;"`, "50% of its limit used", `id="control--site--db-table--glucose_readings"`, "3.00 MB · 75%"}, true},
		{"nearing it", size(2700), nil, []string{`style="width: 90%;"`, "text-amber-800", "90% of its limit used"}, true},
		{"over it", size(4500), nil, []string{`style="width: 100%;"`, "text-red-600", "150% of its limit used"}, true},
		{"measured by its tables", nil, tables, []string{">4.00</span>", `style="width: 100%;"`, "136% of its limit used", `id="control--site--db-part--treatments"`}, true},
		{"not measured", nil, nil, []string{"Not measured yet", "The resource check measures each site's DB"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			site := &controlmodels.Site{ID: "site", State: "live", EstimatedDbSizeKb: test.sizeKB, MaxDbSizeKbLimit: size(3000)}
			require.NoError(t, dbSize(site, test.tables).Render(context.Background(), &buf))
			for _, want := range test.want {
				require.Contains(t, buf.String(), want)
			}
			require.Equal(t, test.bar, bytes.Contains(buf.Bytes(), []byte(`id="control--site--db-used"`)))
			require.Equal(t, test.tables != nil, bytes.Contains(buf.Bytes(), []byte(`id="control--site--db-tables"`)))
		})
	}
}

// The largest tables are shaded darkest to lightest, by their share of the DB, and the rest are Other
func TestDBParts(t *testing.T) {
	tables := []*controlmodels.SiteTable{}
	for i, bytes := range []int64{40, 20, 15, 10, 5, 6, 4} {
		tables = append(tables, &controlmodels.SiteTable{Name: fmt.Sprintf("t%d", i), Bytes: bytes, Records: bytes * 100})
	}
	parts := dbParts(tables)
	require.Len(t, parts, 6)
	require.Equal(t, dbPart{Name: "t0", Bytes: 40, Records: 4000, Percent: 40, Class: "bg-brand-night"}, parts[0])
	require.Equal(t, dbPart{Name: "t4", Bytes: 5, Records: 500, Percent: 5, Class: "bg-brand-night/25"}, parts[4])
	// Other's records are its tables' together
	require.Equal(t, dbPart{Name: "Other", Bytes: 10, Records: 1000, Percent: 10, Class: "bg-gray-300"}, parts[5])
	require.Equal(t, "4.0k · 0.00 MB · 40%", keyText(parts[0]))
	require.Len(t, dbParts(tables[:2]), 2)
	require.Empty(t, dbParts(nil))
	require.Empty(t, dbParts([]*controlmodels.SiteTable{{Name: "empty"}}))
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
