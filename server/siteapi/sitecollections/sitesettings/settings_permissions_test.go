package sitesettings_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/consts"
)

// Connectors, identities and the site's settings are the admins' alone; an identity's own settings are its own
func TestSettingsCollectionPermissions(t *testing.T) {
	admin := &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}
	user := &controldb.SiteIdentity{Role: consts.IdentityRoleUser}
	for _, collection := range []interface{ GetName() string }{&sitesettings.ConnectorSettingsCollection{}, &sitesettings.IdentitiesCollection{}, &sitesettings.SettingsCollection{Site: true}} {
		for _, action := range []can.Action{can.ActionSearch, can.ActionCreate, can.ActionUpdate, can.ActionDelete} {
			require.True(t, can.IdentityAccessStore(admin, collection.GetName(), action), "%s %s", collection.GetName(), action)
			require.False(t, can.IdentityAccessStore(user, collection.GetName(), action), "%s %s", collection.GetName(), action)
		}
	}
	own := (&sitesettings.SettingsCollection{}).GetName()
	require.True(t, can.IdentityAccessStore(user, own, can.ActionUpdate))
}
