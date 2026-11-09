package sitemodels

import (
	"cmp"
	"encoding/json"
	"math"
	"slices"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/glucose"
	"nighttroop.t1d.nz/utils/httputils"
)

// SummaryDailySlotMs is how far apart a summary day's glucose values are, in milliseconds
const SummaryDailySlotMs = int64(5 * 60_000)

// AGPSlotMs is how far apart a week's or month's AGP percentiles are, each of the readings in the agpWindow slots of
// glucose around it, and AGPPercentiles are the percentiles
const (
	AGPSlotMs = SummaryDailySlotMs
	agpWindow = 3
)

// VeryLowMgdl and VeryHighMgdl are the glucose below and above which a reading is very low or very high, rather than
// low or high, the bounds the international consensus on time in range uses
const (
	VeryLowMgdl  = 54.0
	VeryHighMgdl = 250.0
)

var AGPPercentiles = []float64{5, 25, 50, 75, 95}

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
func SummaryDailyFromDB(row *sitedb.SummaryPeriod, siteContext *httputils.SiteContext) *SiteSummaryDaily {
	units := ContextUnits(siteContext)
	model := &SiteSummaryDaily{Date: row.Date, Timezone: row.Timezone, Start: row.StartAt, End: row.EndAt, TimeInRange: row.TimeInRange,
		TotalDailyDose: row.TotalDailyDose, TotalCarbs: row.TotalDailyCarbs, Glucose: []*float64{}, Units: units,
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
	// TimeInRange is the percent of the day's readings from the range low to the range high, null when it has none
	TimeInRange *float64 `json:"time_in_range"`
	// TotalDailyDose is the day's insulin doses and the basal in force through it, in U
	TotalDailyDose float64 `json:"total_daily_dose"`
	// Days is how many of its days had data, 0 or 1 for a day, so a 0 dose with none is not taken for a real one
	Days int64 `json:"days"`
}

// SummaryDayFromPeriod is the summary's day of a summary day, or the mean day of a week or month
func SummaryDayFromPeriod(row *sitedb.SummaryPeriod) SummaryDay {
	return SummaryDay{Start: row.StartAt, End: row.EndAt, TimeInRange: row.TimeInRange, TotalDailyDose: row.TotalDailyDose, Days: row.Days}
}

// NewSummaryDaily is the day named date from start until end, built from readings, doses, food, loop reports and
// minutes that may run outside it. Its time in range is of the readings from low to high mg/dL, and its times very
// low, low, high and very high of those outside it, with their mean and standard deviation. Its total daily dose is
// the doses and the basal in force each minute from minutesSince, its glucose is the newest reading in each 5 minutes,
// and its loop IOB and COB the newest report's, newest first, in each 5 minutes, or else in the 5 minutes before,
// rounded to 2 decimals.
func NewSummaryDaily(date, timezone string, start, end int64, low, high float64, readings []*glucose.Clean, doses []*sitedb.InsulinDose, food []*sitedb.FoodIntake, reports []*sitedb.DeviceReportsLoopBetweenRow, minutesSince int64, minutes []SummaryMinute) *sitedb.SummaryPeriod {
	day := &sitedb.SummaryPeriod{Period: "day", Date: date, Timezone: timezone, StartAt: start, EndAt: end}
	within := func(at int64) bool { return at >= start && at < end }
	slots := make([]*int64, (end-start+SummaryDailySlotMs-1)/SummaryDailySlotMs)
	// bands count the readings very low, low, in range, high and very high
	bands, sum, squares, cgm := [5]int{}, 0.0, 0.0, make([]bool, len(slots))
	for _, r := range readings {
		if !within(r.Row.OccurredAt) {
			continue
		}
		bands[rangeBand(r.Row.Mgdl, low, high)]++
		sum, squares = sum+r.Row.Mgdl, squares+r.Row.Mgdl*r.Row.Mgdl
		slot := (r.Row.OccurredAt - start) / SummaryDailySlotMs
		if r.Row.ReadingType == "mbg" {
			day.MeterReadings++
		} else {
			cgm[slot] = true
		}
		// Readings are newest first, so a slot keeps its newest
		if slots[slot] == nil {
			slots[slot] = new(int64(math.Round(r.Row.Mgdl)))
		}
	}
	for _, r := range bands {
		day.GlucoseReadings += int64(r)
	}
	day.CgmSlots = int64(len(slices.DeleteFunc(cgm, func(read bool) bool { return !read })))
	if count := float64(day.GlucoseReadings); count > 0 {
		percents := [5]*float64{}
		for i, r := range bands {
			percents[i] = new(consts.Round(100*float64(r)/count, 1))
		}
		day.TimeVeryLow, day.TimeLow, day.TimeInRange, day.TimeHigh, day.TimeVeryHigh = percents[0], percents[1], percents[2], percents[3], percents[4]
		mean := sum / count
		day.GlucoseMeanMgdl, day.GlucoseSdMgdl = new(consts.Round(mean, 1)), new(consts.Round(math.Sqrt(max(squares/count-mean*mean, 0)), 1))
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
			day.TotalDailyCarbs += *f.Carbs
		}
	}
	day.TotalDailyCarbs = consts.Round(day.TotalDailyCarbs, 1)
	if day.GlucoseReadings > 0 || day.TotalDailyDose > 0 || day.TotalDailyCarbs > 0 {
		day.Days = 1
	}
	glucoseJSON, _ := json.Marshal(slots)
	day.GlucoseMgdlJson = string(glucoseJSON)
	loop, taken := [4][]*float64{}, make([]bool, len(slots))
	for i := range loop {
		loop[i] = make([]*float64, len(slots))
	}
	for _, r := range reports {
		slot := (r.OccurredAt - start) / SummaryDailySlotMs
		if !within(r.OccurredAt) || taken[slot] {
			continue
		}
		taken[slot] = true
		for i, value := range []*float64{r.LoopIob, r.LoopIobBasal, r.LoopIobTreatment, r.LoopCob} {
			if value != nil {
				loop[i][slot] = new(consts.Round(*value, 2))
			}
		}
	}
	// Before the last hour a device keeps a report only every 10 minutes, so a report also stands for the 5 minutes
	// after its own when they have none
	for slot := len(slots) - 1; slot > 0; slot-- {
		for i := range loop {
			if !taken[slot] && taken[slot-1] {
				loop[i][slot] = loop[i][slot-1]
			}
		}
	}
	for i, column := range loopColumns(day) {
		loopJSON, _ := json.Marshal(loop[i])
		*column = string(loopJSON)
	}
	return day
}

