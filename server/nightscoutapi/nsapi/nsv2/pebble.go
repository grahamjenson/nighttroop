package nsv2

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	// maxPebbleReadings is the most readings a /pebble count asks for, as many as a search returns
	maxPebbleReadings = 1000
	// pebbleMinutes is how far back /pebble's readings go, Nightscout's two days
	pebbleMinutes = 48 * 60
)

// pebbleTrends are Nightscout's trend numbers for each direction, which watch faces show arrows for
var pebbleTrends = map[string]int{"NONE": 0, "DoubleUp": 1, "SingleUp": 2, "FortyFiveUp": 3, "Flat": 4, "FortyFiveDown": 5,
	"SingleDown": 6, "DoubleDown": 7, "NOT COMPUTABLE": 8, "RATE OUT OF RANGE": 9}

// pebbleBG is a reading as Nightscout's /pebble has it: the glucose as a string in the units asked for, and on the
// newest the delta, the uploader's battery and IOB
type pebbleBG struct {
	SGV       string   `json:"sgv"`
	Trend     int      `json:"trend"`
	Direction *string  `json:"direction,omitempty"`
	Datetime  int64    `json:"datetime"`
	BGDelta   *float64 `json:"bgdelta,omitempty"`
	Battery   *string  `json:"battery,omitempty"`
	IOB       *string  `json:"iob,omitempty"`
}

// Pebble is Nightscout's /pebble, which Pebble and other watch faces read: the newest count readings (1 by default),
// in the units asked for (mmol or mg/dl) or the settings', with no calibrations
func Pebble() func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, _ *nsutils.Identity) {
		siteContext := httputils.GetSiteContext(r.Context())
		count, err := strconv.Atoi(r.URL.Query().Get("count"))
		if err != nil || count < 1 {
			count = 1
		}
		count = min(count, maxPebbleReadings)
		units := *siteContext.GetOwnSettings().GetUnits()
		if asked := r.URL.Query().Get("units"); asked != "" {
			units = consts.UnitsMGDL
			if strings.HasPrefix(strings.ToLower(asked), "mmol") {
				units = consts.UnitsMMOLL
			}
		}
		q := siteContext.SiteDB.Queries()
		// One more than asked for, the delta's previous reading
		readings, _ := q.AlarmGlucoseReadings(r.Context(), sitedb.AlarmGlucoseReadingsParams{Minutes: pebbleMinutes, Limit: int64(count + 1)})
		cutoff := time.Now().Add(-pebbleMinutes * time.Minute).UnixMilli()
		bgs := []pebbleBG{}
		for i, reading := range readings {
			if i == count || reading.OccurredAt < cutoff {
				break
			}
			bgs = append(bgs, pebbleBG{SGV: pebbleGlucose(units, reading.Mgdl), Trend: pebbleTrend(reading.Direction),
				Direction: reading.Direction, Datetime: reading.OccurredAt})
		}
		if len(bgs) > 0 {
			bgs[0].BGDelta = new(0.0)
			if len(readings) > 1 {
				bgs[0].BGDelta = new(units.FromMGDL(GetDelta(readings[0], readings[1]).Mgdl))
				if units == consts.UnitsMMOLL {
					*bgs[0].BGDelta = consts.Round(*bgs[0].BGDelta, 1)
				}
			}
			dr := LastDeviceStatusFromPWD(q, r.Context())
			if battery := GetUptimeBattery(dr); battery.Min != nil {
				bgs[0].Battery = new(fmt.Sprint(*battery.Min.Value))
			}
			bgs[0].IOB = GetIOB(dr).Display
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": []map[string]int64{{"now": time.Now().UnixMilli()}}, "bgs": bgs, "cals": []any{}})
	}
}

// pebbleGlucose is a glucose as /pebble writes it: whole mg/dL, or mmol/L to one decimal
func pebbleGlucose(units consts.Units, mgdl float64) string {
	if units == consts.UnitsMMOLL {
		return strconv.FormatFloat(consts.Round(units.FromMGDL(mgdl), 1), 'f', 1, 64)
	}
	return strconv.FormatFloat(mgdl, 'f', 0, 64)
}

// pebbleTrend is a direction's trend number, 0 (NONE) when it has none
func pebbleTrend(direction *string) int {
	if direction == nil {
		return 0
	}
	return pebbleTrends[*direction]
}
