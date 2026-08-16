package v2

import (
	"context"
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// PropertiesList lists properties using the Nightscout v1 query format.
func PropertiesList(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		siteContext := httputils.GetSiteContext(r.Context())

		dr := LastDeviceStatusFromPWD(siteContext.SiteDB.Queries, r.Context())

		lastSensorStart, lastSensorChange, lastCannulaChange := GetLastSensorChangeCannulaChange(siteContext.SiteDB.Queries, r.Context())

		properties := Properties{
			UPBat: GetUptimeBattery(dr),
			IOB:   GetIOB(dr),
			COB:   GetCOB(dr),
			CAGE:  GetDeviceTime(lastCannulaChange),
			SAGE:  GetSensorAge(lastSensorStart, lastSensorChange),
		}

		nsutils.WriteJSON(w, http.StatusOK, properties)
	}
}

func LastDeviceStatusFromPWD(q *site.Queries, ctx context.Context) *site.DeviceReport {
	// We are going to ignore the error here, because not important
	ret, _ := q.GetLatestDeviceReportWithPWDInformation(ctx)
	return ret
}

func GetLastSensorChangeCannulaChange(q *site.Queries, ctx context.Context) (*site.ComponentEvent, *site.ComponentEvent, *site.ComponentEvent) {
	lastSensorChange, _ := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Sensor Change",
	})

	lastSensorStart, _ := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Sensor Start",
	})

	lastCannulaChange, _ := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Insulin Change",
	})

	return lastSensorStart, lastSensorChange, lastCannulaChange
}

/////
// MODELS
/////

func GetUptimeBattery(dr *site.DeviceReport) UptimeBattery {
	if dr == nil || dr.DeviceBatteryPercent == nil {
		return UptimeBattery{
			Found: ptr(false),
		}
	}

	return UptimeBattery{
		Min: &BatteryStatus{
			Battery: dr.DeviceBatteryPercent,
			Value:   dr.DeviceBatteryPercent,
			Display: ptr(fmt.Sprintf("%v%%", *dr.DeviceBatteryPercent)),
			Level:   dr.DeviceBatteryPercent,
		},
	}
}

func GetIOB(dr *site.DeviceReport) IOB {
	if dr == nil {
		return IOB{
			Found: ptr(false),
		}
	}

	return IOB{
		IOB:          dr.LoopIob,
		BasalIOB:     dr.LoopIobBasal,
		TreatmentIOB: dr.LoopIobTreatment,
	}
}

func GetCOB(dr *site.DeviceReport) COB {
	if dr == nil {
		return COB{
			Found: ptr(false),
		}
	}

	return COB{
		COB: dr.LoopCob,
	}
}

func GetDeviceTime(lastCannulaChange *site.ComponentEvent) DeviceTime {
	if lastCannulaChange == nil {
		return DeviceTime{
			Found: ptr(false),
		}
	}

	return DeviceTime{
		TreatmentDate: &lastCannulaChange.OccurredAt,
	}
}

func GetSensorAge(lastSensorStart, lastSensorChange *site.ComponentEvent) SensorAge {
	ssTime := GetDeviceTime(lastSensorStart)
	scTime := GetDeviceTime(lastSensorChange)
	var min *string
	if ssTime.TreatmentDate == nil && scTime.TreatmentDate != nil {
		min = ptr("Sensor Change")
	} else if ssTime.TreatmentDate != nil && scTime.TreatmentDate == nil {
		min = ptr("Sensor Start")
	} else if ssTime.TreatmentDate != nil && scTime.TreatmentDate != nil {
		if *ssTime.TreatmentDate < *scTime.TreatmentDate {
			min = ptr("Sensor Start")
		} else {
			min = ptr("Sensor Change")
		}
	}

	return SensorAge{
		SensorStart:  &ssTime,
		SensorChange: &scTime,
		Min:          min,
	}
}

func ptr[T any](v T) *T {
	return &v
}

// {
//   "upbat": {
//     "min": {
//       "battery": 87,
//       "value": 87,
//       "display": "87%",
//       "level": 75
//     }
//   },
//   "iob": {
//     "iob": 1.983,
//     "basaliob": -0.669,
//     "treatmentIob": 3.464
//   },
//   "cob": {
//     "cob": 12.39360495744192
//   },
//   "cage": {
//     "treatmentDate": 1722497240926,
//   },
//   "sage": {
//     "Sensor Start": {
//       "found": false
//     },
//     "Sensor Change": {
//       "treatmentDate": 1721978583000
//     },
//     "min": "Sensor Change"
//   }
// }

type Properties struct {
	UPBat UptimeBattery `json:"upbat,omitempty"`
	IOB   IOB           `json:"iob,omitempty"`
	COB   COB           `json:"cob,omitempty"`
	CAGE  DeviceTime    `json:"cage,omitempty"`
	SAGE  SensorAge     `json:"sage,omitempty"`
}

type UptimeBattery struct {
	Min   *BatteryStatus `json:"min,omitempty"`
	Found *bool          `json:"found,omitempty"`
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
	Found        *bool    `json:"found,omitempty"`
}

type COB struct {
	COB   *float64 `json:"cob,omitempty"`
	Found *bool    `json:"found,omitempty"`
}

type DeviceTime struct {
	TreatmentDate *int64 `json:"treatmentDate,omitempty"`
	Found         *bool  `json:"found,omitempty"`
}

type SensorAge struct {
	SensorStart  *DeviceTime `json:"Sensor Start,omitempty"`
	SensorChange *DeviceTime `json:"Sensor Change,omitempty"`
	Min          *string     `json:"min,omitempty"`
}
