package sitecollections

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
)

// BackfillHandler starts a backfill of the site's sync source in the URL, answering with what it will fetch. A source
// with one queued or running is not started again.
func BackfillHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.SyncSettingsCollection{}).GetName(), siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	syncType, err := consts.NewSyncSourceTypeFromString(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	source, err := siteContext.ControlDB.Queries.GetSyncSource(r.Context(), controldb.GetSyncSourceParams{Type: syncType, SiteID: siteContext.SiteID})
	if err != nil || source == nil || !source.Enabled {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint("turn on and save this sync before backfilling it"))
		return
	}

	backfill, err := sync.BackfillRange(r.Context(), siteContext.SiteDB, syncType, siteContext.Config.BackfillDays)
	if err == nil {
		err = sync.QueueBackfill(r.Context(), siteContext.ControlDB, siteContext.SiteID, syncType, backfill)
	}
	if errors.Is(err, sync.ErrBackfillRunning) {
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusConflict, Message: "A backfill of this sync is already running"})
		return
	}
	if err != nil {
		slog.Error("failed to start a backfill", "error", err, "sync_type", syncType)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusAccepted, backfill, site_settings_templ.BackfillQueued(syncType, backfill.From, backfill.To), false)
}
