package sitemodels

import (
	"reflect"
	"testing"
)

func TestNewSummaryDay(t *testing.T) {
	const hour = int64(60 * 60_000)
	// The day is from 1 hour until 3 hours, with the range from 4 to 10
	start, end := hour, 3*hour
	reading := func(at int64, value float64) *SiteCleanGlucoseReading {
		r := &SiteCleanGlucoseReading{}
		r.OccurredAt, r.Reading = at, value
		return r
	}
	dose := func(at int64, units float64) *SiteInsulinDose { return &SiteInsulinDose{OccurredAt: at, Dose: units} }
	// basal is a minute from minutesSince each, with the rate in force for minutes of them
	basal := func(rate float64, minutes int) []SummaryMinute {
		list := make([]SummaryMinute, minutes)
		for i := range list {
			list[i].Basal = &rate
		}
		return list
	}

	tests := []struct {
		name         string
		readings     []*SiteCleanGlucoseReading
		doses        []*SiteInsulinDose
		minutesSince int64
		minutes      []SummaryMinute
		inRange      *float64
		dose         float64
	}{
		{name: "nothing in the day", inRange: nil, dose: 0},
		{name: "readings on the range's edges are in it", readings: []*SiteCleanGlucoseReading{reading(start, 4), reading(start+1, 10)}, inRange: new(100.0)},
		{name: "readings below and above are out of it", readings: []*SiteCleanGlucoseReading{reading(start, 3.9), reading(start+1, 10.1), reading(start+2, 7)}, inRange: new(33.3)},
		{name: "readings outside the day are left out", readings: []*SiteCleanGlucoseReading{reading(start-1, 2), reading(end, 2), reading(start, 5)}, inRange: new(100.0)},
		{name: "doses in the day add up", doses: []*SiteInsulinDose{dose(start, 1.5), dose(end-1, 2), dose(end, 9), dose(start-1, 9)}, dose: 3.5},
		{name: "the basal in force each minute of the day adds up", minutesSince: start, minutes: basal(1.2, 120), dose: 2.4},
		{name: "basal outside the day is left out", minutesSince: 0, minutes: basal(0.6, 4*60), dose: 1.2},
		{name: "minutes without a basal add nothing", minutesSince: start, minutes: make([]SummaryMinute, 60), dose: 0},
		{name: "doses and basal add up", doses: []*SiteInsulinDose{dose(start, 4)}, minutesSince: start, minutes: basal(1, 60), dose: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day := NewSummaryDay(start, end, 4, 10, tt.readings, tt.doses, tt.minutesSince, tt.minutes)
			if day.Start != start || day.End != end {
				t.Errorf("day is %d to %d, want %d to %d", day.Start, day.End, start, end)
			}
			if !reflect.DeepEqual(day.TimeInRange, tt.inRange) {
				t.Errorf("time in range is %v, want %v", ptrValue(day.TimeInRange), ptrValue(tt.inRange))
			}
			if day.TotalDailyDose != tt.dose {
				t.Errorf("total daily dose is %v, want %v", day.TotalDailyDose, tt.dose)
			}
		})
	}
}

func ptrValue(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}
