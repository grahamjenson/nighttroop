package sitemodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// SiteSummaryPeriod is a day, week, fortnight or month of the site's data from Start until End, in milliseconds, in
// the settings' units. Its time in range, daily dose and daily carbs are the means of its days with data, and its
// glucose the mean of those days every 5 minutes from midnight, a day's own readings for a day. Its AGP is each 15
// minutes from midnight, the 5th, 25th, 50th, 75th and 95th percentiles of its days' glucose then, empty when there
// was none; a day's is of the 7 days to it, as one day has too few readings for percentiles. Its loop IOB (total, basal
// and treatment) and COB are every 5 minutes from midnight too, a day's newest loop report's then, null when none.
type SiteSummaryPeriod struct {
	Period          string        `json:"period"`
	Date            string        `json:"date"`
	Timezone        string        `json:"timezone"`
	Start           int64         `json:"start"`
	End             int64         `json:"end"`
	Days            int64         `json:"days"`
	TimeInRange     *float64      `json:"time_in_range"`
	TotalDailyDose  float64       `json:"total_daily_dose"`
	TotalDailyCarbs float64       `json:"total_daily_carbs"`
	Glucose         []*float64    `json:"glucose"`
	AGP             [][]float64   `json:"agp"`
	Units           *consts.Units `json:"units"`

	LoopIob          []*float64 `json:"loop_iob"`
	LoopIobBasal     []*float64 `json:"loop_iob_basal"`
	LoopIobTreatment []*float64 `json:"loop_iob_treatment"`
	LoopCob          []*float64 `json:"loop_cob"`
}

// SummaryPeriodFromDB is the site model of a summary period, with its glucose and AGP in the site context's units
func SummaryPeriodFromDB(row *sitedb.SummaryPeriod, siteContext *httputils.SiteContext) *SiteSummaryPeriod {
	day := SummaryDailyFromDB(row, siteContext)
	model := &SiteSummaryPeriod{Period: row.Period, Date: row.Date, Timezone: row.Timezone, Start: row.StartAt, End: row.EndAt,
		Days: row.Days, TimeInRange: row.TimeInRange, TotalDailyDose: row.TotalDailyDose, TotalDailyCarbs: row.TotalDailyCarbs,
		Glucose: day.Glucose, AGP: [][]float64{}, Units: day.Units}
	var agp [][]int64
	if row.AgpMgdlJson != nil {
		_ = json.Unmarshal([]byte(*row.AgpMgdlJson), &agp)
	}
	for i, field := range []*[]*float64{&model.LoopIob, &model.LoopIobBasal, &model.LoopIobTreatment, &model.LoopCob} {
		_ = json.Unmarshal([]byte(*loopColumns(row)[i]), field)
	}
	for _, percentiles := range agp {
		values := []float64{}
		for _, mgdl := range percentiles {
			values = append(values, day.Units.FromMGDL(float64(mgdl)))
		}
		model.AGP = append(model.AGP, values)
	}
	return model
}
