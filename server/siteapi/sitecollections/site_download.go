package sitecollections

import (
	"log/slog"
	"net/http"

	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/httputils"
)

// DownloadDataHandler sends a site's admin the site's download (backup.SiteDownload), a zip of a snapshot of its DB, as
// the control site's Download DB does. Each copies the whole DB, so a site has only a few a minute.
func DownloadDataHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), siteapiutils.ActionGet)
	if siteContext == nil {
		return
	}
	if allowed, wait := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SiteDownloadLimiter(siteContext.SiteID)); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(wait))
		return
	}

	download, err := backup.SnapshotSite(r.Context(), siteContext.SiteDB.SQLiteDB, siteContext.SiteID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to snapshot a site's data", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	defer download.Close()
	if err := download.Send(w); err != nil {
		slog.ErrorContext(r.Context(), "failed to send a site's data", "error", err)
	}
}
