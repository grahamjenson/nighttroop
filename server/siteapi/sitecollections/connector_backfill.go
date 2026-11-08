package sitecollections

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
)

// BackfillHandler starts a backfill of the site's connector in the URL, answering with what it will fetch. A source
// with one queued or running is not started again.
func BackfillHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.ConnectorSettingsCollection{}).GetName(), siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	connectorType, err := consts.NewConnectorTypeFromString(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	source, err := siteContext.ControlDB.Queries.GetConnector(r.Context(), controldb.GetConnectorParams{Type: connectorType, SiteID: siteContext.SiteID})
	if err != nil || source == nil || !source.Enabled {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint("turn on and save this connector before backfilling it"))
		return
	}

	backfill, err := connectors.BackfillRange(r.Context(), siteContext.SiteDB, connectorType, siteContext.Config.BackfillDays)
	if err == nil {
		err = connectors.QueueBackfill(r.Context(), siteContext.ControlDB, siteContext.SiteID, connectorType, backfill)
	}
	if errors.Is(err, connectors.ErrBackfillRunning) {
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusConflict, Message: "A backfill of this connector is already running"})
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to start a backfill", "error", err, "connector_type", connectorType)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusAccepted, backfill, site_settings_templ.BackfillQueued(connectorType, backfill.From, backfill.To), false)
}
