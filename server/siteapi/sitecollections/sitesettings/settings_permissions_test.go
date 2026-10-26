package sitesettings_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/consts"
)

// Sync sources, identities and the site's settings are the admins' alone; an identity's own settings are its own
func TestSettingsCollectionPermissions(t *testing.T) {
	admin := &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}
	user := &controldb.SiteIdentity{Role: consts.IdentityRoleUser}
	for _, collection := range []interface{ GetName() string }{&sitesettings.SyncSettingsCollection{}, &sitesettings.IdentitiesCollection{}, &sitesettings.SettingsCollection{Site: true}} {
		for _, action := range []siteapiutils.Action{siteapiutils.ActionSearch, siteapiutils.ActionCreate, siteapiutils.ActionUpdate, siteapiutils.ActionDelete} {
			require.True(t, siteapiutils.HasPermission(collection.GetName(), action, admin), "%s %s", collection.GetName(), action)
			require.False(t, siteapiutils.HasPermission(collection.GetName(), action, user), "%s %s", collection.GetName(), action)
		}
	}
	own := (&sitesettings.SettingsCollection{}).GetName()
	require.True(t, siteapiutils.HasPermission(own, siteapiutils.ActionUpdate, user))
}
