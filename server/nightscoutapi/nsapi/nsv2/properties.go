package nsv2

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/display"
	"nighttroop.t1d.nz/utils/httputils"
)

// PropertiesList lists the properties, or with {names} only the comma separated ones asked for, so an unknown name
// gives {} as Nightscout does
func PropertiesList(store nsutils.Store) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		properties := Properties(r.Context(), httputils.GetSiteContext(r.Context()).SiteDB.Queries(), time.Now())
		if names := chi.URLParam(r, "names"); names != "" {
			asked := map[string]any{}
			for _, name := range strings.Split(names, ",") {
				if property, ok := properties[name]; ok {
					asked[name] = property
				}
			}
			properties = asked
		}
		apiutils.WriteJSON(w, http.StatusOK, properties)
	}
}

// Properties are the Nightscout plugin properties Nighttroop has the data for, in the plugins' shapes
func Properties(ctx context.Context, q *sitedb.Queries, now time.Time) map[string]any {
	dr := LastDeviceStatusFromPWD(q, ctx)
	last := func(eventType string) *sitedb.ComponentEvent {
		event, err := q.GetComponentEventByType(ctx, sitedb.GetComponentEventByTypeParams{EventType: eventType})
		if err != nil {
			return nil
		}
		return event
	}
	properties := map[string]any{
		"upbat": GetUptimeBattery(dr),
		"iob":   GetIOB(dr),
		"cob":   GetCOB(dr),
		"cage":  GetAge(last("Site Change"), now),
		"iage":  GetAge(last("Insulin Change"), now),
		"bage":  GetAge(last("Pump Battery Change"), now),
		"sage":  GetSensorAge(last("Sensor Start"), last("Sensor Change"), now),
		"bgnow": struct{}{},
	}

	// Like Nightscout, bgnow is empty, and delta and direction left out, without a reading in the last 15 minutes
	readings, err := q.AlarmGlucoseReadings(ctx, sitedb.AlarmGlucoseReadingsParams{Minutes: 15, Limit: 2})
	if err != nil || len(readings) == 0 || readings[0].OccurredAt < now.Add(-15*time.Minute).UnixMilli() {
		return properties
	}
	properties["bgnow"] = GetBucket(readings[0])
	if readings[0].Direction != nil {
		label := display.Direction(readings[0].Direction)
		properties["direction"] = Direction{Value: *readings[0].Direction, Label: label, Entity: fmt.Sprintf("&#%d;", []rune(label)[0])}
	}
	if len(readings) > 1 {
		properties["delta"] = GetDelta(readings[0], readings[1])
	}
	return properties
}

type UptimeBattery struct {
	Min     *BatteryStatus `json:"min,omitempty"`
	Display *string        `json:"display,omitempty"`
	Level   *float64       `json:"level,omitempty"`
	Found   *bool          `json:"found,omitempty"`
}

type BatteryStatus struct {
	Battery *float64 `json:"battery,omitempty"`
	Value   *float64 `json:"value,omitempty"`
	Display *string  `json:"display,omitempty"`
	Level   *float64 `json:"level,omitempty"`
}

type IOB struct {
	IOB          *float64 `json:"iob,omitempty"`
	BasalIOB     *float64 `json:"basaliob,omitempty"`
	TreatmentIOB *float64 `json:"treatmentIob,omitempty"`
	Display      *string  `json:"display,omitempty"`
	DisplayLine  *string  `json:"displayLine,omitempty"`
	Found        *bool    `json:"found,omitempty"`
	Source
}

type COB struct {
	COB         *float64 `json:"cob,omitempty"`
	Display     *float64 `json:"display,omitempty"`
	DisplayLine *string  `json:"displayLine,omitempty"`
	Found       *bool    `json:"found,omitempty"`
	Source
}

// Source is the device status IOB and COB came from
type Source struct {
	Mills  *int64  `json:"mills,omitempty"`
	Device *string `json:"device,omitempty"`
	Source *string `json:"source,omitempty"`
}

// Age is how long ago a component, such as the cannula, was changed, in hours
type Age struct {
	Found         bool   `json:"found"`
	TreatmentDate int64  `json:"treatmentDate"`
	Age           int64  `json:"age"`
	Days          int64  `json:"days"`
	Hours         int64  `json:"hours"`
	Display       string `json:"display"`
	DisplayLong   string `json:"displayLong,omitempty"`
}

type NotFound struct {
	Found bool `json:"found"`
}

type SensorAge struct {
	SensorStart  any    `json:"Sensor Start"`
	SensorChange any    `json:"Sensor Change"`
	Min          string `json:"min"`
}

// Bucket is Nightscout's readings at a time: bgnow, and the previous one in delta
type Bucket struct {
	Last  float64 `json:"last"`
	Mean  float64 `json:"mean"`
	Mills int64   `json:"mills"`
	SGVs  []SGV   `json:"sgvs"`
}

type SGV struct {
	ID        string   `json:"_id"`
	Mgdl      float64  `json:"mgdl"`
	Scaled    float64  `json:"scaled"`
	Mills     int64    `json:"mills"`
	Type      string   `json:"type"`
	Device    *string  `json:"device,omitempty"`
	Direction *string  `json:"direction,omitempty"`
	Noise     *float64 `json:"noise,omitempty"`
}

