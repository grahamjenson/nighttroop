package control_sites_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestSiteActionsTruncate checks the Truncate button shows for a site whose DB is here, whatever its size, and not for
// an archived one
func TestSiteActionsTruncate(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		name     string
		state    string
		sizeKB   *int64
		truncate bool
	}{
		{"over its limit", "live", size(4000), true},
		{"within its limit", "live", size(1000), true},
		{"disabled", "disabled", size(1000), true},
		{"archived", "archived", nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			site := &controlmodels.Site{ID: "site", State: test.state, EstimatedDbSizeKb: test.sizeKB, MaxDbSizeKbLimit: size(3000)}
			admin := httputils.SetAdminControlContext(httputils.CreateControlContext(context.Background(), nil, "", 0),
				&controldb.ControlIdentity{Role: consts.ControlRoleAdmin}, config.ServerStatus{})
			require.NoError(t, siteActions(site).Render(admin, &buf))
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
