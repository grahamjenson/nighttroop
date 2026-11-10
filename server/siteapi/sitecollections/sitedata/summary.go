package sitedata

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_summary_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type SummaryCollection struct{}

// ////
// Shared
// ////
func (s *SummaryCollection) GetName() string {
	return "summary"
}

// GetRoute is where the collection is in its URL
func (s *SummaryCollection) GetRoute() string {
	return "summary"
}

// defaultHours is how far back the summary goes when the request does not say, as when a site is signed in to
const defaultHours = 6

// maxRows caps each list at a week of readings five minutes apart
const maxRows = 7 * 24 * 12

// maxDevices caps how many devices' latest statuses are shown
const maxDevices = 20

// maxSwitchMinutes is how long before the summary a profile switch with a duration is looked for, as one still in force
const maxSwitchMinutes = 7 * 24 * 60

// Watches is whether a change is to something the summary shows, so its events stream tells the summary page to update.
// An alarm is the identity's own, as the stream sends an identity only its own.
func (s *SummaryCollection) Watches(event events.Event, _ *httputils.SiteContext) bool {
	switch event.Model.(type) {
	case *sitedb.GlucoseReading, *sitedb.InsulinDose, *sitedb.FoodIntake, *sitedb.TempTarget, *sitedb.TempBasal,
		*sitedb.DeviceReport, *sitedb.ProfileSwitch, *sitedb.ProfilePreset, *sitedb.ComponentEvent,
		*sitedb.EventGroupsBundle, *sitedb.ProfilePresetBundle, *livedb.Alarm:
		return true
	}
	return false
}

func (s *SummaryCollection) ToModel(row *sitemodels.Summary, siteContext *httputils.SiteContext) *sitemodels.Summary {
	if row == nil {
		return nil
	}
	return row
}

func (s *SummaryCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.Summary, error) {

	hours := int64(defaultHours)
	if query.Hours != nil {
		hours = *query.Hours
	}
	summary, err := getSummary(ctx, collectionTx, hours, siteContext)
	if err != nil {
		return nil, err
	}
	return []*sitemodels.Summary{summary}, nil
}

func (s *SummaryCollection) SearchHTMXResponse(models []*sitemodels.Summary, query *apiutils.Query) templ.Component {
	if len(models) == 0 {
		return nil
	}
	model := models[0]
	return site_summary_templ.Summary(model)
}

func (s *SummaryCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Summary, error) {
	return getSummary(ctx, collectionTx, defaultHours, siteContext)
}

func (s *SummaryCollection) GetHTMXResponse(model *sitemodels.Summary) templ.Component {
	return site_summary_templ.Summary(model)
}

