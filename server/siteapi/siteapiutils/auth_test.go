package siteapiutils_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
)

// A clinician reads only the site's summaries, for its printable report: none of its data, settings or identities, and
// it changes nothing
func TestClinicianPermissions(t *testing.T) {
	clinician := &controldb.SiteIdentity{Role: consts.IdentityRoleClinician}
	tests := []struct {
		collection string
		action     siteapiutils.Action
		allowed    bool
	}{
		{"summary_periods", siteapiutils.ActionGet, true},
		{"summary_periods", siteapiutils.ActionSearch, true},
		{"summary_periods", siteapiutils.ActionCreate, false},
		{"summary_daily", siteapiutils.ActionSearch, false},
		{"glucose_readings", siteapiutils.ActionSearch, false},
		{"identity_settings", siteapiutils.ActionGet, false},
		{"alarms", siteapiutils.ActionSearch, false},
		{"identities", siteapiutils.ActionSearch, false},
		{"site_logs", siteapiutils.ActionSearch, false},
	}
	for _, test := range tests {
		require.Equal(t, test.allowed, siteapiutils.HasPermission(test.collection, test.action, clinician), "%s %s", test.collection, test.action)
	}
}

// A Caregiver device reads everything and changes only treatments, as on the Nightscout API: not the readings, device
// statuses, profiles, food or loop settings the PWD phone uploads. The PWD phone changes them all
func TestCaregiverPermissions(t *testing.T) {
	caregiver := &controldb.SiteIdentity{Role: consts.IdentityRoleDeviceReadWrite}
	pwdPhone := &controldb.SiteIdentity{Role: consts.IdentityRoleDevicePWDPhone}
	reads := []siteapiutils.Action{siteapiutils.ActionSearch, siteapiutils.ActionGet}
	changes := []siteapiutils.Action{siteapiutils.ActionCreate, siteapiutils.ActionUpdate, siteapiutils.ActionDelete}
	treatments := []string{"event_groups", "insulin_doses", "food", "temp_basals", "temp_targets", "profile_switches",
		"activities", "notification_events"}
	uploaded := []string{"glucose", "device_statuses", "profile_presets", "food_presets", "loop_settings"}
	own := []string{"identity_settings", "alarms", "push_registrations"}
	for _, collection := range append(append(treatments, uploaded...), own...) {
		for _, action := range reads {
			require.True(t, siteapiutils.HasPermission(collection, action, caregiver), "%s %s", collection, action)
		}
		for _, action := range changes {
			require.True(t, siteapiutils.HasPermission(collection, action, pwdPhone), "%s %s", collection, action)
			allowed := !slices.Contains(uploaded, collection)
			require.Equal(t, allowed, siteapiutils.HasPermission(collection, action, caregiver), "%s %s", collection, action)
		}
	}
	require.False(t, siteapiutils.HasPermission("identities", siteapiutils.ActionSearch, caregiver))
}
