package siteapiutils_test

import (
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
