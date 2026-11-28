package sitedata

import (
	"context"
	"math"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/settings"
)

// Packaging a site's day builds its treatment series here, where profiles and temp basals are read, for the DB package
// that packages, which cannot import this one
func init() {
	db.TreatmentDaySeries = treatmentDaySeries
}

// treatmentDaySeries are the treatment series of the site's UTC day that starts at start, in ms, each 5 minutes, from
// its rows, as a day's summary works them out: the basal rate in force, the mean of its minutes, from the profile, its
// switches and the temp basals; the insulin of its boluses, an extended one as it is given, with the part of one that
// started before the day; the basal insulin injected; and the carbs eaten. A step with no insulin or carbs has none,
// not 0, and one with no profile has no rate.
func treatmentDaySeries(ctx context.Context, q *sitedb.Queries, site *controldb.Site, start int64) ([]db.DaySeries, error) {
	end, excludeDeleted, step := start+24*time.Hour.Milliseconds(), false, retention.BasalRate.Step.Milliseconds()
	doses, err := q.SearchInsulinDoses(ctx, sitedb.SearchInsulinDosesParams{DateGTE: &start, DateLT: &end, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	extended, err := q.InsulinDosesExtendedInto(ctx, sitedb.InsulinDosesExtendedIntoParams{Since: start - extendedLookback.Milliseconds(), At: start, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	food, err := q.SearchFoodIntakes(ctx, sitedb.SearchFoodIntakesParams{DateGTE: &start, DateLT: &end, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	basalRows, err := q.SearchTempBasals(ctx, sitedb.SearchTempBasalsParams{DateGTE: &start, DateLT: &end, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	basals, herr := apiutils.Searched(basalRows, err, "packaged temp basals", sitemodels.TempBasalFromDB)
	if herr != nil {
		return nil, herr
	}
	minutes, herr := summaryMinutes(ctx, q, start, end-time.Minute.Milliseconds(), new(consts.UnitsMGDL), settings.NewContextSettings(site, nil).SiteLocation(), nil, nil, basals)
	if herr != nil {
		return nil, herr
	}

	rate, bolus, injected, carbs := retention.BasalRate.Empty(), retention.InsulinBolus.Empty(), retention.InsulinBasal.Empty(), retention.Carbs.Empty()
	add := func(values []float64, i int, amount float64) {
		if amount == 0 {
			return
		}
		if math.IsNaN(values[i]) {
			values[i] = 0
		}
		values[i] += amount
	}
	for i := range rate {
		// The mean of the step's minutes that have a rate
		sum, with := 0.0, 0
		for _, minute := range minutes[i*5 : min(i*5+5, len(minutes))] {
			if minute.Basal != nil {
				sum, with = sum+*minute.Basal, with+1
			}
		}
		if with > 0 {
			rate[i] = sum / float64(with)
		}
		from := start + int64(i)*step
		for _, dose := range append(doses, extended...) {
			into := bolus
			if dose.IsBasalInsulin != nil && *dose.IsBasalInsulin {
				into = injected
			}
			add(into, i, sitemodels.InsulinWithin(dose, from, from+step))
		}
	}
	for _, f := range food {
		if f.Carbs != nil {
			add(carbs, int((f.OccurredAt-start)/step), *f.Carbs)
		}
	}
	return []db.DaySeries{{Series: retention.BasalRate, Values: rate}, {Series: retention.InsulinBolus, Values: bolus},
		{Series: retention.InsulinBasal, Values: injected}, {Series: retention.Carbs, Values: carbs}}, nil
}
