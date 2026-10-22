package siteapiutils_test

import (
	"testing"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
)

var (
	everything = []siteapiutils.Action{siteapiutils.ActionSearch, siteapiutils.ActionGet, siteapiutils.ActionCreate, siteapiutils.ActionUpdate, siteapiutils.ActionDelete}
	reads      = []siteapiutils.Action{siteapiutils.ActionSearch, siteapiutils.ActionGet}
	nothing    = []siteapiutils.Action{}
	dataStores = []string{"glucose_readings", "insulin_doses", "temp_targets", "device_statuses", "event_groups", "summary"}
	adminOnly  = []string{"identities", "sync_settings", "site_settings"}
)

// Each role's allowed actions on data, on what only admins manage, and on its own settings
func TestHasPermission(t *testing.T) {
	cases := []struct {
		role                     consts.IdentityRole
		data, admin, ownSettings []siteapiutils.Action
	}{
		{consts.IdentityRoleAdmin, everything, everything, everything},
		{consts.IdentityRoleUser, everything, nothing, everything},
		{consts.IdentityRoleUserReadOnly, reads, nothing, everything},
		{consts.IdentityRoleDevicePWDPhone, everything, nothing, everything},
		{consts.IdentityRoleDeviceReadWrite, everything, nothing, everything},
		{consts.IdentityRoleDeviceRead, reads, nothing, everything},
		{consts.IdentityRole("superuser"), nothing, nothing, nothing},
	}
	for _, c := range cases {
		identity := &controldb.SiteIdentity{Role: c.role}
		check := func(stores []string, allowed []siteapiutils.Action) {
			for _, store := range stores {
				for _, action := range everything {
					want := false
					for _, a := range allowed {
						want = want || a == action
					}
					if got := siteapiutils.HasPermission(store, action, identity); got != want {
						t.Errorf("%s %s %s: got %v want %v", c.role, action, store, got, want)
					}
				}
			}
		}
		check(dataStores, c.data)
		check(adminOnly, c.admin)
		check([]string{"identity_settings"}, c.ownSettings)
	}
}

// Without an identity nothing is allowed
func TestHasPermissionNoIdentity(t *testing.T) {
	for _, store := range append(dataStores, adminOnly...) {
		for _, action := range everything {
			if siteapiutils.HasPermission(store, action, nil) {
				t.Errorf("nil identity may %s %s", action, store)
			}
		}
	}
}
