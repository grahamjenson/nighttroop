package collections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type DeviceStatusCollection struct {
	converter func(c *site.EventGroupsBundle) nsutils.Record
}

func NewDeviceStatusCollection() DeviceStatusCollection {
	return DeviceStatusCollection{
		converter: nsutils.DBModelToRecord(nsutils.Devicestatus, nightscout.DBModelToDeviceStatus),
	}
}

func (c DeviceStatusCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	deviceStatusModel := nightscout.DeviceStatus{}
	err := utils.FromMap(data, &deviceStatusModel)
	if err != nil {
		slog.Error("failed to map data to device status model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	cc := nightscout.CreateContext{
		IDPrefix:              "ns",
		CreatedByIdentityID:   siteContext.IdentityModel.ID,
		CreatedByIdentityName: siteContext.IdentityModel.Name,
	}

	createDevice := nightscout.DeviceStatusToCreateDevice(deviceStatusModel, cc)
	createLoop := nightscout.DeviceStatusToCreateLoop(deviceStatusModel, cc)
	createPump := nightscout.DeviceStatusToCreatePump(deviceStatusModel, cc)
	createEventGroup := nightscout.DeviceStatusToCreateEventGroup(deviceStatusModel, cc)

	dbModel, rerr := EventGroupCreate(ctx, siteContext, site.CreateEventGroupBundle{
		DeviceState: &createDevice,
		LoopState:   &createLoop,
		PumpState:   &createPump,
		EventGroup:  createEventGroup,
	})

	if rerr != nil {
		slog.Error("failed to create device status in database", "error", rerr)
		return nil, rerr
	}
	if dbModel == nil {
		slog.Error("failed to create device status in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), nil
}
func (c DeviceStatusCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	deviceStatusModel := nightscout.DeviceStatus{}
	err := utils.FromMap(patch, &deviceStatusModel)

	if err != nil {
		slog.Error("failed to map data to device status model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	cc := nightscout.UpdateContext{
		UpdatedByIdentityID:   siteContext.IdentityModel.ID,
		UpdatedByIdentityName: siteContext.IdentityModel.Name,
	}

	createDevice := nightscout.DeviceStatusToUpdateDevice(deviceStatusModel, cc)
	createLoop := nightscout.DeviceStatusToUpdateLoop(deviceStatusModel, cc)
	createPump := nightscout.DeviceStatusToUpdatePump(deviceStatusModel, cc)
	createEventGroup := nightscout.DeviceStatusToUpdateEventGroup(deviceStatusModel, cc)

	dbModel, rerr := EventGroupUpdate(ctx, siteContext, site.UpdateEventGroupBundle{
		DeviceState: &createDevice,
		LoopState:   &createLoop,
		PumpState:   &createPump,
		EventGroup:  createEventGroup,
	})

	if rerr != nil {
		slog.Error("failed to update device status in database", "error", rerr)
		return nil, rerr
	}
	if dbModel == nil {
		slog.Error("failed to update device status in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), nil
}

// ///////
// Default EventGroup stuff
// ///////
func (c DeviceStatusCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	seg, rerr := EventGroupGet(ctx, siteContext, identifier)
	if rerr != nil {
		return nil, rerr
	}
	return ptr(c.converter(seg)), nil
}

func (c DeviceStatusCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	deviceStatuses, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(deviceStatuses, c.converter), nil
}

func (c DeviceStatusCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return EventGroupDelete(ctx, siteContext, identifier, permanent)
}

func (c DeviceStatusCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return EventGroupDeleteType(ctx, siteContext, typeString, permanent)
}
