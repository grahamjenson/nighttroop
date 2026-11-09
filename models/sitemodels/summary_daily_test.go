package sitemodels

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/glucose"
)

func TestNewSummaryDaily(t *testing.T) {
	const hour = int64(60 * 60_000)
	// The day is from 1 hour until 2 hours, 12 glucose slots, with the range from 70 to 180 mg/dL
	start, end := hour, 2*hour
	reading := func(at int64, mgdl float64) *glucose.Clean {
		return &glucose.Clean{Row: &sitedb.GlucoseReading{OccurredAt: at, Mgdl: mgdl}}
	}
	dose := func(at int64, units float64) *sitedb.InsulinDose { return &sitedb.InsulinDose{OccurredAt: at, Dose: units} }
	carbs := func(at int64, grams float64) *sitedb.FoodIntake { return &sitedb.FoodIntake{OccurredAt: at, Carbs: &grams} }
	// basal is a minute from minutesSince each, with the rate in force for minutes of them
	basal := func(rate float64, minutes int) []SummaryMinute {
		list := make([]SummaryMinute, minutes)
		for i := range list {
			list[i].Basal = &rate
		}
		return list
	}
	// slots is the day's 12 glucose slots, with values at the slots given
	slots := func(values map[int]int64) []*int64 {
		list := make([]*int64, 12)
		for i, v := range values {
			list[i] = &v
		}
		return list
	}

	tests := []struct {
		name         string
		readings     []*glucose.Clean
		doses        []*sitedb.InsulinDose
		food         []*sitedb.FoodIntake
		minutesSince int64
		minutes      []SummaryMinute
		inRange      *float64
		dose         float64
		carbs        float64
		glucose      []*int64
	}{
		{name: "nothing in the day", glucose: slots(nil)},
		{name: "readings on the range's edges are in it", readings: []*glucose.Clean{reading(start+1, 70), reading(start, 180)},
			inRange: new(100.0), glucose: slots(map[int]int64{0: 70})},
		{name: "readings below and above are out of it", readings: []*glucose.Clean{reading(start+2, 69), reading(start+1, 181), reading(start, 100)},
			inRange: new(33.3), glucose: slots(map[int]int64{0: 69})},
		{name: "readings outside the day are left out", readings: []*glucose.Clean{reading(end, 40), reading(start, 100), reading(start-1, 40)},
			inRange: new(100.0), glucose: slots(map[int]int64{0: 100})},
		{name: "each 5 minutes keeps its newest reading, rounded", readings: []*glucose.Clean{reading(end-1, 150.6), reading(start+5*60_000, 121), reading(start+4*60_000, 99)},
			inRange: new(100.0), glucose: slots(map[int]int64{1: 121, 0: 99, 11: 151})},
		{name: "doses in the day add up", doses: []*sitedb.InsulinDose{dose(start, 1.5), dose(end-1, 2), dose(end, 9), dose(start-1, 9)}, dose: 3.5, glucose: slots(nil)},
		{name: "the basal in force each minute of the day adds up", minutesSince: start, minutes: basal(1.2, 60), dose: 1.2, glucose: slots(nil)},
		{name: "basal outside the day is left out", minutesSince: 0, minutes: basal(0.6, 3*60), dose: 0.6, glucose: slots(nil)},
		{name: "doses and basal add up", doses: []*sitedb.InsulinDose{dose(start, 4)}, minutesSince: start, minutes: basal(1, 60), dose: 5, glucose: slots(nil)},
		{name: "carbs in the day add up", food: []*sitedb.FoodIntake{carbs(start, 30), carbs(end-1, 12.5), carbs(end, 99), {OccurredAt: start}},
			carbs: 42.5, glucose: slots(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day := NewSummaryDaily("1970-01-01", "UTC", start, end, 70, 180, tt.readings, tt.doses, tt.food, nil, tt.minutesSince, tt.minutes)
			if day.Period != "day" || day.Date != "1970-01-01" || day.Timezone != "UTC" || day.StartAt != start || day.EndAt != end {
				t.Errorf("day is a %s %s in %s from %d to %d", day.Period, day.Date, day.Timezone, day.StartAt, day.EndAt)
			}
			// A day with readings, insulin or carbs has data, so counts towards its week and month
			if withData := tt.inRange != nil || tt.dose > 0 || tt.carbs > 0; day.Days != map[bool]int64{true: 1}[withData] {
				t.Errorf("day has %d days with data, want data %v", day.Days, withData)
			}
			if !reflect.DeepEqual(day.TimeInRange, tt.inRange) {
				t.Errorf("time in range is %v, want %v", day.TimeInRange, tt.inRange)
			}
			if day.TotalDailyDose != tt.dose {
				t.Errorf("total daily dose is %v, want %v", day.TotalDailyDose, tt.dose)
			}
			if day.TotalDailyCarbs != tt.carbs {
				t.Errorf("total carbs is %v, want %v", day.TotalDailyCarbs, tt.carbs)
			}
			want, _ := json.Marshal(tt.glucose)
			if day.GlucoseMgdlJson != string(want) {
				t.Errorf("glucose is %s, want %s", day.GlucoseMgdlJson, want)
			}
		})
	}
}