// rangeBand is which of very low, low, in range, high and very high mg/dL is in, the range being from low to high
func rangeBand(mgdl, low, high float64) int {
	switch {
	case mgdl < min(VeryLowMgdl, low):
		return 0
	case mgdl < low:
		return 1
	case mgdl <= high:
		return 2
	case mgdl <= max(VeryHighMgdl, high):
		return 3
	}
	return 4
}

// timeColumns are the summary's percents of its readings very low, low, in range, high and very high
func timeColumns(row *sitedb.SummaryPeriod) [5]**float64 {
	return [5]**float64{&row.TimeVeryLow, &row.TimeLow, &row.TimeInRange, &row.TimeHigh, &row.TimeVeryHigh}
}

// loopColumns are the summary's loop IOB, basal IOB, treatment IOB and COB JSON columns
func loopColumns(row *sitedb.SummaryPeriod) [4]*string {
	return [4]*string{&row.LoopIobJson, &row.LoopIobBasalJson, &row.LoopIobTreatmentJson, &row.LoopCobJson}
}

// NewSummaryPeriod is the period, a week or month named by its first date, from start until end, of its summary days.
// Only days with data count, so its times in each range, daily dose and daily carbs are the means of theirs, and its
// glucose at each 5 minutes from midnight the mean of the days with a reading then, as its loop IOB and COB are,
// rounded to 2 decimals. Its readings are all its days', with their mean and standard deviation. Its AGP is the
// percentiles of their glucose in the 15 minutes around each 5 minutes from midnight.
func NewSummaryPeriod(period, date, timezone string, start, end int64, days []*sitedb.SummaryPeriod) *sitedb.SummaryPeriod {
	row := &sitedb.SummaryPeriod{Period: period, Date: date, Timezone: timezone, StartAt: start, EndAt: end}
	times, withRange, sum, squares := [5]float64{}, 0, 0.0, 0.0
	sums, counts := make([]int64, 24*60*60_000/SummaryDailySlotMs), make([]int64, 24*60*60_000/SummaryDailySlotMs)
	bins := make([][]int64, 24*60*60_000/AGPSlotMs)
	loopSums, loopCounts := [4][]float64{}, [4][]int64{}
	for i := range loopSums {
		loopSums[i], loopCounts[i] = make([]float64, len(sums)), make([]int64, len(sums))
	}
	for _, day := range days {
		if day.Days == 0 {
			continue
		}
		row.Days++
		row.TotalDailyDose += day.TotalDailyDose
		row.TotalDailyCarbs += day.TotalDailyCarbs
		if day.TimeInRange != nil {
			for i, column := range timeColumns(day) {
				times[i] += *cmp.Or(*column, new(0.0))
			}
			withRange++
		}
		row.GlucoseReadings += day.GlucoseReadings
		row.MeterReadings += day.MeterReadings
		row.CgmSlots += day.CgmSlots
		if day.GlucoseMeanMgdl != nil && day.GlucoseSdMgdl != nil {
			n, mean, sd := float64(day.GlucoseReadings), *day.GlucoseMeanMgdl, *day.GlucoseSdMgdl
			sum, squares = sum+n*mean, squares+n*(sd*sd+mean*mean)
		}
		var mgdl []*int64
		_ = json.Unmarshal([]byte(day.GlucoseMgdlJson), &mgdl)
		for slot, value := range mgdl[:min(len(mgdl), len(sums))] {
			if value != nil {
				sums[slot] += *value
				counts[slot]++
				for bin := slot - agpWindow/2; bin <= slot+agpWindow/2; bin++ {
					if bin >= 0 && bin < len(bins) {
						bins[bin] = append(bins[bin], *value)
					}
				}
			}
		}
		for i, column := range loopColumns(day) {
			var values []*float64
			_ = json.Unmarshal([]byte(*column), &values)
			for slot, value := range values[:min(len(values), len(sums))] {
				if value != nil {
					loopSums[i][slot] += *value
					loopCounts[i][slot]++
				}
			}
		}
	}
	if withRange > 0 {
		for i, column := range timeColumns(row) {
			*column = new(consts.Round(times[i]/float64(withRange), 1))
		}
	}
	if count := float64(row.GlucoseReadings); count > 0 && sum > 0 {
		mean := sum / count
		row.GlucoseMeanMgdl, row.GlucoseSdMgdl = new(consts.Round(mean, 1)), new(consts.Round(math.Sqrt(max(squares/count-mean*mean, 0)), 1))
	}
	if row.Days > 0 {
		row.TotalDailyDose = consts.Round(row.TotalDailyDose/float64(row.Days), 2)
		row.TotalDailyCarbs = consts.Round(row.TotalDailyCarbs/float64(row.Days), 1)
	}
	slots := make([]*int64, len(sums))
	for slot, count := range counts {
		if count > 0 {
			slots[slot] = new(int64(math.Round(float64(sums[slot]) / float64(count))))
		}
	}
	glucoseJSON, _ := json.Marshal(slots)
	row.GlucoseMgdlJson = string(glucoseJSON)
	for i, column := range loopColumns(row) {
		means := make([]*float64, len(sums))
		for slot, count := range loopCounts[i] {
			if count > 0 {
				means[slot] = new(consts.Round(loopSums[i][slot]/float64(count), 2))
			}
		}
		loopJSON, _ := json.Marshal(means)
		*column = string(loopJSON)
	}
	agp := make([][]int64, len(bins))
	for bin, values := range bins {
		slices.Sort(values)
		for _, p := range AGPPercentiles {
			if len(values) > 0 {
				agp[bin] = append(agp[bin], int64(math.Round(glucose.Percentile(values, p))))
			}
		}
	}
	agpJSON, _ := json.Marshal(agp)
	row.AgpMgdlJson = new(string(agpJSON))
	return row
}
