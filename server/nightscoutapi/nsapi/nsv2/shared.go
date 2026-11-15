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

// LastBatteryStatus is the newest device status with the uploader's battery in the last 30 minutes, from any
// device, nil when there is none
func LastBatteryStatus(q *sitedb.Queries, ctx context.Context) *sitedb.DeviceReport {
	ret, err := q.GetLatestDeviceReportWithBattery(ctx)
	if err != nil {
		return nil
	}
	return ret
}

// LastOfType is the newest treatment of a type not deleted, like a Sensor Change, nil when there is none
func LastOfType(q *sitedb.Queries, ctx context.Context, groupType string) *sitedb.EventGroup {
	groups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: groupType, Limit: 1})
	if err != nil || len(groups) == 0 {
		return nil
	}
	return groups[0]
}
