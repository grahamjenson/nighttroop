package nsv2

import (
	"net/http"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// PropertiesList lists properties using the Nightscout v1 query format.
func Summary(store nsutils.Store) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		siteContext := httputils.GetSiteContext(r.Context())

		dr := LastDeviceStatusFromPWD(siteContext.SiteDB.Queries, r.Context())

		_, lastSensorChange, lastCannulaChange := GetLastSensorChangeCannulaChange(siteContext.SiteDB.Queries, r.Context())

		summary := SummaryStr{
			IOB:           GetIOBFloat(dr),
			COB:           GetCOBFloat(dr),
			CannulaChange: GetComponentEventOccuredAt(lastCannulaChange),
			SensorChange:  GetComponentEventOccuredAt(lastSensorChange),
		}

		apiutils.WriteJSON(w, http.StatusOK, summary)
	}
}

// {
//   "current": { "sgv": 116, "direction": "FortyFiveDown", "trendRate": -0.5, "delta": -3, "mills": 1719300000000, "noise": 1 },
//   "history": [
//     { "sgv": 98, "direction": "Flat", "mills": 1719290700000 },
//     { "sgv": 101, "direction": "FortyFiveUp", "mills": 1719291000000 },
//     { "sgv": 108, "direction": "SingleUp", "mills": 1719291300000 },
//     { "sgv": 132, "direction": "SingleUp", "mills": 1719291600000 },
//     { "sgv": 151, "direction": "FortyFiveUp", "mills": 1719291900000 },
//     { "sgv": 162, "direction": "Flat", "mills": 1719292200000 },
//     { "sgv": 160, "direction": "Flat", "mills": 1719292500000 },
//     { "sgv": 150, "direction": "FortyFiveDown", "mills": 1719292800000 },
//     { "sgv": 138, "direction": "FortyFiveDown", "mills": 1719293100000 },
//     { "sgv": 128, "direction": "FortyFiveDown", "mills": 1719293400000 },
//     { "sgv": 122, "direction": "Flat", "mills": 1719293700000 },
//     { "sgv": 120, "direction": "Flat", "mills": 1719294000000 },
//     { "sgv": 121, "direction": "Flat", "mills": 1719294300000 },
//     { "sgv": 119, "direction": "FortyFiveDown", "mills": 1719294600000 },
//     { "sgv": 116, "direction": "FortyFiveDown", "mills": 1719300000000 }
//   ],
//   "iob": 1.23,
//   "cob": 14.0,
//   "predictions": {
//     "values": [113, 110, 107, 104, 101, 99, 97, 96, 95, 94, 94, 93],
//     "startMills": 1719300300000,
//     "intervalMills": 300000,
//     "source": "loop"
//   },
//   "alarm": {
//     "level": 0,
//     "type": "none",
//     "message": "",
//     "triggeredMills": 0,
//     "isSilenced": false,
//     "silenceExpiresMills": null
//   },
//   "serverMills": 1719300001000
// }

type SummaryStr struct {
	Current       *SummaryGlucose   `json:"current"`
	History       []*SummaryGlucose `json:"history"`
	IOB           *float64          `json:"iob"`
	COB           *float64          `json:"cob"`
	SensorChange  *int64            `json:"sensor_change"`
	CannulaChange *int64            `json:"cannula_change"`
	Predictions   *Predictions      `json:"predictions"`
	Alarm         *Alarm            `json:"alarm"`
	ServerMills   *int64            `json:"serverMills"`
}

type SummaryGlucose struct {
	SGV       float64 `json:"sgv"`
	Direction string  `json:"direction"`
	TrendRate float64 `json:"trendRate,omitempty"`
	Delta     float64 `json:"delta,omitempty"`
	Mills     int64   `json:"mills"`
	Noise     int64   `json:"noise,omitempty"`
}

type Predictions struct {
	Values        []float64 `json:"values"`
	StartMills    int64     `json:"startMills"`
	IntervalMills int64     `json:"intervalMills"`
	Source        string    `json:"source"`
}

type Alarm struct {
	Level               int    `json:"level"`
	Type                string `json:"type"`
	Message             string `json:"message"`
	TriggeredMills      int64  `json:"triggeredMills"`
	IsSilenced          bool   `json:"isSilenced"`
	SilenceExpiresMills *int64 `json:"silenceExpiresMills"`
}

func GetIOBFloat(dr *sitedb.DeviceReport) *float64 {
	if dr == nil || dr.LoopIob == nil {
		return nil
	}
	return dr.LoopIob
}

func GetCOBFloat(dr *sitedb.DeviceReport) *float64 {
	if dr == nil || dr.LoopCob == nil {
		return nil
	}
	return dr.LoopCob
}

func GetComponentEventOccuredAt(lastSensorChange *sitedb.ComponentEvent) *int64 {
	if lastSensorChange == nil || lastSensorChange.OccurredAt == 0 {
		return nil
	}
	return &lastSensorChange.OccurredAt
}
