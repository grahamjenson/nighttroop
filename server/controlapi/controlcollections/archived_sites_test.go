package controlcollections

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestHydratingRefused checks only an archived site is hydrated, to live, and only with S3, which the tests do not have
func TestHydratingRefused(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	controlContext := &httputils.ControlContext{ControlDB: controlDB, IdentityModel: &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}}
	siteID := site.SiteDB.Subdomain
	for _, test := range []struct {
		name, from, to, message string
	}{
		{"only to live", "archived", "disabled", "an archived site's state can only be changed to live"},
		{"only an archived site", "live", "live", "only an archived site can be hydrated"},
		{"only with S3", "archived", "live", "a site can only be hydrated when S3 is connected"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := (&ArchivedSitesCollection{}).UpdateRow(t.Context(), controlContext, &controlmodels.Site{ID: siteID, State: test.from}, &controlmodels.Site{State: test.to})
			require.ErrorContains(t, err, test.message)
		})
	}
}