func getSummary(ctx context.Context, tx *siteapiutils.CollectionTx, hours int64, siteContext *httputils.SiteContext) (*sitemodels.Summary, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()
	summary := &sitemodels.Summary{Hours: hours, Since: since, Units: sitemodels.ContextUnits(siteContext)}
	settings := siteContext.GetSettings()
	summary.RangeLow = summary.Units.FromMGDL(float64(*settings.GetRangeLowMgdl()))
	summary.RangeHigh = summary.Units.FromMGDL(float64(*settings.GetRangeHighMgdl()))
	summary.RangeUrgentLow = summary.Units.FromMGDL(float64(*settings.GetRangeUrgentLowMgdl()))
	summary.RangeUrgentHigh = summary.Units.FromMGDL(float64(*settings.GetRangeUrgentHighMgdl()))
	// A false IncludeDeleted leaves out deleted rows in these searches
	excludeDeleted := false

	var herr *apiutils.HTTPError
	readings, err := cleanGlucoseReadings(ctx, q, since, nil, siteContext)
	if err != nil {
		slog.ErrorContext(ctx, "Error searching summary glucose readings", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	summary.GlucoseReadings = readings[:min(len(readings), maxRows)]
	doses, err := q.SearchInsulinDoses(ctx, sitedb.SearchInsulinDosesParams{DateGTE: &since, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if summary.InsulinDoses, herr = apiutils.Searched(doses, err, "summary insulin doses", sitemodels.InsulinDoseFromDB); herr != nil {
		return nil, herr
	}
	food, err := q.SearchFoodIntakes(ctx, sitedb.SearchFoodIntakesParams{DateGTE: &since, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if summary.FoodIntakes, herr = apiutils.Searched(food, err, "summary food intakes", sitemodels.FoodIntakeFromDB); herr != nil {
		return nil, herr
	}
	targets, err := q.SearchTempTargets(ctx, sitedb.SearchTempTargetsParams{DateGTE: &since, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if summary.TempTargets, herr = apiutils.Searched(targets, err, "summary temp targets", apiutils.WithContext(sitemodels.TempTargetFromDB, siteContext)); herr != nil {
		return nil, herr
	}
	basals, err := q.SearchTempBasals(ctx, sitedb.SearchTempBasalsParams{DateGTE: &since, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if summary.TempBasals, herr = apiutils.Searched(basals, err, "summary temp basals", sitemodels.TempBasalFromDB); herr != nil {
		return nil, herr
	}

	devices, err := q.RecentDeviceNames(ctx, sitedb.RecentDeviceNamesParams{Since: since, Limit: maxDevices})
	if err != nil {
		slog.ErrorContext(ctx, "Error searching summary devices", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	summary.DeviceStatuses = []*sitemodels.SiteDeviceStatus{}
	for _, device := range devices {
		reports, err := q.DeviceReportsForDevice(ctx, sitedb.DeviceReportsForDeviceParams{Device: device, Limit: 1})
		latest, herr := apiutils.Searched(reports, err, "summary device statuses", sitemodels.DeviceStatusFromDB)
		if herr != nil {
			return nil, herr
		}
		summary.DeviceStatuses = append(summary.DeviceStatuses, latest...)
	}
	loop, err := q.GetLatestDeviceReportWithPWDInformation(ctx)
	if summary.Loop, herr = apiutils.Found(loop, err, "summary loop status", sitemodels.DeviceStatusFromDB); herr != nil {
		return nil, herr
	}
	if herr = addMinutes(ctx, q, summary, siteContext); herr != nil {
		return nil, herr
	}
	if herr = addDays(ctx, q, summary, siteContext); herr != nil {
		return nil, herr
	}
	if herr = addChanges(ctx, q, summary); herr != nil {
		return nil, herr
	}
	alarms, err := tx.Live().IdentityAlarms(ctx, livedb.IdentityAlarmsParams{SiteID: siteContext.SiteID, IdentityID: siteContext.IdentityModel.ID,
		IncludeSnoozed: true, Now: time.Now().UnixMilli()})
	if summary.Alarms, herr = apiutils.Searched(alarms, err, "summary alarms", apiutils.WithContext((&AlarmsCollection{}).ToModel, siteContext)); herr != nil {
		return nil, herr
	}
	return summary, nil
}

// addMinutes works out the targets and basal in force each minute of the summary. A profile sent without its own
// timezone is read in the site's, as the daily summaries are, not the viewer's.
func addMinutes(ctx context.Context, q *sitedb.Queries, summary *sitemodels.Summary, siteContext *httputils.SiteContext) (herr *apiutils.HTTPError) {
	summary.Minutes, herr = summaryMinutes(ctx, q, summary.Since, time.Now().UnixMilli(), summary.Units, siteContext.GetSiteTimezone(), siteContext, summary.TempTargets, summary.TempBasals)
	return herr
}

// summaryMinutes are the targets and basal in force each minute from since until until, in location's time, from the
// profiles, profile switches, temp targets and temp basals in force through it. The temp targets and basals from since
// are given, and those that started before it are looked up; without a site context only the basal is wanted, so
// earlier temp targets are not.
func summaryMinutes(ctx context.Context, q *sitedb.Queries, since, until int64, units *consts.Units, location *time.Location, siteContext *httputils.SiteContext, targets []*sitemodels.SiteTempTarget, basals []*sitemodels.SiteTempBasal) ([]sitemodels.SummaryMinute, *apiutils.HTTPError) {
	excludeDeleted := false
	defaults, err := q.SearchSummaryDefaultProfiles(ctx, sitedb.SearchSummaryDefaultProfilesParams{Since: since, Until: until, Limit: maxRows})
	if err != nil {
		slog.ErrorContext(ctx, "Error searching summary profiles", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	switches, err := q.SearchSummaryProfileSwitches(ctx, sitedb.SearchSummaryProfileSwitchesParams{Since: since, Until: until, Earliest: since - maxSwitchMinutes*60_000, Limit: maxRows})
	if err != nil {
		slog.ErrorContext(ctx, "Error searching summary profile switches", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	targetsSince, basalsSince, before := since-maxTempTargetMinutes*60_000, since-maxTempBasalMinutes*60_000, since-1
	var earlier []*sitemodels.SiteTempTarget
	if siteContext != nil {
		earlierTargets, err := q.SearchTempTargets(ctx, sitedb.SearchTempTargetsParams{DateGTE: &targetsSince, DateLTE: &before, IncludeDeleted: &excludeDeleted, Limit: maxRows})
		var herr *apiutils.HTTPError
		if earlier, herr = apiutils.Searched(earlierTargets, err, "summary temp targets", apiutils.WithContext(sitemodels.TempTargetFromDB, siteContext)); herr != nil {
			return nil, herr
		}
	}
	earlierBasals, err := q.SearchTempBasals(ctx, sitedb.SearchTempBasalsParams{DateGTE: &basalsSince, DateLTE: &before, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	earlierBasal, herr := apiutils.Searched(earlierBasals, err, "summary temp basals", sitemodels.TempBasalFromDB)
	if herr != nil {
		return nil, herr
	}
	slices.Reverse(defaults)
	slices.Reverse(switches)
	return sitemodels.SummaryMinutes(since, until, units, location, defaults, switches,
		append(earlier, targets...), append(earlierBasal, basals...)), nil
}

// addDays is today so far, yesterday, last week and last month, as days from midnight in the site's timezone: today
// built up to now, and the others from their summaries, or built when those are not yet built
func addDays(ctx context.Context, q *sitedb.Queries, summary *sitemodels.Summary, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	now, location := time.Now(), siteContext.GetSiteTimezone()
	today, yesterday := LocalDate(now, location, 0), LocalDate(now, location, -1)
	todayRow, err := buildSummaryDaily(ctx, q, siteContext, today, now.UnixMilli())
	if err != nil {
		slog.ErrorContext(ctx, "Error building the summary's today", "error", err)
		return apiutils.ErrInternalServerError
	}
	yesterdayRow, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: yesterday})
	if errors.Is(err, sql.ErrNoRows) {
		yesterdayRow, err = buildSummaryDaily(ctx, q, siteContext, yesterday, now.UnixMilli())
	}
	if err != nil {
		slog.ErrorContext(ctx, "Error finding the summary's yesterday", "error", err)
		return apiutils.ErrInternalServerError
	}
	summary.Today, summary.Yesterday = sitemodels.SummaryDayFromPeriod(todayRow), sitemodels.SummaryDayFromPeriod(yesterdayRow)
	summary.Today.End = now.UnixMilli()
	for period, day := range map[string]*sitemodels.SummaryDay{"week": &summary.LastWeek, "fortnight": &summary.LastFortnight, "month": &summary.LastMonth} {
		row, _, err := LastSummaryPeriod(ctx, q, siteContext, period, now)
		if err != nil {
			slog.ErrorContext(ctx, "Error finding the summary's last "+period, "error", err)
			return apiutils.ErrInternalServerError
		}
		*day = sitemodels.SummaryDayFromPeriod(row)
	}
	return nil
}

// addChanges finds when the latest CGM sensor was started or changed, and when the pump site was last changed
func addChanges(ctx context.Context, q *sitedb.Queries, summary *sitemodels.Summary) *apiutils.HTTPError {
	excludeDeleted := false
	latest := func(types string) (*int64, *apiutils.HTTPError) {
		events, err := q.SearchComponentEvents(ctx, sitedb.SearchComponentEventsParams{TypeIn: &types, IncludeDeleted: &excludeDeleted, Limit: 1})
		if err != nil {
			slog.ErrorContext(ctx, "Error searching summary component events", "error", err)
			return nil, apiutils.ErrInternalServerError
		}
		if len(events) == 0 {
			return nil, nil
		}
		return &events[0].OccurredAt, nil
	}
	var herr *apiutils.HTTPError
	if summary.SensorStartedAt, herr = latest(`["Sensor Start","Sensor Change"]`); herr != nil {
		return herr
	}
	summary.SiteChangedAt, herr = latest(`["Site Change"]`)
	return herr
}
