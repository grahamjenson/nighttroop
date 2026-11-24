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
	Date           string   `json:"date"`
	Timezone       string   `json:"timezone"`
	Start          int64    `json:"start"`
	End            int64    `json:"end"`
	TimeInRange    *float64 `json:"time_in_range"`
	TotalDailyDose float64  `json:"total_daily_dose"`
	// TotalDailyBasal is the basal of its daily dose; the rest is boluses
	TotalDailyBasal float64       `json:"total_daily_basal"`
	TotalCarbs      float64       `json:"total_carbs"`
	Glucose         []*float64    `json:"glucose"`
	Units           *consts.Units `json:"units"`
	CreatedAt       int64         `json:"created_at"`
	UpdatedAt       int64         `json:"updated_at"`
}

// SummaryDailyFromDB is the site model of a summary day, with its glucose in the site context's units
func SummaryDailyFromDB(row *sitedb.SummaryPeriod, siteContext *httputils.SiteContext) *SiteSummaryDaily {
	units := ContextUnits(siteContext)
	model := &SiteSummaryDaily{Date: row.Date, Timezone: row.Timezone, Start: row.StartAt, End: row.EndAt, TimeInRange: row.TimeInRange,
		TotalDailyDose: row.TotalDailyDose, TotalDailyBasal: row.TotalDailyBasal, TotalCarbs: row.TotalDailyCarbs, Glucose: []*float64{}, Units: units,
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

// NewSummaryDaily is the day named date from start until end, as it is at until, built from readings, doses, food,
// loop reports and minutes that may run outside it: its glucose, then its insulin and carbs, then its loop IOB and COB.
func NewSummaryDaily(date, timezone string, start, end, until int64, low, high float64, readings []*glucose.Clean, doses []*sitedb.InsulinDose, food []*sitedb.FoodIntake, reports []*sitedb.DeviceReportsLoopBetweenRow, minutesSince int64, minutes []SummaryMinute) *sitedb.SummaryPeriod {
	day := summaryDay{row: &sitedb.SummaryPeriod{Period: "day", Date: date, Timezone: timezone, StartAt: start, EndAt: end},
		start: start, end: end, slots: (end - start + SummaryDailySlotMs - 1) / SummaryDailySlotMs}
	day.addGlucoseStats(readings, low, high)
	day.addGlucoseSlots(readings)
	day.addInsulin(doses, until, minutesSince, minutes)
	day.addCarbs(food)
	if day.row.GlucoseReadings > 0 || day.row.TotalDailyDose > 0 || day.row.TotalDailyCarbs > 0 {
		day.row.Days = 1
	}
	day.addLoop(reports)
	return day.row
}

// summaryDay builds the row of a summary day from start until end, in its slots of SummaryDailySlotMs
type summaryDay struct {
	row        *sitedb.SummaryPeriod
	start, end int64
	slots      int64
}

func (d summaryDay) within(at int64) bool { return at >= d.start && at < d.end }

func (d summaryDay) slot(at int64) int64 { return (at - d.start) / SummaryDailySlotMs }

// addGlucoseStats sets the day's count of readings, its times very low, low, in range, high and very high, the range
// being from low to high mg/dL, and their mean and standard deviation
func (d summaryDay) addGlucoseStats(readings []*glucose.Clean, low, high float64) {
	// bands count the readings very low, low, in range, high and very high
	bands, sum, squares := [5]int{}, 0.0, 0.0
	for _, r := range readings {
		if d.within(r.Row.OccurredAt) {
			bands[rangeBand(r.Row.Mgdl, low, high)]++
			sum, squares = sum+r.Row.Mgdl, squares+r.Row.Mgdl*r.Row.Mgdl
		}
	}
	for _, r := range bands {
		d.row.GlucoseReadings += int64(r)
	}
	count := float64(d.row.GlucoseReadings)
	if count == 0 {
		return
	}
	for i, column := range timeColumns(d.row) {
		*column = new(consts.Round(100*float64(bands[i])/count, 1))
	}
	mean := sum / count
	d.row.GlucoseMeanMgdl, d.row.GlucoseSdMgdl = new(consts.Round(mean, 1)), new(consts.Round(math.Sqrt(max(squares/count-mean*mean, 0)), 1))
}

// addGlucoseSlots sets the day's meter readings, how many slots have a CGM reading, and its glucose: the newest
// reading in each slot, rounded, null where there is none
func (d summaryDay) addGlucoseSlots(readings []*glucose.Clean) {
	slots, cgm := make([]*int64, d.slots), make([]bool, d.slots)
	for _, r := range readings {
		if !d.within(r.Row.OccurredAt) {
			continue
		}
		slot := d.slot(r.Row.OccurredAt)
		if r.Row.ReadingType == "mbg" {
			d.row.MeterReadings++
		} else {
			cgm[slot] = true
		}
		// Readings are newest first, so a slot keeps its newest
		if slots[slot] == nil {
			slots[slot] = new(int64(math.Round(r.Row.Mgdl)))
		}
	}
	d.row.CgmSlots = int64(len(slices.DeleteFunc(cgm, func(read bool) bool { return !read })))
	glucoseJSON, _ := json.Marshal(slots)
	d.row.GlucoseMgdlJson = string(glucoseJSON)
}

// addInsulin sets the day's total daily dose: the doses' insulin in it given by until, so an extended dose counts as
// it is given, and its basal, the basal in force each minute of minutes from minutesSince. A dose of basal insulin, as
// the long acting insulin of someone on injections that AAPS and Juggluco send, is basal too, not a bolus.
func (d summaryDay) addInsulin(doses []*sitedb.InsulinDose, until, minutesSince int64, minutes []SummaryMinute) {
	for _, dose := range doses {
		if dose.IsBasalInsulin != nil && *dose.IsBasalInsulin {
			d.row.TotalDailyBasal += insulinWithin(dose, d.start, min(d.end, until))
			continue
		}
		d.row.TotalDailyDose += insulinWithin(dose, d.start, min(d.end, until))
	}
	for i, m := range minutes {
		if d.within(minutesSince+int64(i)*60_000) && m.Basal != nil {
			d.row.TotalDailyBasal += *m.Basal / 60
		}
	}
	d.row.TotalDailyDose = consts.Round(d.row.TotalDailyDose+d.row.TotalDailyBasal, 2)
	d.row.TotalDailyBasal = consts.Round(d.row.TotalDailyBasal, 2)
}

// addCarbs sets the day's total carbs, of the food in it
func (d summaryDay) addCarbs(food []*sitedb.FoodIntake) {
	for _, f := range food {
		if d.within(f.OccurredAt) && f.Carbs != nil {
			d.row.TotalDailyCarbs += *f.Carbs
		}
	}
	d.row.TotalDailyCarbs = consts.Round(d.row.TotalDailyCarbs, 1)
}

// addLoop sets the day's loop IOB, basal IOB, treatment IOB and COB: the newest report's, newest first, in each slot,
// or else in the slot before, rounded to 2 decimals, null where there is none
func (d summaryDay) addLoop(reports []*sitedb.DeviceReportsLoopBetweenRow) {
	loop, taken := [4][]*float64{}, make([]bool, d.slots)
	for i := range loop {
		loop[i] = make([]*float64, d.slots)
	}
	for _, r := range reports {
		slot := d.slot(r.OccurredAt)
		if !d.within(r.OccurredAt) || taken[slot] {
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
	for slot := len(taken) - 1; slot > 0; slot-- {
		for i := range loop {
			if !taken[slot] && taken[slot-1] {
				loop[i][slot] = loop[i][slot-1]
			}
		}
	}
	for i, column := range loopColumns(d.row) {
		loopJSON, _ := json.Marshal(loop[i])
		*column = string(loopJSON)
	}
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

// NewSummaryPeriod is the period, a week or month named by its first date, from start until end, of its summary days,
// as they are at until. Only days with data count. Its times in each range are of all their readings, so each day's are
// weighed by its readings and a day of only a few, such as today so far, counts for little. Its daily dose and carbs
// are the means of its days over by until, as a day not over has had only some of its insulin and food, or of all of
// them when none is over. Its glucose at each 5 minutes from midnight is the mean of the days with a reading then, as
// its loop IOB and COB are,
// rounded to 2 decimals. Its readings are all its days', with their mean and standard deviation. Its AGP is the
// percentiles of their glucose in the 15 minutes around each 5 minutes from midnight.
func NewSummaryPeriod(period, date, timezone string, start, end, until int64, days []*sitedb.SummaryPeriod) *sitedb.SummaryPeriod {
	row := &sitedb.SummaryPeriod{Period: period, Date: date, Timezone: timezone, StartAt: start, EndAt: end}
	times, weighed, sum, squares := [5]float64{}, 0.0, 0.0, 0.0
	// over is the totals of the days over by until, and all of every day with data
	over, all := dailyTotals{}, dailyTotals{}
	glucoseMeans, bins := newSlotMeans[int64](), make(agpBins, 24*60*60_000/AGPSlotMs)
	loopMeans := [4]slotMeans[float64]{}
	for i := range loopMeans {
		loopMeans[i] = newSlotMeans[float64]()
	}
	for _, day := range days {
		if day.Days == 0 {
			continue
		}
		row.Days++
		all.add(day)
		if day.EndAt <= until {
			over.add(day)
		}
		if day.TimeInRange != nil {
			weight := float64(max(day.GlucoseReadings, 1))
			for i, column := range timeColumns(day) {
				times[i] += weight * *cmp.Or(*column, new(0.0))
			}
			weighed += weight
		}
		row.GlucoseReadings += day.GlucoseReadings
		row.MeterReadings += day.MeterReadings
		row.CgmSlots += day.CgmSlots
		if day.GlucoseMeanMgdl != nil && day.GlucoseSdMgdl != nil {
			n, mean, sd := float64(day.GlucoseReadings), *day.GlucoseMeanMgdl, *day.GlucoseSdMgdl
			sum, squares = sum+n*mean, squares+n*(sd*sd+mean*mean)
		}
		bins.add(glucoseMeans.add(day.GlucoseMgdlJson))
		for i, column := range loopColumns(day) {
			loopMeans[i].add(*column)
		}
	}
	if weighed > 0 {
		for i, column := range timeColumns(row) {
			*column = new(consts.Round(times[i]/weighed, 1))
		}
	}
	if count := float64(row.GlucoseReadings); count > 0 && sum > 0 {
		mean := sum / count
		row.GlucoseMeanMgdl, row.GlucoseSdMgdl = new(consts.Round(mean, 1)), new(consts.Round(math.Sqrt(max(squares/count-mean*mean, 0)), 1))
	}
	if over.days == 0 {
		over = all
	}
	over.fill(row)
	row.GlucoseMgdlJson = glucoseMeans.json(func(mean float64) int64 { return int64(math.Round(mean)) })
	for i, column := range loopColumns(row) {
		*column = loopMeans[i].json(func(mean float64) float64 { return consts.Round(mean, 2) })
	}
	row.AgpMgdlJson = new(bins.json())
	return row
}

// dailyTotals sums the daily dose, basal and carbs of some of a period's days, to give their means
type dailyTotals struct {
	dose, basal, carbs float64
	days               int
}

func (t *dailyTotals) add(day *sitedb.SummaryPeriod) {
	t.dose, t.basal, t.carbs, t.days = t.dose+day.TotalDailyDose, t.basal+day.TotalDailyBasal, t.carbs+day.TotalDailyCarbs, t.days+1
}

// fill sets the row's daily dose, basal and carbs to the means of the days, if there are any
func (t dailyTotals) fill(row *sitedb.SummaryPeriod) {
	if t.days == 0 {
		return
	}
	row.TotalDailyDose = consts.Round(t.dose/float64(t.days), 2)
	row.TotalDailyBasal = consts.Round(t.basal/float64(t.days), 2)
	row.TotalDailyCarbs = consts.Round(t.carbs/float64(t.days), 1)
}

// slotMeans sums the values days have at each 5 minutes from midnight, to give the mean of the days with a value then
type slotMeans[T int64 | float64] struct {
	sums   []T
	counts []int64
}

func newSlotMeans[T int64 | float64]() slotMeans[T] {
	slots := 24 * 60 * 60_000 / SummaryDailySlotMs
	return slotMeans[T]{sums: make([]T, slots), counts: make([]int64, slots)}
}

// add adds a day's JSON column of values, null where it has none, and returns the values in the day
func (m slotMeans[T]) add(column string) []*T {
	var values []*T
	_ = json.Unmarshal([]byte(column), &values)
	values = values[:min(len(values), len(m.sums))]
	for slot, value := range values {
		if value != nil {
			m.sums[slot] += *value
			m.counts[slot]++
		}
	}
	return values
}

// json is the JSON of the mean at each slot, rounded by round, null where no day has a value
func (m slotMeans[T]) json(round func(mean float64) T) string {
	means := make([]*T, len(m.sums))
	for slot, count := range m.counts {
		if count > 0 {
			means[slot] = new(round(float64(m.sums[slot]) / float64(count)))
		}
	}
	meansJSON, _ := json.Marshal(means)
	return string(meansJSON)
}

// agpBins gathers, for each 5 minutes from midnight, the glucose of the days in the agpWindow slots around it
type agpBins [][]int64

func (b agpBins) add(mgdl []*int64) {
	for slot, value := range mgdl {
		if value == nil {
			continue
		}
		for bin := max(slot-agpWindow/2, 0); bin <= min(slot+agpWindow/2, len(b)-1); bin++ {
			b[bin] = append(b[bin], *value)
		}
	}
}

// json is the JSON of each bin's AGPPercentiles of its glucose, null where it has none
func (b agpBins) json() string {
	agp := make([][]int64, len(b))
	for bin, values := range b {
		slices.Sort(values)
		for _, p := range AGPPercentiles {
			if len(values) > 0 {
				agp[bin] = append(agp[bin], int64(math.Round(glucose.Percentile(values, p))))
			}
		}
	}
	agpJSON, _ := json.Marshal(agp)
	return string(agpJSON)
}

// insulinWithin is a dose's units given from start until end: its dose given at once, when it is then, and the share
// of its extended dose, spread evenly over its extended duration, that is then, so one over midnight is split between
// the days, and a day not yet over has what is due in it. An extended dose with no duration is given at once.
func insulinWithin(d *sitedb.InsulinDose, start, end int64) float64 {
	units := 0.0
	if d.OccurredAt >= start && d.OccurredAt < end {
		units = d.Dose
	}
	if d.ExtendedDose == nil || *d.ExtendedDose <= 0 {
		return units
	}
	if d.ExtendedDurationMs == nil || *d.ExtendedDurationMs <= 0 {
		if d.OccurredAt >= start && d.OccurredAt < end {
			units += *d.ExtendedDose
		}
		return units
	}
	if overlap := min(d.OccurredAt+*d.ExtendedDurationMs, end) - max(d.OccurredAt, start); overlap > 0 {
		units += *d.ExtendedDose * float64(overlap) / float64(*d.ExtendedDurationMs)
	}
	return units
}
