package sitecollections

import (
	"archive/zip"
	"fmt"
	"io"
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

// DownloadDataHandler sends a site's admin a zip of the site's data, as it is now: a snapshot of its DB, to look through,
// in a zip so more files can be added beside it later. Each copies the whole DB, so a site has only a few a minute.
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
	name := fmt.Sprintf("%s-%s", siteContext.SiteID, time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, name))
	if err := writeDownloadZip(w, file, name+".sqlite"); err != nil {
		// The zip is already being sent, so the download is cut short rather than answered with an error
		slog.ErrorContext(r.Context(), "failed to send a site's data", "error", err)
	}
}

// writeDownloadZip writes a zip of the DB snapshot, named dbName, to w
func writeDownloadZip(w io.Writer, db *os.File, dbName string) error {
	archive := zip.NewWriter(w)
	entry, err := archive.CreateHeader(&zip.FileHeader{Name: dbName, Method: zip.Deflate, Modified: time.Now()})
	if err == nil {
		_, err = io.Copy(entry, db)
	}
	if err != nil {
		return err
	}
	return archive.Close()
}
