package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

type DeviceStatusCollection struct {
	named[nightscoutmodels.DeviceStatus, sitedb.DeviceReport]
}

func NewDeviceStatusCollection() CollectionStore {
	return NewCollection[nightscoutmodels.DeviceStatus, sitedb.DeviceReport](DeviceStatusCollection{
		named[nightscoutmodels.DeviceStatus, sitedb.DeviceReport]{nsutils.Devicestatus, DBToNightscoutDeviceStatus},
	})
}

func (c DeviceStatusCollection) Validate(nsModel nightscoutmodels.DeviceStatus, dbModel *sitedb.DeviceReport) *nsutils.HTTPError {
	return nil
}

func (c DeviceStatusCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.DeviceReport, error) {
	createParams := NightscoutToDBCreateDeviceReport(nsModel, cc)

	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}

	return q.CreateDeviceReport(ctx, createParams)
}

func (c DeviceStatusCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus) (*sitedb.DeviceReport, error) {
	return first(q.SearchDeviceReports(ctx, sitedb.SearchDeviceReportsParams{DateEQ: nsModel.Date.Int64p(), DeviceEq: nsModel.Device.Stringp(), Limit: 1}))
}

func (c DeviceStatusCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.DeviceReport, error) {
	return q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c DeviceStatusCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error) {
	if permanent {
		return q.HardDeleteDeviceReport(ctx, sitedb.HardDeleteDeviceReportParams{ID: identifier})
	}
	return q.SoftDeleteDeviceReport(ctx, sitedb.SoftDeleteDeviceReportParams{ID: identifier})
}

func (c DeviceStatusCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, identifier string, dbModel *sitedb.DeviceReport, uc nightscoutmodels.ConverterContext) (*sitedb.DeviceReport, error) {
	updateParams := nightscoutmodels.DeviceStatusToUpdateDeviceReport(nsModel, uc)
	updateParams.ID = identifier
	return q.UpdateDeviceReport(ctx, updateParams)
}

func (c DeviceStatusCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.DeviceReport, error) {
	if !query.HasFilters {
		return q.AllDeviceReports(ctx, nsquery.DeviceReportAll(query))
	}
	return q.SearchDeviceReports(ctx, nsquery.DeviceReportSearch(query))
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
		// Loop and Trio send it in uploader, where LoopFollow reads it
		nsModel.Uploader = &nightscoutmodels.UploaderStatus{Battery: ptr(types.FlexibleInt64(*source.DeviceBatteryPercent))}
	}
	if source.OverrideJson != nil {
		nsModel.OverrideRaw = ptr(json.RawMessage([]byte(*source.OverrideJson)))
	}

	/////
	// Pump
	/////
	pump := nightscoutmodels.PumpStatus{Manufacturer: (*types.ShortString)(source.PumpManufacturer), Model: (*types.ShortString)(source.PumpModel)}
	if source.PumpBatteryPercent != nil {
		pump.Battery = &nightscoutmodels.PumpBattery{Percent: ptr(types.FlexibleFloat64(*source.PumpBatteryPercent))}
	}
	if source.PumpReservoir != nil {
		pump.Reservoir = ptr(types.FlexibleFloat64(*source.PumpReservoir))
	}
	if source.PumpClock != nil {
		pump.Clock = ptr(types.DateStringFromInt64AndUTCOffset(*source.PumpClock, nil))
	}
	if pump != (nightscoutmodels.PumpStatus{}) {
		nsModel.Pump = &pump
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
	// Loop, Trio, xDrip+ and xDrip4iOS send the phone's battery in uploader
	if source.Uploader != nil && source.Uploader.Battery != nil {
		createParams.DeviceBatteryPercent = ptr(float64(*source.Uploader.Battery))
	}
	if source.OverrideRaw != nil {
		createParams.OverrideJson = ptr(string(*source.OverrideRaw))
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
		if source.Pump.Clock != nil {
			createParams.PumpClock = source.Pump.Clock.UnixMillip()
		}
		createParams.PumpManufacturer = source.Pump.Manufacturer.Stringp()
		createParams.PumpModel = source.Pump.Model.Stringp()
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
