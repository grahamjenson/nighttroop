package can

// site_report.go: the clinician, who sees only the printable report of the site.

import (
	"nighttroop.t1d.nz/utils/consts"
)

// IdentityOnlySeeReport is whether the role sees only the printable report, its home page: a clinician. All other pages
// open the report, and the server forbids all other requests. site_stores.go has the summaries a clinician reads.
func IdentityOnlySeeReport(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleClinician
}
