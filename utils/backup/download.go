package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// downloadWriteTimeout is how long sending a download can take, past the server's usual write timeout
const downloadWriteTimeout = 10 * time.Minute

// SiteDownload is a snapshot of a site's DB, in a temp file, to send as the site's download: a zip named for the site
// and the day, with the DB inside, so more files can be added beside it later. A site admin's Download Data and the
// control site's Download DB both send it, so the two are the same.
type SiteDownload struct {
	file *os.File
	name string
}

// SnapshotSite copies the site's DB, as it is now, for its download, which the caller closes
func SnapshotSite(ctx context.Context, db *sql.DB, siteID string) (*SiteDownload, error) {
	file, err := os.CreateTemp("", "site-download-*.sqlite")
	if err != nil {
		return nil, err
	}
	download := &SiteDownload{file: file, name: fmt.Sprintf("%s-%s", siteID, time.Now().UTC().Format(time.DateOnly))}
	if err := SqliteBackup(ctx, db, file.Name()); err != nil {
		download.Close()
		return nil, err
	}
	return download, nil
}

// Send writes the download, <site>-<date>.zip with <site>-<date>.sqlite inside. The zip is sent as it is made, so an
// error after it starts cuts the download short rather than answering with an error.
func (d *SiteDownload) Send(w http.ResponseWriter) error {
	// The snapshot took part of the server's 10 second write timeout, and a large DB takes longer to send, which would
	// cut the download short
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(downloadWriteTimeout))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, d.name))
	archive := zip.NewWriter(w)
	entry, err := archive.CreateHeader(&zip.FileHeader{Name: d.name + ".sqlite", Method: zip.Deflate, Modified: time.Now()})
	if err == nil {
		_, err = io.Copy(entry, d.file)
	}
	if err != nil {
		return err
	}
	return archive.Close()
}

// Close removes the snapshot
func (d *SiteDownload) Close() {
	d.file.Close()
	os.Remove(d.file.Name())
}
