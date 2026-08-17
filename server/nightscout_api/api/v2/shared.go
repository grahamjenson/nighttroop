package v2

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/site"
)

func LastDeviceStatusFromPWD(q *site.Queries, ctx context.Context) *site.DeviceReport {
	// We are going to ignore the error here, because not important
	ret, err := q.GetLatestDeviceReportWithPWDInformation(ctx)
	if err != nil {
		ret = nil
	}
	return ret
}

func GetLastSensorChangeCannulaChange(q *site.Queries, ctx context.Context) (*site.ComponentEvent, *site.ComponentEvent, *site.ComponentEvent) {
	lastSensorChange, err := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Sensor Change",
	})
	if err != nil {
		lastSensorChange = nil
	}
	lastSensorStart, err := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Sensor Start",
	})
	if err != nil {
		lastSensorStart = nil
	}

	lastCannulaChange, err := q.GetComponentEventByType(ctx, site.GetComponentEventByTypeParams{
		EventType: "Insulin Change",
	})
	if err != nil {
		lastCannulaChange = nil
	}
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

func ptr[T any](v T) *T {
	return &v
}