// A day's loop IOB and COB are the newest loop report's in each 5 minutes, or else in the 5 minutes before, as a
// device keeps one every 10 minutes once an hour old, rounded, null where there was none
func TestNewSummaryDailyLoop(t *testing.T) {
	const hour = int64(60 * 60_000)
	start, end := hour, 2*hour
	report := func(at int64, iob, basal, treatment, cob *float64) *sitedb.DeviceReportsLoopBetweenRow {
		return &sitedb.DeviceReportsLoopBetweenRow{OccurredAt: at, LoopIob: iob, LoopIobBasal: basal, LoopIobTreatment: treatment, LoopCob: cob}
	}
	// Newest first, as DeviceReportsLoopBetween reads them: two in the first 5 minutes, none in the next, one in the
	// third without basal or treatment IOB, and one before and one after the day
	reports := []*sitedb.DeviceReportsLoopBetweenRow{
		report(end, new(9.0), nil, nil, new(9.0)),
		report(start+2*SummaryDailySlotMs+1, new(0.5), nil, nil, new(10.0)),
		report(start+4*60_000, new(2.345), new(0.5), new(1.845), new(30.0)),
		report(start, new(1.0), new(0.25), new(0.75), new(40.0)),
		report(start-1, new(9.0), nil, nil, new(9.0)),
	}
	day := NewSummaryDaily("1970-01-01", "UTC", start, end, 70, 180, nil, nil, nil, reports, start, nil)
	tail := strings.Repeat(",null", 8) + "]"
	for name, got := range map[string][2]string{
		"iob":           {day.LoopIobJson, "[2.35,2.35,0.5,0.5" + tail},
		"basal iob":     {day.LoopIobBasalJson, "[0.5,0.5,null,null" + tail},
		"treatment iob": {day.LoopIobTreatmentJson, "[1.85,1.85,null,null" + tail},
		"cob":           {day.LoopCobJson, "[30,30,10,10" + tail},
	} {
		if got[0] != got[1] {
			t.Errorf("%s is %s, want %s", name, got[0], got[1])
		}
	}
}

// A week's loop IOB and COB at each time of day are the means of its days' then, rounded
func TestNewSummaryPeriodLoop(t *testing.T) {
	day := func(withData int64, iob, cob string) *sitedb.SummaryPeriod {
		return &sitedb.SummaryPeriod{Days: withData, GlucoseMgdlJson: "[]", LoopIobJson: iob, LoopIobBasalJson: iob, LoopIobTreatmentJson: "[]", LoopCobJson: cob}
	}
	week := NewSummaryPeriod("week", "2026-10-26", "UTC", 1, 2, []*sitedb.SummaryPeriod{
		day(1, "[1,null,2.5]", "[10,20]"), day(1, "[2,null,null,-1]", "[null,25]"), day(1, "[1.005]", "[11]"), day(0, "[100]", "[100]")})
	tail := func(set int) string { return strings.Repeat(",null", 288-set) + "]" }
	for name, got := range map[string][2]string{
		"iob":           {week.LoopIobJson, "[1.34,null,2.5,-1" + tail(4)},
		"basal iob":     {week.LoopIobBasalJson, "[1.34,null,2.5,-1" + tail(4)},
		"treatment iob": {week.LoopIobTreatmentJson, "[null" + tail(1)},
		"cob":           {week.LoopCobJson, "[10.5,22.5" + tail(2)},
	} {
		if got[0] != got[1] {
			t.Errorf("%s is %s, want %s", name, got[0], got[1])
		}
	}
}

