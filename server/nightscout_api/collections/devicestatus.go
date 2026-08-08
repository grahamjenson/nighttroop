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

	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteDeviceReport(ctx, site.HardDeleteDeviceReportParams{ID: identifier})
		if err != nil {
			slog.Error("failed to permanently delete device status by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for device status delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Deleted device status", "eid", identifier, "rowsAffected", rowsAffected)

	} else {
		res, err := siteContext.SiteDB.Queries.SoftDeleteDeviceReport(ctx, site.SoftDeleteDeviceReportParams{ID: identifier})
		if err != nil {
			slog.Error("failed to soft delete device status by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for device status delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Soft Deleted device status", "eid", identifier, "rowsAffected", rowsAffected)

	}
	return nil
}

func (c DeviceStatusCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
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

func (c DeviceStatusCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	entryModel := &nightscout.DeviceStatus{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	updateParams := nightscout.DeviceStatusToUpdateDeviceReport(
		*entryModel,
		nightscout.UpdateContext{UpdatedByIdentityID: siteContext.IdentityModel.ID, UpdatedByIdentityName: siteContext.IdentityModel.Name},
	)
	updateParams.ID = identifier

	ssss, _ := utils.ToMap[site.UpdateDeviceReportParams](&updateParams)
	dbUpdateModel, err := siteContext.SiteDB.Queries.UpdateDeviceReport(ctx, updateParams)
	if err != nil {
		slog.Error("failed to update device status in database", "error", err, "updateParams", ssss)
		return nil, nsutils.ErrInternal
	}
	if dbUpdateModel == nil {
		slog.Error("failed to update device status in database, dbUpdateModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbUpdateModel)), nil
}
