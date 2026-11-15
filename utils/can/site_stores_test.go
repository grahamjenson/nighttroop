package can_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/consts"
)

// A clinician reads only the site's summaries, for its printable report: none of its data, settings or identities, and
// it changes nothing
func TestClinicianPermissions(t *testing.T) {
	clinician := &controldb.SiteIdentity{Role: consts.IdentityRoleClinician}
	tests := []struct {
		collection string
		action     can.Action
		allowed    bool
	}{
		{"summary_periods", can.ActionGet, true},
		{"summary_periods", can.ActionSearch, true},
		{"summary_periods", can.ActionCreate, false},
		{"summary_daily", can.ActionSearch, false},
		{"glucose_readings", can.ActionSearch, false},
		{"identity_settings", can.ActionGet, false},
		{"alarms", can.ActionSearch, false},
		{"identities", can.ActionSearch, false},
		{"site_logs", can.ActionSearch, false},
	}
	for _, test := range tests {
		require.Equal(t, test.allowed, can.IdentityAccessStore(clinician, test.collection, test.action), "%s %s", test.collection, test.action)
	}
}

// A Caregiver device reads everything and changes only treatments, as on the Nightscout API: not the readings, device
// statuses, profiles, food or loop settings the PWD phone uploads. The PWD phone changes them all
func TestCaregiverPermissions(t *testing.T) {
	caregiver := &controldb.SiteIdentity{Role: consts.IdentityRoleDeviceReadWrite}
	pwdPhone := &controldb.SiteIdentity{Role: consts.IdentityRoleDevicePWDPhone}
	reads := []can.Action{can.ActionSearch, can.ActionGet}
	changes := []can.Action{can.ActionCreate, can.ActionUpdate, can.ActionDelete}
	treatments := []string{"event_groups", "insulin_doses", "food", "temp_basals", "temp_targets", "profile_switches",
		"activities", "notification_events"}
	uploaded := []string{"glucose", "device_statuses", "profile_presets", "food_presets", "loop_settings"}
	own := []string{"identity_settings", "alarms", "push_registrations"}
	for _, collection := range append(append(treatments, uploaded...), own...) {
		for _, action := range reads {
			require.True(t, can.IdentityAccessStore(caregiver, collection, action), "%s %s", collection, action)
		}
		for _, action := range changes {
			require.True(t, can.IdentityAccessStore(pwdPhone, collection, action), "%s %s", collection, action)
			allowed := !slices.Contains(uploaded, collection)
			require.Equal(t, allowed, can.IdentityAccessStore(caregiver, collection, action), "%s %s", collection, action)
		}
	}
	require.False(t, can.IdentityAccessStore(caregiver, "identities", can.ActionSearch))
}