func TestNewSummaryPeriod(t *testing.T) {
	day := func(inRange *float64, dose, carbs float64) *sitedb.SummaryPeriod {
		withData := int64(0)
		if inRange != nil || dose > 0 || carbs > 0 {
			withData = 1
		}
		return &sitedb.SummaryPeriod{Days: withData, TimeInRange: inRange, TotalDailyDose: dose, TotalDailyCarbs: carbs, GlucoseMgdlJson: "[]"}
	}
	tests := []struct {
		name    string
		days    []*sitedb.SummaryPeriod
		count   int64
		inRange *float64
		dose    float64
		carbs   float64
	}{
		{name: "no days is no data", days: nil},
		{name: "a day without data is left out", days: []*sitedb.SummaryPeriod{day(nil, 0, 0)}},
		{name: "the means are of the days with data", days: []*sitedb.SummaryPeriod{day(new(50.0), 30, 100), day(new(75.0), 40, 150), day(nil, 0, 0)},
			count: 2, inRange: new(62.5), dose: 35, carbs: 125},
		{name: "a day with insulin but no readings counts for the dose, not the time in range",
			days: []*sitedb.SummaryPeriod{day(new(80.0), 30, 0), day(nil, 31, 0)}, count: 2, inRange: new(80.0), dose: 30.5},
		{name: "the means are rounded", days: []*sitedb.SummaryPeriod{day(new(10.0), 1, 1), day(new(20.0), 1, 1), day(new(20.0), 2, 0)},
			count: 3, inRange: new(16.7), dose: 1.33, carbs: 0.7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period := NewSummaryPeriod("week", "2026-10-26", "UTC", 1, 2, tt.days)
			if period.Period != "week" || period.Date != "2026-10-26" || period.Timezone != "UTC" || period.StartAt != 1 || period.EndAt != 2 {
				t.Errorf("period is %s %s in %s from %d to %d", period.Period, period.Date, period.Timezone, period.StartAt, period.EndAt)
			}
			if period.Days != tt.count || !reflect.DeepEqual(period.TimeInRange, tt.inRange) || period.TotalDailyDose != tt.dose || period.TotalDailyCarbs != tt.carbs {
				t.Errorf("period has %d days, %v in range, %v U and %v g a day, want %d, %v, %v and %v",
					period.Days, period.TimeInRange, period.TotalDailyDose, period.TotalDailyCarbs, tt.count, tt.inRange, tt.dose, tt.carbs)
			}
		})
	}
}

