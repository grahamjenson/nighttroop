package sitemodels

import (
	"encoding/json"
	"reflect"
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
			day := NewSummaryDaily("1970-01-01", "UTC", start, end, 70, 180, tt.readings, tt.doses, tt.food, tt.minutesSince, tt.minutes)
			if day.Date != "1970-01-01" || day.Timezone != "UTC" || day.StartAt != start || day.EndAt != end {
				t.Errorf("day is %s in %s from %d to %d", day.Date, day.Timezone, day.StartAt, day.EndAt)
			}
			if !reflect.DeepEqual(day.TimeInRange, tt.inRange) {
				t.Errorf("time in range is %v, want %v", day.TimeInRange, tt.inRange)
			}
			if day.TotalDailyDose != tt.dose {
				t.Errorf("total daily dose is %v, want %v", day.TotalDailyDose, tt.dose)
			}
			if day.TotalCarbs != tt.carbs {
				t.Errorf("total carbs is %v, want %v", day.TotalCarbs, tt.carbs)
			}
			want, _ := json.Marshal(tt.glucose)
			if day.GlucoseMgdlJson != string(want) {
				t.Errorf("glucose is %s, want %s", day.GlucoseMgdlJson, want)
			}
		})
	}
}
