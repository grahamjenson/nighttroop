package sitedata

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_summary_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
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

// defaultHours is how far back the summary goes when the request does not say
const defaultHours = 48

// maxRows caps each list at a week of readings five minutes apart
const maxRows = 7 * 24 * 12

// maxDevices caps how many devices' latest statuses are shown
const maxDevices = 20

// maxSwitchMinutes is how long before the summary a profile switch with a duration is looked for, as one still in force
const maxSwitchMinutes = 7 * 24 * 60

// Watches is whether a change is to something the summary shows, so its events stream tells the summary page to update
func (s *SummaryCollection) Watches(event events.Event) bool {
	switch event.Model.(type) {
	case *sitedb.GlucoseReading, *sitedb.InsulinDose, *sitedb.FoodIntake, *sitedb.TempTarget, *sitedb.TempBasal,
		*sitedb.DeviceReport, *sitedb.ProfileSwitch, *sitedb.ProfilePreset,
		*sitedb.EventGroupsBundle, *sitedb.ProfilePresetBundle:
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
	summary.LowAlarm = summary.Units.FromMGDL(float64(*settings.GetAlarmLow()))
	summary.HighAlarm = summary.Units.FromMGDL(float64(*settings.GetAlarmHigh()))
	// A set IncludeDeleted leaves out deleted rows in these searches
	excludeDeleted := false

	var herr *apiutils.HTTPError
	readings, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: &since, Limit: maxRows})
	if summary.GlucoseReadings, herr = apiutils.Searched(readings, err, "summary glucose readings", apiutils.WithContext(sitemodels.GlucoseReadingFromDB, siteContext)); herr != nil {
		return nil, herr
	}
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
		slog.Error("Error searching summary devices", "error", err)
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
	return summary, nil
}

// addMinutes works out the targets and basal in force each minute of the summary, from the profiles, profile switches,
// temp targets and temp basals in force through it, including those that started before it
func addMinutes(ctx context.Context, q *sitedb.Queries, summary *sitemodels.Summary, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	since, now, excludeDeleted := summary.Since, time.Now().UnixMilli(), false
	defaults, err := q.SearchSummaryDefaultProfiles(ctx, sitedb.SearchSummaryDefaultProfilesParams{Since: since, Until: now, Limit: maxRows})
	if err != nil {
		slog.Error("Error searching summary profiles", "error", err)
		return apiutils.ErrInternalServerError
	}
	switches, err := q.SearchSummaryProfileSwitches(ctx, sitedb.SearchSummaryProfileSwitchesParams{Since: since, Until: now, Earliest: since - maxSwitchMinutes*60_000, Limit: maxRows})
	if err != nil {
		slog.Error("Error searching summary profile switches", "error", err)
		return apiutils.ErrInternalServerError
	}
	targetsSince, basalsSince, before := since-maxTempTargetMinutes*60_000, since-maxTempBasalMinutes*60_000, since-1
	earlierTargets, err := q.SearchTempTargets(ctx, sitedb.SearchTempTargetsParams{DateGTE: &targetsSince, DateLTE: &before, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	targets, herr := apiutils.Searched(earlierTargets, err, "summary temp targets", apiutils.WithContext(sitemodels.TempTargetFromDB, siteContext))
	if herr != nil {
		return herr
	}
	earlierBasals, err := q.SearchTempBasals(ctx, sitedb.SearchTempBasalsParams{DateGTE: &basalsSince, DateLTE: &before, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	basals, herr := apiutils.Searched(earlierBasals, err, "summary temp basals", sitemodels.TempBasalFromDB)
	if herr != nil {
		return herr
	}
	slices.Reverse(defaults)
	slices.Reverse(switches)
	summary.Minutes = sitemodels.SummaryMinutes(since, now, summary.Units, siteContext.GetSettings().Location(), defaults, switches,
		append(targets, summary.TempTargets...), append(basals, summary.TempBasals...))
	return nil
}
