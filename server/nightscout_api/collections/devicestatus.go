package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type DeviceStatusCollection struct {
	converter func(c *site.DeviceReport) nsutils.Record
}

func NewDeviceStatusCollection() DeviceStatusCollection {
	return DeviceStatusCollection{
		converter: nsutils.DBModelToRecord(nsutils.Devicestatus, nightscout.DBModelToDeviceStatus),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///
func (c DeviceStatusCollection) Name() string {
	return "devicestatus"
}

func (c DeviceStatusCollection) Convert(dr *site.DeviceReport) nsutils.Record {
	return c.converter(dr)
}

func (c DeviceStatusCollection) CreateDBFunc(siteContext *httputils.SiteContext, ctx context.Context, nsModel nightscout.DeviceStatus, cc nightscout.CreateContext) (*site.DeviceReport, error) {
	return siteContext.SiteDB.Queries.CreateDeviceReport(ctx, nightscout.DeviceStatusToCreateDeviceReport(nsModel, cc))
}

func (c DeviceStatusCollection) UpdateSBFunc(siteContext *httputils.SiteContext, ctx context.Context, updateParams site.UpdateDeviceReportParams) (*site.DeviceReport, error) {
	return siteContext.SiteDB.Queries.UpdateDeviceReport(ctx, updateParams)
}

func (c DeviceStatusCollection) GetDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (*site.DeviceReport, error) {
	return siteContext.SiteDB.Queries.GetDeviceReport(ctx, site.GetDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) HardDeleteDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (sql.Result, error) {
	return siteContext.SiteDB.Queries.HardDeleteDeviceReport(ctx, site.HardDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) SoftDeleteDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (sql.Result, error) {
	return siteContext.SiteDB.Queries.SoftDeleteDeviceReport(ctx, site.SoftDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) PatchDBFunc(siteContext *httputils.SiteContext, ctx context.Context, nsModel nightscout.DeviceStatus, identifier string, uc nightscout.UpdateContext) (*site.DeviceReport, error) {
	updateParams := nightscout.DeviceStatusToUpdateDeviceReport(nsModel, uc)
	updateParams.ID = identifier

	dbUpdateModel, err := siteContext.SiteDB.Queries.UpdateDeviceReport(ctx, updateParams)
	return dbUpdateModel, err
}

/////
// END OF FUNCTIONS USED BY META COLLECTIONS
/////

func (c DeviceStatusCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c DeviceStatusCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c DeviceStatusCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var DeviceStatus []*site.DeviceReport
	var rerr error

	// There are three options,
	if !query.HasFilters {
		DeviceStatus, rerr = siteContext.SiteDB.Queries.AllDeviceReports(ctx, nsquery.DeviceReportAll(query))
	} else {
		DeviceStatus, rerr = siteContext.SiteDB.Queries.SearchDeviceReports(ctx, nsquery.DeviceReportSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query device statuss from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(DeviceStatus, c.converter), nil
}

func (c DeviceStatusCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c DeviceStatusCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c DeviceStatusCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch)
}

// Replace will delete then create the record again, BUT it will not
func (c DeviceStatusCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	entryModel := &nightscout.DeviceStatus{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	getParams := site.GetDeviceReportParams{ID: identifier}
	deleteParams := site.HardDeleteDeviceReportParams{ID: identifier}
	createParams := nightscout.DeviceStatusToCreateDeviceReport(
		*entryModel,
		nightscout.CreateContext{CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name},
	)
	creatingNewRecord := false
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}
	defer tx.Rollback()

	// GET
	q := siteContext.SiteDB.Queries.WithTx(tx)
	dbModel, err := q.GetDeviceReport(ctx, getParams)
	if err == sql.ErrNoRows {
		// Check no Record exists
		if !canCreate {
			return nil, false, nsutils.ErrNotFound
		}
		creatingNewRecord = true
	} else if err != nil {
		slog.Error("failed to get device status by ID", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	// VALIDATE ATTRIBUTES
	if !creatingNewRecord && dbModel != nil {
		if dbModel.CreatedByDeviceName != createParams.CreatedByDeviceName {
			slog.Error("cannot change created_by_device_name attribute on device status", "existingCreatedByDeviceName", dbModel.CreatedByDeviceName, "newCreatedByDeviceName", createParams.CreatedByDeviceName)
			return nil, false, nsutils.ErrCannotChangeDevice
		}
	}

	createParams.ID = identifier

	// DELETE record if we are not creating a new record
	if !creatingNewRecord {
		_, err = q.HardDeleteDeviceReport(ctx, deleteParams)
		if err != nil {
			slog.Error("failed to hard delete device status in database", "error", err)
			return nil, false, nsutils.ErrInternal
		}
	}

	// CREATE
	dbCreateModel, err := q.CreateDeviceReport(ctx, createParams)
	if err != nil {
		slog.Error("failed to create device status in database", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	if dbCreateModel == nil {
		slog.Error("failed to create device status in database, dbCreateModel is nil")
		return nil, false, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create device status in database", "error", err, "entryModel", entryModel)
		return nil, false, nsutils.ErrInternal
	}

	return ptr(c.converter(dbCreateModel)), creatingNewRecord, nil
}
