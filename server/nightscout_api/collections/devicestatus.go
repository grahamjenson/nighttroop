package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
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

func (c DeviceStatusCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.DeviceStatus, cc nightscout.CreateContext, replacing *site.DeviceReport, overrideIdentifier *string) (*site.DeviceReport, error) {
	createParams := nightscout.DeviceStatusToCreateDeviceReport(nsModel, cc)
	// Validate change if replacing a model
	if replacing != nil {
		if replacing.CreatedByDeviceName != nil && createParams.CreatedByDeviceName != nil && *replacing.CreatedByDeviceName != *createParams.CreatedByDeviceName {
			slog.Error("cannot change created_by_device_name attribute on device status", "existingCreatedByDeviceName", replacing.CreatedByDeviceName, "newCreatedByDeviceName", createParams.CreatedByDeviceName)
			return nil, nsutils.ErrCannotChangeDevice
		}

	}

	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}

	return q.CreateDeviceReport(ctx, createParams)
}

func (c DeviceStatusCollection) UpdateSBFunc(q *site.Queries, ctx context.Context, updateParams site.UpdateDeviceReportParams) (*site.DeviceReport, error) {
	return q.UpdateDeviceReport(ctx, updateParams)
}

func (c DeviceStatusCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.DeviceReport, error) {
	return q.GetDeviceReport(ctx, site.GetDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteDeviceReport(ctx, site.HardDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteDeviceReport(ctx, site.SoftDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.DeviceStatus, identifier string, uc nightscout.UpdateContext) (*site.DeviceReport, error) {
	updateParams := nightscout.DeviceStatusToUpdateDeviceReport(nsModel, uc)
	updateParams.ID = identifier
	return q.UpdateDeviceReport(ctx, updateParams)
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
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate)
}
