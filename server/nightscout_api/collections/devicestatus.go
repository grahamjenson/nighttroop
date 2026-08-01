package collections

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type DeviceStatusCollection struct {
	converter func(c *site.SearchEventGroupsRow) nsutils.Record
}

func NewDeviceStatusCollection() DeviceStatusCollection {
	return DeviceStatusCollection{
		converter: nsutils.DBModelToRecord(nsutils.Devicestatus, nightscout.DBModelToDeviceStatus),
	}
}

func (c DeviceStatusCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	deviceStatusModel := &nightscout.DeviceStatus{}
	err := utils.FromMap(data, deviceStatusModel)

	if err != nil {
		slog.Error("failed to map data to device status model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	dbModel, err := createDeviceStatus(ctx, siteContext, *deviceStatusModel)

	if err != nil {
		slog.Error("failed to create device status in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create device status in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), nil
}

func (c DeviceStatusCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.GetEventGroup(ctx, site.GetEventGroupParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows || dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get device status", "error", err)
		return nil, nsutils.ErrInternal
	}

	seg := site.GetEventGroupToSearchEventGroup(*dbModel)
	record := c.converter(&seg)
	return &record, nil
}

func (c DeviceStatusCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c DeviceStatusCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c DeviceStatusCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c DeviceStatusCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func createDeviceStatus(ctx context.Context, siteContext *httputils.SiteContext, deviceStatusModel nightscout.DeviceStatus) (*site.SearchEventGroupsRow, error) {
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	cc := nightscout.CreateContext{
		IDPrefix:              "ns",
		CreatedByIdentityID:   siteContext.IdentityModel.ID,
		CreatedByIdentityName: siteContext.IdentityModel.Name,
	}

	defer tx.Rollback()

	createDevice := nightscout.DeviceStatusToCreateDevice(deviceStatusModel, cc)
	createLoop := nightscout.DeviceStatusToCreateLoop(deviceStatusModel, cc)
	createPump := nightscout.DeviceStatusToCreatePump(deviceStatusModel, cc)
	createEventGroup := nightscout.DeviceStatusToCreateEventGroup(deviceStatusModel, cc)

	createEventGroup.DeviceStateID = &createDevice.ID
	createEventGroup.LoopStateID = &createLoop.ID
	createEventGroup.PumpStateID = &createPump.ID

	q := siteContext.SiteDB.Queries.WithTx(tx)

	ds, err := q.CreateDeviceState(ctx, createDevice)
	if err != nil {
		return nil, err
	}

	ls, err := q.CreateLoopState(ctx, createLoop)
	if err != nil {
		return nil, err
	}

	ps, err := q.CreatePumpState(ctx, createPump)
	if err != nil {
		return nil, err
	}

	dbEG, err := q.CreateEventGroup(ctx, createEventGroup)
	if err != nil {
		return nil, err
	}

	if dbEG == nil || ds == nil || ls == nil || ps == nil {
		return nil, fmt.Errorf("create DeviceStatus has failed")
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	sEG := site.EventGroupToSearchEventGroup(*dbEG)
	sEG.LoopState = *ls
	sEG.PumpState = *ps
	sEG.DeviceState = *ds
	return &sEG, nil
}
