package sitecollections

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/httputils"
)

// DownloadDataHandler sends a site's admin a snapshot of the site's DB, as it is now, to look through its data. Each
// copies the whole DB, so a site has only a few a minute.
func DownloadDataHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), siteapiutils.ActionGet)
	if siteContext == nil {
		return
	}
	if allowed, wait := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SiteDownloadLimiter(siteContext.SiteID)); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(wait))
		return
	}

	file, err := os.CreateTemp("", "site-download-*.sqlite")
	if err == nil {
		defer os.Remove(file.Name())
		defer file.Close()
		err = backup.SqliteBackup(r.Context(), siteContext.SiteDB.SQLiteDB, file.Name())
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to snapshot a site's data", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.sqlite"`, siteContext.SiteID, time.Now().UTC().Format("2006-01-02")))
	http.ServeContent(w, r, "", time.Now(), file)
}
