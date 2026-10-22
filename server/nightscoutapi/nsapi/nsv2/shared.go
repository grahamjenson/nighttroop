package nsv2

import (
	"context"

	"nighttroop.t1d.nz/data/sql/sitedb"
)

func LastDeviceStatusFromPWD(q *sitedb.Queries, ctx context.Context) *sitedb.DeviceReport {
	// We are going to ignore the error here, because not important
	ret, err := q.GetLatestDeviceReportWithPWDInformation(ctx)
	if err != nil {
		ret = nil
	}
	return ret
}

func GetLastSensorChangeCannulaChange(q *sitedb.Queries, ctx context.Context) (*sitedb.ComponentEvent, *sitedb.ComponentEvent, *sitedb.ComponentEvent) {
	lastSensorChange, err := q.GetComponentEventByType(ctx, sitedb.GetComponentEventByTypeParams{
		EventType: "Sensor Change",
	})
	if err != nil {
		lastSensorChange = nil
	}
	lastSensorStart, err := q.GetComponentEventByType(ctx, sitedb.GetComponentEventByTypeParams{
		EventType: "Sensor Start",
	})
	if err != nil {
		lastSensorStart = nil
	}

	lastCannulaChange, err := q.GetComponentEventByType(ctx, sitedb.GetComponentEventByTypeParams{
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

func ptr[T any](v T) *T {
	return &v
}
