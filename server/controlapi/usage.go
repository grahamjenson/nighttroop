package controlapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_usage_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// usageRoutes is how many of the most asked for routes the usage page shows
const usageRoutes = 25

// UsageHandler is how many use the server and how much, and its resources, which its page asks for every 10 seconds
func UsageHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	q, config, now := controlContext.ControlDB.Queries, controlContext.ControlDB.ServerConfig, time.Now()
	day, week := now.Add(-24*time.Hour).UnixMilli(), now.AddDate(0, 0, -7).UnixMilli()
	ctx := r.Context()
	activeSites := func(since int64) (int64, error) {
		return q.CountActiveSites(ctx, controldb.CountActiveSitesParams{Since: &since})
	}
	activeIdentities := func(since int64) (int64, error) {
		return q.CountActiveIdentities(ctx, controldb.CountActiveIdentitiesParams{Since: &since})
	}
	created := func(since time.Time) (int64, error) {
		return q.CountSitesCreated(ctx, controldb.CountSitesCreatedParams{Since: since.UnixMilli()})
	}
	midnight, monday, firstOfMonth := periodStarts(now)
	model := controlmodels.Usage{Now: now.UnixMilli(), SitesByState: map[string]int64{}}
	var errs [8]error
	var states []*controldb.CountSitesByStateRow
	model.ActiveSitesDay, errs[0] = activeSites(day)
	model.ActiveSitesWeek, errs[1] = activeSites(week)
	model.ActiveIdentitiesDay, errs[2] = activeIdentities(day)
	model.ActiveIdentitiesWeek, errs[3] = activeIdentities(week)
	states, errs[4] = q.CountSitesByState(ctx)
	model.SitesCreatedToday, errs[5] = created(midnight)
	model.SitesCreatedWeek, errs[6] = created(monday)
	model.SitesCreatedMonth, errs[7] = created(firstOfMonth)
	if err := errors.Join(errs[:]...); err != nil {
		slog.ErrorContext(r.Context(), "failed to count the server's usage", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	for _, state := range states {
		model.SitesByState[state.State] = state.Count
	}
	model.SinceStart = config.Usage.Snapshot(now, usageRoutes)
	model.Server = config.Usage.Server(config.DBDir(), now)
	model.ResourcesSince = now.Add(-db.KeepResourceReadings).UnixMilli()
	model.MemoryAlertPercent, model.DiskAlertPercent = config.MemoryAlertPercent, config.DiskAlertPercent
	readings, err := controlContext.ControlDB.Live.Queries.ResourceReadingsSince(ctx, livedb.ResourceReadingsSinceParams{Since: model.ResourcesSince})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to read the server's resource checks", "error", err)
	}
	for _, reading := range readings {
		model.ResourceReadings = append(model.ResourceReadings, controlmodels.ResourceReading{At: reading.At, CPUPercent: reading.CpuPercent,
			CPUPeggedSeconds: reading.CpuPeggedSeconds, MemoryPercent: reading.MemoryPercent, DiskPercent: reading.DiskPercent})
	}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, model, control_usage_templ.Usage(model))
}

// periodStarts are the starts of now's day, week and month in UTC: midnight, the Monday and the 1st
func periodStarts(now time.Time) (time.Time, time.Time, time.Time) {
	utc := now.UTC()
	midnight := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	monday := midnight.AddDate(0, 0, -(int(midnight.Weekday())+6)%7)
	return midnight, monday, time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
}
