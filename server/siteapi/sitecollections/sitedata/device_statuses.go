package sitedata

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	// statusesPerDevice is how many of each device's latest statuses the grouped view shows
	statusesPerDevice = 3
	// maxStatusDevices caps the grouped view, so a flood of device names cannot make it expensive
	maxStatusDevices = 20
	// recentDeviceWindow is how far back a device must have reported to be in the grouped view
	recentDeviceWindow = 7 * 24 * time.Hour
)

type DeviceStatusesCollection struct{}

// ////
// Shared
// ////
func (s *DeviceStatusesCollection) GetName() string {
	return "device_statuses"
}

// ////
// Get Model
// /////
func (s *DeviceStatusesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.DeviceReport, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: id})
}

// ToModel is the site model of a row
func (s *DeviceStatusesCollection) ToModel(row *sitedb.DeviceReport, siteContext *httputils.SiteContext) *sitemodels.SiteDeviceStatus {
	return sitemodels.DeviceStatusFromDB(row)
}

// GetHTMXResponse is the status's details, shown in a modal
func (s *DeviceStatusesCollection) GetHTMXResponse(model *sitemodels.SiteDeviceStatus) templ.Component {
	return site_data_templ.DeviceStatusModal(model)
}

// ///
// Search Models
// ///

// SearchRows pages through one device's statuses when the query names a device,
// otherwise it returns the latest few statuses of each device that reported recently, grouped by device
func (s *DeviceStatusesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.DeviceReport, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	if query.Device != nil {
		return q.DeviceReportsForDevice(ctx, siteapiutils.QueryToDeviceReportsForDevice(query))
	}

	devices, err := q.RecentDeviceNames(ctx, sitedb.RecentDeviceNamesParams{
		Since: time.Now().Add(-recentDeviceWindow).UnixMilli(),
		Limit: maxStatusDevices,
	})
	if err != nil {
		return nil, err
	}
	rows := []*sitedb.DeviceReport{}
	for _, device := range devices {
		statuses, err := q.DeviceReportsForDevice(ctx, sitedb.DeviceReportsForDeviceParams{Device: device, Limit: statusesPerDevice})
		if err != nil {
			return nil, err
		}
		rows = append(rows, statuses...)
	}
	return rows, nil
}

func (s *DeviceStatusesCollection) SearchHTMXResponse(models []*sitemodels.SiteDeviceStatus, query *apiutils.Query) templ.Component {
	if query.Device != nil {
		return site_data_templ.DeviceStatusHistory(*query.Device, models, query.Skip, query.Limit)
	}
	return site_data_templ.DeviceStatusGroups(models)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *DeviceStatusesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.SoftDeleteDeviceReport(ctx, sitedb.SoftDeleteDeviceReportParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *DeviceStatusesCollection) DeleteHTMXResponse(model *sitemodels.SiteDeviceStatus) templ.Component {
	return site_data_templ.DeviceStatusDeleted()
}

///////
// UTIL METHODS
////////
