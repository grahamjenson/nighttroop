package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type DeviceStatusCollection struct {
	converter func(c *site.DeviceReport, raw bool) nsutils.Record
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

func (c DeviceStatusCollection) Convert(dr *site.DeviceReport, raw bool) nsutils.Record {
	return c.converter(dr, raw)
}

func (c DeviceStatusCollection) Validate(nsModel nightscout.DeviceStatus, dbModel *site.DeviceReport) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}
	return nil
}

func (c DeviceStatusCollection) ValidatePatch(nsModel nightscout.DeviceStatus, dbModel *site.DeviceReport) *nsutils.HTTPError {
	return nil
}

func (c DeviceStatusCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.DeviceStatus, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.DeviceReport, error) {
	createParams := NightscoutToDBCreateDeviceReport(nsModel, cc)

	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}

	return q.CreateDeviceReport(ctx, createParams)
}

func (c DeviceStatusCollection) GetDuplicateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.DeviceStatus) (*site.DeviceReport, error) {
	deviceStatuses, rerr := q.SearchDeviceReports(ctx, site.SearchDeviceReportsParams{
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

func (c DeviceStatusCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.DeviceReport, error) {
	return q.GetDeviceReport(ctx, site.GetDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteDeviceReport(ctx, site.HardDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteDeviceReport(ctx, site.SoftDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.DeviceStatus, identifier string, dbModel *site.DeviceReport, uc nightscout.ConverterContext) (*site.DeviceReport, error) {
	updateParams := nightscout.DeviceStatusToUpdateDeviceReport(nsModel, uc)
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
	return MetaGet(c, ctx, siteContext, identifier, query.Raw)
}

func (c DeviceStatusCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var deviceStatuses []*site.DeviceReport
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

	return nsutils.CreateRecordSlice(deviceStatuses, c.converter, query.Raw), nil
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

func DBToNightscoutDeviceStatus(source *site.DeviceReport) nightscout.DeviceStatus {
	if source == nil {
		return nightscout.DeviceStatus{}
	}

	nsModel := nightscout.DBModelToDeviceStatus(source)

	if source.IsDeviceCharging != nil {
		nsModel.IsCharging = ptr(nightscout.FlexibleBool(*source.IsDeviceCharging))
	}
	if source.DeviceBatteryPercent != nil {
		nsModel.UploaderBattery = ptr(nightscout.FlexibleFloat64(*source.DeviceBatteryPercent))
	}

	///////
	/// Extracting values
	var predictionsDefault []*nightscout.FlexibleFloat64
	if source.LoopPredictionsDefault != nil {
		err := json.Unmarshal([]byte(*source.LoopPredictionsDefault), &predictionsDefault)
		if err != nil {
			slog.Error("failed to unmarshal loop predictions", "error", err)
		}
	}

	var predictionsZt []*nightscout.FlexibleFloat64
	if source.LoopPredictionsZt != nil {
		err := json.Unmarshal([]byte(*source.LoopPredictionsZt), &predictionsZt)
		if err != nil {
			slog.Error("failed to unmarshal loop predictions", "error", err)
		}
	}

	var iob *nightscout.FlexibleFloat64
	if source.LoopIob != nil {
		iob = ptr(nightscout.FlexibleFloat64(*source.LoopIob))
	}
	var basalIob *nightscout.FlexibleFloat64
	if source.LoopIobBasal != nil {
		basalIob = ptr(nightscout.FlexibleFloat64(*source.LoopIobBasal))
	}
	var cob *nightscout.FlexibleFloat64
	if source.LoopCob != nil {
		cob = ptr(nightscout.FlexibleFloat64(*source.LoopCob))
	}

	//////
	if source.LoopName != nil {
		switch *source.LoopName {
		case "openaps":

			nsModel.OpenAPS = &nightscout.OpenAPSStatus{
				IOB: &nightscout.OpenAPSIob{
					IOB:      iob,
					BasalIOB: basalIob,
				},
				Suggested: &nightscout.OpenAPSSuggested{
					COB: cob,

					PredBGs: &nightscout.OpenAPSPredBGs{
						IOB: predictionsDefault,
						ZT:  predictionsZt,
					},
				},
			}
			if source.LoopPredictionsStartTime != nil {
				nsModel.OpenAPS.Suggested.Timestamp = ptr(nightscout.DateStringFromInt64AndUTCOffset(*source.LoopPredictionsStartTime, nil))
			}
		case "loop":
			nsModel.Loop = &nightscout.LoopStatus{
				IOB: &nightscout.LoopIOB{
					IOB: iob,
				},
				COB: &nightscout.LoopCOB{
					COB: cob,
				},
				Predicted: &nightscout.LoopPrediction{
					Values: predictionsDefault,
				},
			}
		}

	}

	if source.PumpBatteryPercent != nil {
		nsModel.Pump = &nightscout.PumpStatus{
			Battery: &nightscout.PumpBattery{
				Percent: ptr(nightscout.FlexibleFloat64(*source.PumpBatteryPercent)),
			},
		}
	}

	if source.PumpReservoir != nil {
		if nsModel.Pump == nil {
			nsModel.Pump = &nightscout.PumpStatus{}
		}
		nsModel.Pump.Reservoir = ptr(nightscout.FlexibleFloat64(*source.PumpReservoir))
	}
	return nsModel
}

///////
// NIGHTSCOUT TO DB
///////

func NightscoutToDBCreateDeviceReport(source nightscout.DeviceStatus, ctx nightscout.ConverterContext) site.CreateDeviceReportParams {
	createParams := nightscout.DeviceStatusToCreateDeviceReport(source, ctx)
	if source.IsCharging != nil {
		createParams.IsDeviceCharging = ptr(bool(*source.IsCharging))
	}
	if source.UploaderBattery != nil {
		createParams.DeviceBatteryPercent = ptr(float64(*source.UploaderBattery))
	}

	////
	// OpenAPS
	////
	if source.OpenAPS != nil {

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

	/////
	// Uploader
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
	return createParams
}
