package sitemodels

import (
	"encoding/json"
	"math"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/glucose"
	"nighttroop.t1d.nz/utils/httputils"
)

// SummaryDailySlotMs is how far apart a summary day's glucose values are, in milliseconds
const SummaryDailySlotMs = int64(5 * 60_000)

// SiteSummaryDaily is a day of the site's data, from midnight in its timezone: its time in range, total daily dose and
// carbs, and its glucose every 5 minutes from midnight, null when there was no reading, in the settings' units
type SiteSummaryDaily struct {
	Date           string        `json:"date"`
	Timezone       string        `json:"timezone"`
	Start          int64         `json:"start"`
	End            int64         `json:"end"`
	TimeInRange    *float64      `json:"time_in_range"`
	TotalDailyDose float64       `json:"total_daily_dose"`
	TotalCarbs     float64       `json:"total_carbs"`
	Glucose        []*float64    `json:"glucose"`
	Units          *consts.Units `json:"units"`
	CreatedAt      int64         `json:"created_at"`
	UpdatedAt      int64         `json:"updated_at"`
}

// SummaryDailyFromDB is the site model of a summary day, with its glucose in the site context's units
func SummaryDailyFromDB(row *sitedb.SummaryDaily, siteContext *httputils.SiteContext) *SiteSummaryDaily {
	units := ContextUnits(siteContext)
	model := &SiteSummaryDaily{Date: row.Date, Timezone: row.Timezone, Start: row.StartAt, End: row.EndAt, TimeInRange: row.TimeInRange,
		TotalDailyDose: row.TotalDailyDose, TotalCarbs: row.TotalCarbs, Glucose: []*float64{}, Units: units,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	var mgdl []*int64
	_ = json.Unmarshal([]byte(row.GlucoseMgdlJson), &mgdl)
	for _, value := range mgdl {
		if value == nil {
			model.Glucose = append(model.Glucose, nil)
		} else {
			model.Glucose = append(model.Glucose, new(units.FromMGDL(float64(*value))))
		}
	}
	return model
}

// SummaryDay is a day from Start until End, in milliseconds, with its time in range and total daily dose
type SummaryDay struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	// TimeInRange is the percent of the day's readings from the low alarm to the high alarm, null when it has none
	TimeInRange *float64 `json:"time_in_range"`
	// TotalDailyDose is the day's insulin doses and the basal in force through it, in U
	TotalDailyDose float64 `json:"total_daily_dose"`
}

// SummaryDayFromDaily is the summary's day of a summary day
func SummaryDayFromDaily(row *sitedb.SummaryDaily) SummaryDay {
	return SummaryDay{Start: row.StartAt, End: row.EndAt, TimeInRange: row.TimeInRange, TotalDailyDose: row.TotalDailyDose}
}

// NewSummaryDaily is the day named date from start until end, built from readings, doses, food and minutes that may
// run outside it. Its time in range is of the readings from low to high mg/dL, its total daily dose is the doses and
// the basal in force each minute from minutesSince, and its glucose is the newest reading in each 5 minutes.
func NewSummaryDaily(date, timezone string, start, end int64, low, high float64, readings []*glucose.Clean, doses []*sitedb.InsulinDose, food []*sitedb.FoodIntake, minutesSince int64, minutes []SummaryMinute) *sitedb.SummaryDaily {
	day := &sitedb.SummaryDaily{Date: date, Timezone: timezone, StartAt: start, EndAt: end}
	within := func(at int64) bool { return at >= start && at < end }
	slots := make([]*int64, (end-start+SummaryDailySlotMs-1)/SummaryDailySlotMs)
	count, inRange := 0, 0
	for _, r := range readings {
		if !within(r.Row.OccurredAt) {
			continue
		}
		count++
		if r.Row.Mgdl >= low && r.Row.Mgdl <= high {
			inRange++
		}
		// Readings are newest first, so a slot keeps its newest
		if slot := (r.Row.OccurredAt - start) / SummaryDailySlotMs; slots[slot] == nil {
			slots[slot] = new(int64(math.Round(r.Row.Mgdl)))
		}
	}
	if count > 0 {
		day.TimeInRange = new(consts.Round(100*float64(inRange)/float64(count), 1))
	}
	for _, d := range doses {
		if within(d.OccurredAt) {
			day.TotalDailyDose += d.Dose
		}
	}
	for i, m := range minutes {
		if within(minutesSince+int64(i)*60_000) && m.Basal != nil {
			day.TotalDailyDose += *m.Basal / 60
		}
	}
	day.TotalDailyDose = consts.Round(day.TotalDailyDose, 2)
	for _, f := range food {
		if within(f.OccurredAt) && f.Carbs != nil {
			day.TotalCarbs += *f.Carbs
		}
	}
	day.TotalCarbs = consts.Round(day.TotalCarbs, 1)
	glucoseJSON, _ := json.Marshal(slots)
	day.GlucoseMgdlJson = string(glucoseJSON)
	return day
}
