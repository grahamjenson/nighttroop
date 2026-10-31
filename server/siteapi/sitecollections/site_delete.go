package sitecollections

import (
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/s3utils"
)

// DeleteSiteHandler lets a site's admin delete their site. It is archived: its data is uploaded to S3 and deleted here,
// so it is no longer served, then the admin is signed out and sent to Nighttroop's home page. Without S3 it is refused.
func DeleteSiteHandler(w http.ResponseWriter, r *http.Request) {
	deleteSite(w, r, nil)
}

// deleteSite archives the site with upload, or to S3 when upload is nil
func deleteSite(w http.ResponseWriter, r *http.Request, upload s3utils.UploadFunc) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	if upload == nil {
		if !siteContext.Config.Status.S3Working {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint("a site can only be deleted when its data can be archived to S3"))
			return
		}
		upload = s3utils.NewS3Client(siteContext.Config).UploadFilesToS3
	}
	started := time.Now()
	if err := siteContext.ControlDB.ArchiveSite(r.Context(), siteContext.SiteID, upload); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete a site", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	siteContext.ControlDB.Logs.SiteArchivedEvent(r.Context(), siteContext.SiteID, "site admin", started)

	siteContext.Config.Security.ClearUserLoginCookie(w)
	w.Header().Set("HX-Redirect", siteContext.Config.HostURLWithPort)
	w.WriteHeader(http.StatusOK)
}
