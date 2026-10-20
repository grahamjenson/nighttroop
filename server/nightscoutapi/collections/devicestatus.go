package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

type DeviceStatusCollection struct {
	converter func(c *sitedb.DeviceReport, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewDeviceStatusCollection() DeviceStatusCollection {
	return DeviceStatusCollection{
		converter: nsutils.DBModelToRecord(nsutils.Devicestatus, DBToNightscoutDeviceStatus),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///
func (c DeviceStatusCollection) Name() string {
	return "devicestatus"
}

func (c DeviceStatusCollection) Convert(dr *sitedb.DeviceReport, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c DeviceStatusCollection) Validate(nsModel nightscoutmodels.DeviceStatus, dbModel *sitedb.DeviceReport) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}
	return nil
}

func (c DeviceStatusCollection) ValidatePatch(nsModel nightscoutmodels.DeviceStatus, dbModel *sitedb.DeviceReport) *nsutils.HTTPError {
	return nil
}

func (c DeviceStatusCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.DeviceReport, error) {
	createParams := NightscoutToDBCreateDeviceReport(nsModel, cc)

	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}

	return q.CreateDeviceReport(ctx, createParams)
}

func (c DeviceStatusCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus) (*sitedb.DeviceReport, error) {
	// Without a date every filter below is empty and would match the newest row
	if nsModel.Date == nil {
		return nil, nil
	}
	deviceStatuses, rerr := q.SearchDeviceReports(ctx, sitedb.SearchDeviceReportsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(deviceStatuses) > 0 {
		return deviceStatuses[0], nil
	}
	return nil, nil
}

func (c DeviceStatusCollection) GetDBFunc(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.DeviceReport, error) {
	return q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c DeviceStatusCollection) HardDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteDeviceReport(ctx, sitedb.HardDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) SoftDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteDeviceReport(ctx, sitedb.SoftDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) PatchDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, identifier string, dbModel *sitedb.DeviceReport, uc nightscoutmodels.ConverterContext) (*sitedb.DeviceReport, error) {
	updateParams := nightscoutmodels.DeviceStatusToUpdateDeviceReport(nsModel, uc)
	updateParams.ID = identifier
	return q.UpdateDeviceReport(ctx, updateParams)
}

/////
// END OF FUNCTIONS USED BY META COLLECTIONS
/////

func (c DeviceStatusCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c DeviceStatusCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c DeviceStatusCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var deviceStatuses []*sitedb.DeviceReport
	var rerr error

	// There are three options,
	if !query.HasFilters {
		deviceStatuses, rerr = siteContext.SiteDB.Queries.AllDeviceReports(ctx, nsquery.DeviceReportAll(query))
	} else {
		deviceStatuses, rerr = siteContext.SiteDB.Queries.SearchDeviceReports(ctx, nsquery.DeviceReportSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query device statuss from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(deviceStatuses, c.converter, siteContext, query.Raw), nil
}

func (c DeviceStatusCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c DeviceStatusCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c DeviceStatusCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch, raw)
}

// Replace will delete then create the record again, BUT it will not
func (c DeviceStatusCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}

/////
// Custom Model Stuff
/////

///////
// DB TO NIGHTSCOUT
///////

func DBToNightscoutDeviceStatus(source *sitedb.DeviceReport, siteContext *httputils.SiteContext) nightscoutmodels.DeviceStatus {
	if source == nil {
		return nightscoutmodels.DeviceStatus{}
	}

	nsModel := nightscoutmodels.DBModelToDeviceStatus(source, siteContext)

	if source.IsDeviceCharging != nil {
		nsModel.IsCharging = ptr(types.FlexibleBool(*source.IsDeviceCharging))
	}
	if source.DeviceBatteryPercent != nil {
		nsModel.UploaderBattery = ptr(types.FlexibleFloat64(*source.DeviceBatteryPercent))
	}

	/////
	// Pump
	/////
	if source.PumpBatteryPercent != nil {
		nsModel.Pump = &nightscoutmodels.PumpStatus{
			Battery: &nightscoutmodels.PumpBattery{
				Percent: ptr(types.FlexibleFloat64(*source.PumpBatteryPercent)),
			},
		}
	}

	if source.PumpReservoir != nil {
		if nsModel.Pump == nil {
			nsModel.Pump = &nightscoutmodels.PumpStatus{}
		}
		nsModel.Pump.Reservoir = ptr(types.FlexibleFloat64(*source.PumpReservoir))
	}

	///////
	/// Extracting Loop values
	///////
	if source.ConfigurationSettingsJson != nil {
		nsModel.ConfigurationRaw = ptr(json.RawMessage([]byte(*source.ConfigurationSettingsJson)))
	}

	if source.LoopName != nil && source.LoopSettingsJson != nil {
		switch *source.LoopName {
		case "openaps":
			nsModel.OpenAPSRaw = ptr(json.RawMessage([]byte(*source.LoopSettingsJson)))
		case "loop":
			nsModel.LoopRaw = ptr(json.RawMessage([]byte(*source.LoopSettingsJson)))
		}
	}

	return nsModel
}

///////
// NIGHTSCOUT TO DB
///////

func NightscoutToDBCreateDeviceReport(source nightscoutmodels.DeviceStatus, ctx nightscoutmodels.ConverterContext) sitedb.CreateDeviceReportParams {
	createParams := nightscoutmodels.DeviceStatusToCreateDeviceReport(source, ctx)
	if source.IsCharging != nil {
		createParams.IsDeviceCharging = ptr(bool(*source.IsCharging))
	}
	if source.UploaderBattery != nil {
		createParams.DeviceBatteryPercent = ptr(float64(*source.UploaderBattery))
	}

	if source.ConfigurationRaw != nil {
		createParams.ConfigurationSettingsJson = ptr(string(*source.ConfigurationRaw))
	}

	/////
	// Pumps
	/////
	if source.Pump != nil {
		if source.Pump.Battery != nil {
			if source.Pump.Battery.Percent != nil {
				createParams.PumpBatteryPercent = ptr(float64(*source.Pump.Battery.Percent))
			}
		}
		if source.Pump.Reservoir != nil {
			createParams.PumpReservoir = ptr(float64(*source.Pump.Reservoir))
		}
	}

	////
	// OpenAPS
	////
	if source.OpenAPS != nil {
		createParams.LoopSettingsJson = ptr(string(*source.OpenAPSRaw))

		createParams.LoopName = ptr("openaps")
		if source.OpenAPS.IOB != nil {
			createParams.LoopIob = source.OpenAPS.IOB.IOB.Float64p()
			createParams.LoopIobBasal = source.OpenAPS.IOB.BasalIOB.Float64p()
			if source.OpenAPS.IOB.IOB != nil && source.OpenAPS.IOB.BasalIOB != nil {
				createParams.LoopIobTreatment = ptr(source.OpenAPS.IOB.IOB.Float64() - source.OpenAPS.IOB.BasalIOB.Float64())
			}
		}

		if source.OpenAPS.Suggested != nil {
			createParams.LoopCob = source.OpenAPS.Suggested.COB.Float64p()

			if source.OpenAPS.Suggested.Timestamp != nil {
				createParams.LoopPredictionsStartTime = source.OpenAPS.Suggested.Timestamp.UnixMillip()
			}
			if source.OpenAPS.Suggested.PredBGs != nil {
				if source.OpenAPS.Suggested.PredBGs.IOB != nil {
					out, err := json.Marshal(source.OpenAPS.Suggested.PredBGs.IOB)
					if err == nil {
						createParams.LoopPredictionsDefault = ptr(string(out))
					}
				}

				if source.OpenAPS.Suggested.PredBGs.ZT != nil {
					out, err := json.Marshal(source.OpenAPS.Suggested.PredBGs.ZT)
					if err == nil {
						createParams.LoopPredictionsZt = ptr(string(out))
					}
				}

			}
		}

	}

	/////
	// Loop
	/////
	if source.Loop != nil {
		createParams.LoopSettingsJson = ptr(string(*source.LoopRaw))
		createParams.LoopName = ptr("loop")
		if source.Loop.IOB != nil {
			createParams.LoopIob = source.Loop.IOB.IOB.Float64p()
		}

		if source.Loop.COB != nil {
			createParams.LoopCob = source.Loop.COB.COB.Float64p()
		}

		if source.Uploader != nil {
			if source.Uploader.Battery != nil {
				createParams.DeviceBatteryPercent = ptr(float64(*source.Uploader.Battery))
			}
		}

		if source.Loop.Predicted != nil {
			if source.Loop.Predicted.Values != nil {
				out, err := json.Marshal(source.Loop.Predicted.Values)
				if err == nil {
					createParams.LoopPredictionsDefault = ptr(string(out))
				}
			}
			if source.Loop.Predicted.StartDate != nil {
				createParams.LoopPredictionsStartTime = source.Loop.Predicted.StartDate.UnixMillip()
			}
		}

	}

	return createParams
}