func TestNewSummaryPeriodGlucose(t *testing.T) {
	day := func(withData int64, glucose string) *sitedb.SummaryPeriod {
		return &sitedb.SummaryPeriod{Days: withData, GlucoseMgdlJson: glucose}
	}
	// at is the period's glucose at slots, the rest of its 288 null
	at := func(values map[int]int64) string {
		slots := make([]*int64, 288)
		for slot, value := range values {
			slots[slot] = new(value)
		}
		glucose, _ := json.Marshal(slots)
		return string(glucose)
	}
	// longDay is a day the clocks go back, 300 slots with a reading in its 288th and its 289th
	longDay := strings.TrimSuffix(at(map[int]int64{287: 80}), "]") + ",200" + strings.Repeat(",null", 11) + "]"
	tests := []struct {
		name string
		days []*sitedb.SummaryPeriod
		want string
	}{
		{name: "no days is no glucose", want: at(nil)},
		{name: "each time of day is the mean of the days with a reading then",
			days: []*sitedb.SummaryPeriod{day(1, "[100,null,120]"), day(1, "[200,150]"), day(1, "[null,null,null,90]")},
			want: at(map[int]int64{0: 150, 1: 150, 2: 120, 3: 90})},
		{name: "the mean is rounded", days: []*sitedb.SummaryPeriod{day(1, "[100]"), day(1, "[101]")}, want: at(map[int]int64{0: 101})},
		{name: "a day without data is left out", days: []*sitedb.SummaryPeriod{day(1, "[100]"), day(0, "[300]")}, want: at(map[int]int64{0: 100})},
		{name: "a day the clocks go back keeps only its first 24 hours", days: []*sitedb.SummaryPeriod{day(1, longDay)}, want: at(map[int]int64{287: 80})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSummaryPeriod("month", "2026-10-01", "UTC", 1, 2, tt.days).GlucoseMgdlJson; got != tt.want {
				t.Errorf("glucose is %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNewSummaryPeriodAGP(t *testing.T) {
	day := func(withData int64, glucose string) *sitedb.SummaryPeriod {
		return &sitedb.SummaryPeriod{Days: withData, GlucoseMgdlJson: glucose}
	}
	// agp is the 288 AGP slots, with the percentiles at the slots given and the rest null
	agp := func(slots map[int][]int64) string {
		all := make([][]int64, 288)
		for slot, percentiles := range slots {
			all[slot] = percentiles
		}
		value, _ := json.Marshal(all)
		return string(value)
	}
	tests := []struct {
		name string
		days []*sitedb.SummaryPeriod
		want string
	}{
		{name: "no days has no percentiles", want: agp(nil)},
		{name: "one reading is every percentile of its 5 minutes and those either side",
			days: []*sitedb.SummaryPeriod{day(1, "[null,120]")}, want: agp(map[int][]int64{0: {120, 120, 120, 120, 120}, 1: {120, 120, 120, 120, 120}, 2: {120, 120, 120, 120, 120}})},
		{name: "each 5 minutes is the readings of the 15 minutes around it on every day",
			days: []*sitedb.SummaryPeriod{day(1, "[100,110,120]"), day(1, "[130,null,140]")},
			// 100, 110 and 130 by the first 5 minutes, all five by the second, 110, 120 and 140 by the third, and 120 and
			// 140 by the fourth
			want: agp(map[int][]int64{0: {101, 105, 110, 120, 128}, 1: {102, 110, 120, 130, 138}, 2: {111, 115, 120, 130, 138}, 3: {121, 125, 130, 135, 139}})},
		{name: "a day without data is left out", days: []*sitedb.SummaryPeriod{day(1, "[100]"), day(0, "[300]")},
			want: agp(map[int][]int64{0: {100, 100, 100, 100, 100}, 1: {100, 100, 100, 100, 100}})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSummaryPeriod("week", "2026-10-26", "UTC", 1, 2, tt.days).AgpMgdlJson; got == nil || *got != tt.want {
				t.Errorf("AGP is %v, want %s", *got, tt.want)
			}
		})
	}
}

// A day's readings are counted into very low (below 54 mg/dL), low, in range, high and very high (above 250 mg/dL),
// with their mean and standard deviation, and its meter readings and 5 minutes with a CGM reading counted
func TestNewSummaryDailyGlucoseStats(t *testing.T) {
	const slot = SummaryDailySlotMs
	reading := func(at int64, mgdl float64, readingType string) *glucose.Clean {
		return &glucose.Clean{Row: &sitedb.GlucoseReading{OccurredAt: at, Mgdl: mgdl, ReadingType: readingType}}
	}
	readings := []*glucose.Clean{reading(5*slot, 300, "mbg"), reading(4*slot, 200, "sgv"), reading(3*slot+1, 100, "sgv"),
		reading(3*slot, 100, "sgv"), reading(1*slot, 60, "sgv"), reading(0, 40, "sgv"), reading(288*slot, 40, "sgv")}
	day := NewSummaryDaily("1970-01-01", "UTC", 0, 288*slot, 70, 180, readings, nil, nil, nil, 0, nil)
	got := []any{day.GlucoseReadings, day.MeterReadings, day.CgmSlots, *day.GlucoseMeanMgdl, *day.GlucoseSdMgdl,
		*day.TimeVeryLow, *day.TimeLow, *day.TimeInRange, *day.TimeHigh, *day.TimeVeryHigh}
	// 40, 60, 100, 100, 200 and 300 have a mean of 133.3 and a standard deviation of 89.9
	want := []any{int64(6), int64(1), int64(4), 133.3, 89.9, 16.7, 16.7, 33.3, 16.7, 16.7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("day's readings, meter readings, CGM slots, mean, SD and times are %v, want %v", got, want)
	}
	if empty := NewSummaryDaily("1970-01-01", "UTC", 0, 288*slot, 70, 180, nil, nil, nil, nil, 0, nil); empty.GlucoseMeanMgdl != nil || empty.TimeVeryLow != nil {
		t.Errorf("a day without readings has a mean of %v and %v very low, want neither", empty.GlucoseMeanMgdl, empty.TimeVeryLow)
	}
}

// A week's readings are its days' added up, with the mean and standard deviation of all of them, and its times in each
// range the means of its days'
func TestNewSummaryPeriodGlucoseStats(t *testing.T) {
	day := func(readings, meter, cgm int64, mean, sd float64, times ...float64) *sitedb.SummaryPeriod {
		return &sitedb.SummaryPeriod{Days: 1, GlucoseMgdlJson: "[]", GlucoseReadings: readings, MeterReadings: meter, CgmSlots: cgm,
			GlucoseMeanMgdl: &mean, GlucoseSdMgdl: &sd, TimeVeryLow: &times[0], TimeLow: &times[1], TimeInRange: &times[2], TimeHigh: &times[3], TimeVeryHigh: &times[4]}
	}
	// 100 twice and 200 twice have a mean of 150 and a standard deviation of 50
	week := NewSummaryPeriod("week", "2026-10-26", "UTC", 1, 2, []*sitedb.SummaryPeriod{
		day(2, 1, 2, 100, 0, 10, 20, 70, 0, 0), day(2, 0, 1, 200, 0, 0, 0, 50, 30, 20), {Days: 0, GlucoseMgdlJson: "[]", GlucoseReadings: 9}})
	got := []any{week.GlucoseReadings, week.MeterReadings, week.CgmSlots, *week.GlucoseMeanMgdl, *week.GlucoseSdMgdl,
		*week.TimeVeryLow, *week.TimeLow, *week.TimeInRange, *week.TimeHigh, *week.TimeVeryHigh}
	want := []any{int64(4), int64(1), int64(3), 150.0, 50.0, 5.0, 10.0, 60.0, 15.0, 10.0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("week's readings, meter readings, CGM slots, mean, SD and times are %v, want %v", got, want)
	}
}
