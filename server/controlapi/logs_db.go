package controlapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// downloadingLogsDB is held while a copy of the logs DB is made and sent, so there is only one at a time, filling the
// disk and CPU no more than once
var downloadingLogsDB sync.Mutex

// errDownloadingLogsDB refuses a download of the logs DB while another is being made or sent
var errDownloadingLogsDB = &apiutils.HTTPError{Status: http.StatusTooManyRequests, Message: "The logs DB is already being downloaded, try again once it is done"}

// logsDBWriteTimeout is how long sending the copy can take, past the server's usual write timeout, as staging's captured
// requests can make the logs DB hundreds of MB
const logsDBWriteTimeout = 10 * time.Minute

// LogsDBHandler sends the control admin a copy of the logs DB, as logs-<time>.sqlite, to read the requests staging
// captured. It is copied with SQLite's backup API, so it is consistent while logs are written, into a temp file that is
// deleted once sent. It holds every site's logs, so only an admin can. It is long lived (httputils.LongLived), so a
// large one is not cut off by the request timeout.
func LogsDBHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	if controlContext.IdentityModel.Role != consts.ControlRoleAdmin {
		apiutils.ErrForbidden.Write(acceptType, w)
		return
	}
	if !downloadingLogsDB.TryLock() {
		errDownloadingLogsDB.Write(acceptType, w)
		return
	}
	defer downloadingLogsDB.Unlock()

	// Copied beside the logs DB's backup path, which is otherwise unused, on the same disk as the DBs rather than a
	// small temp dir
	logs := controlContext.ControlDB.Logs
	err := utils.MkParentDirs(logs.BackupPath)
	var file *os.File
	if err == nil {
		file, err = os.CreateTemp(filepath.Dir(logs.BackupPath), ".logs-download-*.sqlite")
	}
	if err == nil {
		defer os.Remove(file.Name())
		defer file.Close()
		err = backup.SqliteBackup(r.Context(), logs.SQLiteDB, file.Name())
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to copy the logs DB", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	slog.InfoContext(r.Context(), "Control admin downloaded the logs DB", "identity", controlContext.IdentityModel.Name)
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(logsDBWriteTimeout))
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="logs-%s.sqlite"`, time.Now().UTC().Format("2006-01-02T15-04-05Z")))
	http.ServeContent(w, r, "", time.Now(), file)
}
