package sitedata

import (
	"context"
	"log/slog"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
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

type DeviceStatusesStore struct{}

// ////
// Shared
// ////
func (s *DeviceStatusesStore) GetName() string {
	return "device_statuses"
}

// ////
// Get Model
// /////
func (s *DeviceStatusesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteDeviceStatus, *apiutils.HTTPError) {
	return getModel(ctx, tx, (*sitedb.Queries).GetDeviceReport, sitedb.GetDeviceReportParams{ID: id}, "device status", sitemodels.DeviceStatusFromDB)
}

// GetHTMXResponse is the status's details, shown in a modal
func (s *DeviceStatusesStore) GetHTMXResponse(model *sitemodels.SiteDeviceStatus) templ.Component {
	return site_data_templ.DeviceStatusModal(model)
}

// ///
// Search Models
// ///

// SearchModels pages through one device's statuses when the query names a device,
// otherwise it returns the latest few statuses of each device that reported recently, grouped by device
func (s *DeviceStatusesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteDeviceStatus, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	if query.Device != nil {
		return deviceStatuses(ctx, q, sitequery.QueryToDeviceReportsForDevice(query))
	}

	devices, err := q.RecentDeviceNames(ctx, sitedb.RecentDeviceNamesParams{
		Since: time.Now().Add(-recentDeviceWindow).UnixMilli(),
		Limit: maxStatusDevices,
	})
	if err != nil {
		slog.Error("Search device names", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*sitemodels.SiteDeviceStatus{}
	for _, device := range devices {
		statuses, herr := deviceStatuses(ctx, q, sitedb.DeviceReportsForDeviceParams{Device: device, Limit: statusesPerDevice})
		if herr != nil {
			return nil, herr
		}
		models = append(models, statuses...)
	}
	return models, nil
}

func (s *DeviceStatusesStore) SearchHTMXResponse(models []*sitemodels.SiteDeviceStatus, query *apiutils.Query) templ.Component {
	if query.Device != nil {
		return site_data_templ.DeviceStatusHistory(*query.Device, models, query.Skip, query.Limit)
	}
	return site_data_templ.DeviceStatusGroups(models)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *DeviceStatusesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteDeviceStatus, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteDeviceReport(ctx, sitedb.SoftDeleteDeviceReportParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting device status", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *DeviceStatusesStore) DeleteHTMXResponse(model *sitemodels.SiteDeviceStatus) templ.Component {
	return site_data_templ.DeviceStatusDeleted()
}

///////
// UTIL METHODS
////////

func deviceStatuses(ctx context.Context, q *sitedb.Queries, params sitedb.DeviceReportsForDeviceParams) ([]*sitemodels.SiteDeviceStatus, *apiutils.HTTPError) {
	reports, err := q.DeviceReportsForDevice(ctx, params)
	return apiutils.Searched(reports, err, "device statuses", sitemodels.DeviceStatusFromDB)
}
