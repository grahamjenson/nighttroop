package sitedata

import (
	"context"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
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

// defaultHours is how far back the summary goes when the request does not say
const defaultHours = 48

// maxRows caps each list at a week of readings five minutes apart
const maxRows = 7 * 24 * 12

// maxDevices caps how many devices' latest statuses are shown
const maxDevices = 20

func (s *SummaryCollection) ToModel(row *sitemodels.Summary, siteContext *httputils.SiteContext) *sitemodels.Summary {
	if row == nil {
		return nil
	}
	return row
}

func (s *SummaryCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Summary, error) {
	return getSummary(ctx, collectionTx, defaultHours, siteContext)
}

func getSummary(ctx context.Context, tx *siteapiutils.CollectionTx, hours int64, siteContext *httputils.SiteContext) (*sitemodels.Summary, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()
	summary := &sitemodels.Summary{Hours: hours, Since: since, Units: sitemodels.ContextUnits(siteContext)}
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
	return summary, nil
}
