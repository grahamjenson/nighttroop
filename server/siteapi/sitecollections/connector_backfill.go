package sitecollections

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/metrics"
)

// BackfillHandler starts a backfill of the site's connector in the URL, answering with what it will fetch. A source
// with one queued or running is not started again, nor is any over the site's backfill limit.
func BackfillHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.ConnectorSettingsCollection{}).GetName(), can.ActionUpdate)
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
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.backfill_off")))
		return
	}
	// All the site's connectors share one backfill limit, so a site cannot hold the connector workers, or its connectors'
	// hosts, with backfills. Only a backfill that starts spends it, so a click on one already running does not.
	// Given back at the time it was reserved, as one given back later is not
	now := time.Now()
	limit := siteContext.Config.RateLimiters.SiteBackfillLimiter(siteContext.SiteID).ReserveN(now, 1)
	if wait := limit.DelayFrom(now); !limit.OK() || wait > 0 {
		limit.CancelAt(now)
		metrics.RateLimited()
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(int(wait.Seconds())+1))
		return
	}

	backfill, err := connectors.BackfillRange(r.Context(), siteContext.SiteDB, connectorType, siteContext.Config.BackfillDays, siteContext.GetSiteTimezone())
	if err == nil {
		err = connectors.QueueBackfill(r.Context(), siteContext.ControlDB, siteContext.SiteID, connectorType, backfill)
	}
	if err != nil {
		limit.CancelAt(now)
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
	WriteResponse(w, r, siteContext, http.StatusAccepted, backfill, site_settings_templ.BackfillQueued(connectorType, backfill.From, backfill.To, backfill.RecentFrom), false)
}
