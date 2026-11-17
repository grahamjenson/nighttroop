package sitecollections

import (
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/i18n"
)

// DeleteSiteHandler lets a site's admin delete their site. It is archived: its last writes are sent to its replica in
// S3 and it is deleted here, so it is no longer served, then the admin is signed out and sent to Nighttroop's home page.
// Without S3 it is refused.
func DeleteSiteHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), can.ActionUpdate)
	if siteContext == nil {
		return
	}
	if !siteContext.Config.Status.S3Working {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.delete_needs_s3")))
		return
	}
	started := time.Now()
	if err := siteContext.ControlDB.ArchiveSite(r.Context(), siteContext.SiteID); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete a site", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	siteContext.ControlDB.Logs.SiteArchivedEvent(r.Context(), siteContext.SiteID, "site admin", started)

	siteContext.Config.Security.ClearUserLoginCookie(w)
	w.Header().Set("HX-Redirect", siteContext.Config.HostURLWithPort)
	w.WriteHeader(http.StatusOK)
}
