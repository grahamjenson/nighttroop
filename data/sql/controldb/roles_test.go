package controldb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site or control identity cannot be stored with a role utils/can does not know, whichever way it is made, as each
// permission is decided only for the known roles; each known role is stored
func TestIdentityRolesKnown(t *testing.T) {
	_, control, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID := control.Queries, t.Context(), site.SiteDB.Subdomain
	siteIdentity := func(name string, role consts.IdentityRole) error {
		return q.CreateSiteIdentity(ctx, controldb.CreateSiteIdentityParams{SiteID: siteID, Name: name, PasswordHash: "hash",
			TotpSecretEnc: "secret", Role: role, CreatedByIdentityName: "test"})
	}
	controlIdentity := func(name string, role consts.ControlRole) error {
		return q.CreateControlIdentity(ctx, controldb.CreateControlIdentityParams{Name: name, PasswordHash: "hash",
			TotpSecretEnc: "secret", Role: role, CreatedByIdentityName: "test"})
	}
	for _, role := range []consts.IdentityRole{consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly,
		consts.IdentityRoleClinician, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite} {
		require.NoError(t, siteIdentity("known-"+string(role), role), role)
	}
	for _, role := range []consts.ControlRole{consts.ControlRoleAdmin, consts.ControlRoleRead} {
		require.NoError(t, controlIdentity("known-"+string(role), role), role)
	}

	for _, err := range []error{siteIdentity("unknown", "owner"), siteIdentity("empty", ""),
		controlIdentity("unknown", "owner"), controlIdentity("empty", "")} {
		require.Error(t, err)
		require.True(t, utilerrors.IsSQLConstraintCheck(err), err.Error())
	}
}