type Delta struct {
	Absolute     float64          `json:"absolute"`
	Mgdl         float64          `json:"mgdl"`
	Scaled       float64          `json:"scaled"`
	Display      string           `json:"display"`
	ElapsedMins  float64          `json:"elapsedMins"`
	Interpolated bool             `json:"interpolated"`
	Mean5MinsAgo float64          `json:"mean5MinsAgo"`
	Times        map[string]int64 `json:"times"`
	Previous     Bucket           `json:"previous"`
}

type Direction struct {
	Display *string `json:"display"`
	Value   string  `json:"value"`
	Label   string  `json:"label"`
	Entity  string  `json:"entity"`
}

func GetSource(dr *sitedb.DeviceReport) Source {
	source := Source{Mills: &dr.OccurredAt, Device: dr.CreatedByDeviceName}
	if dr.LoopName != nil {
		source.Source = new(map[string]string{"loop": "Loop", "openaps": "OpenAPS"}[*dr.LoopName])
	}
	return source
}

func GetIOB(dr *sitedb.DeviceReport) IOB {
	if dr == nil || dr.LoopIob == nil {
		return IOB{Found: new(false)}
	}
	return IOB{
		IOB:          dr.LoopIob,
		BasalIOB:     dr.LoopIobBasal,
		TreatmentIOB: dr.LoopIobTreatment,
		Display:      new(fmt.Sprintf("%.2f", *dr.LoopIob)),
		DisplayLine:  new(fmt.Sprintf("IOB: %.2fU", *dr.LoopIob)),
		Source:       GetSource(dr),
	}
}

func GetCOB(dr *sitedb.DeviceReport) COB {
	if dr == nil || dr.LoopCob == nil {
		return COB{Found: new(false)}
	}
	cob := math.Round(*dr.LoopCob*10) / 10
	return COB{COB: dr.LoopCob, Display: &cob, DisplayLine: new(fmt.Sprintf("COB: %vg", cob)), Source: GetSource(dr)}
}

// GetAge is shown in hours up to a day, then in days and hours, as Nightscout's age plugins do
func GetAge(event *sitedb.ComponentEvent, now time.Time) any {
	if event == nil {
		return NotFound{}
	}
	hours := (now.UnixMilli() - event.OccurredAt) / time.Hour.Milliseconds()
	age := Age{Found: true, TreatmentDate: event.OccurredAt, Age: hours, Days: hours / 24, Hours: hours % 24, Display: fmt.Sprintf("%dh", hours)}
	if hours > 24 {
		age.Display = fmt.Sprintf("%dd%dh", age.Days, age.Hours)
	}
	return age
}

// GetSensorAge has the ages of the last sensor start and change, and min names the latest, as the sensor's age
func GetSensorAge(lastSensorStart, lastSensorChange *sitedb.ComponentEvent, now time.Time) SensorAge {
	sensorAge := SensorAge{SensorStart: GetAge(lastSensorStart, now), SensorChange: GetAge(lastSensorChange, now), Min: "Sensor Start"}
	if lastSensorChange != nil && (lastSensorStart == nil || lastSensorChange.OccurredAt > lastSensorStart.OccurredAt) {
		sensorAge.Min = "Sensor Change"
	}
	for _, age := range []*any{&sensorAge.SensorStart, &sensorAge.SensorChange} {
		if found, ok := (*age).(Age); ok {
			found.DisplayLong = fmt.Sprintf("%d days %d hours", found.Days, found.Hours)
			*age = found
		}
	}
	return sensorAge
}

// GetUptimeBattery levels the battery in quarters, as Nightscout does
func GetUptimeBattery(dr *sitedb.DeviceReport) UptimeBattery {
	if dr == nil || dr.DeviceBatteryPercent == nil {
		return UptimeBattery{Found: new(false)}
	}
	battery := *dr.DeviceBatteryPercent
	level := 25.0
	switch {
	case battery >= 95:
		level = 100
	case battery >= 55:
		level = 75
	case battery >= 30:
		level = 50
	}
	shown := fmt.Sprintf("%v%%", battery)
	return UptimeBattery{
		Min:     &BatteryStatus{Battery: &battery, Value: &battery, Display: &shown, Level: &level},
		Display: &shown,
		Level:   &level,
	}
}

func GetBucket(reading *sitedb.GlucoseReading) Bucket {
	return Bucket{Last: reading.Mgdl, Mean: reading.Mgdl, Mills: reading.OccurredAt, SGVs: []SGV{{
		ID: reading.ID, Mgdl: reading.Mgdl, Scaled: reading.Mgdl, Mills: reading.OccurredAt, Type: reading.ReadingType,
		Device: reading.CreatedByDeviceName, Direction: reading.Direction, Noise: reading.Noise,
	}}}
}

// GetDelta is the change from the previous reading, scaled to 5 minutes when they are more than 9 minutes apart
func GetDelta(recent, previous *sitedb.GlucoseReading) Delta {
	absolute := recent.Mgdl - previous.Mgdl
	elapsed := float64(recent.OccurredAt-previous.OccurredAt) / float64(time.Minute.Milliseconds())
	mgdl := absolute
	if elapsed > 9 {
		mgdl = math.Round(absolute * 5 / elapsed)
	}
	return Delta{
		Absolute: absolute, Mgdl: mgdl, Scaled: mgdl, Display: fmt.Sprintf("%+.0f", mgdl), ElapsedMins: elapsed,
		Interpolated: elapsed > 9, Mean5MinsAgo: recent.Mgdl - mgdl, Previous: GetBucket(previous),
		Times: map[string]int64{"recent": recent.OccurredAt, "previous": previous.OccurredAt},
	}
}
