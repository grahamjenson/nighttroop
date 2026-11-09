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
	// AAPS's configuration is kept once in loop_settings, and the report refers to it
	if nsModel.Configuration != nil {
		id, err := recordLoopSettings(q, ctx, aapsApp, fromDeviceStatus, nsModel.Configuration, createParams.OccurredAt, createParams.CreatedByDeviceName)
		if err != nil {
			return nil, err
		}
		createParams.ConfigurationID = id
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

func (c DeviceStatusCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.DeviceReportAscendingIDs(ctx, nsquery.DeviceReportAscending(query))
}

func (c DeviceStatusCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.DeviceReportHistoryIDs(ctx, sitedb.DeviceReportHistoryIDsParams{Since: since, Limit: limit})
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
	}
	other := source.OtherAttributes
	if other.Override != nil {
		nsModel.OverrideRaw = &other.Override
	}

	/////
	// Pump
	/////
	// Its fields without a column, like AAPS's extended with the profile and temp basal Nightguard shows, are read
	// from the pump and uploader kept as sent, through the model so only known fields go out; the columns fill in the
	// rest, as for one kept before them
	pump := nightscoutmodels.PumpStatus{}
	uploader := nightscoutmodels.UploaderStatus{}
	if other.Pump != nil {
		_ = json.Unmarshal(other.Pump, &pump)
	}
	if other.Uploader != nil {
		_ = json.Unmarshal(other.Uploader, &uploader)
	}
	// Loop and Trio send the phone's battery and charging in uploader, where LoopFollow reads them
	if uploader.Battery == nil && source.DeviceBatteryPercent != nil {
		uploader.Battery = ptr(types.FlexibleInt64(*source.DeviceBatteryPercent))
	}
	if uploader.IsCharging == nil && source.IsDeviceCharging != nil {
		uploader.IsCharging = ptr(types.FlexibleBool(*source.IsDeviceCharging))
	}
	if uploader != (nightscoutmodels.UploaderStatus{}) {
		nsModel.Uploader = &uploader
	}
	if pump.Manufacturer == nil && other.PumpManufacturer != nil {
		pump.Manufacturer = (*types.ShortString)(other.PumpManufacturer)
	}
	if pump.Model == nil && other.PumpModel != nil {
		pump.Model = (*types.ShortString)(other.PumpModel)
	}
	if pump.Battery == nil && source.PumpBatteryPercent != nil {
		pump.Battery = &nightscoutmodels.PumpBattery{Percent: ptr(types.FlexibleFloat64(*source.PumpBatteryPercent))}
	}
	if pump.Reservoir == nil && source.PumpReservoir != nil {
		pump.Reservoir = ptr(types.FlexibleFloat64(*source.PumpReservoir))
	}
	if pump.Clock == nil && other.PumpClock != nil {
		pump.Clock = ptr(types.DateStringFromInt64AndUTCOffset(*other.PumpClock, nil))
	}
	if pump != (nightscoutmodels.PumpStatus{}) {
		nsModel.Pump = &pump
	}

	///////
	/// Extracting Loop values
	///////
	// Its configuration is the loop_settings it refers to, left out when they were deleted
	if configuration := loopSettingsJSON(context.Background(), siteContext, source.ConfigurationID); configuration != nil {
		nsModel.ConfigurationRaw = &configuration
	}

	if other.LoopName != nil && other.LoopSettings != nil {
		switch *other.LoopName {
		case "openaps":
			nsModel.OpenAPSRaw = &other.LoopSettings
		case "loop":
			nsModel.LoopRaw = &other.LoopSettings
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
	if source.Uploader != nil && source.Uploader.IsCharging != nil {
		createParams.IsDeviceCharging = ptr(bool(*source.Uploader.IsCharging))
	}
	if source.OverrideRaw != nil {
		createParams.OtherAttributes.Override = *source.OverrideRaw
	}
	// The pump and uploader are kept as the model reads them, so only their known fields, for those without a column
	if source.Pump != nil {
		createParams.OtherAttributes.Pump, _ = json.Marshal(source.Pump)
	}
	if source.Uploader != nil {
		createParams.OtherAttributes.Uploader, _ = json.Marshal(source.Uploader)
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
			createParams.OtherAttributes.PumpClock = source.Pump.Clock.UnixMillip()
		}
		createParams.OtherAttributes.PumpManufacturer = source.Pump.Manufacturer.Stringp()
		createParams.OtherAttributes.PumpModel = source.Pump.Model.Stringp()
	}

	// Loop's are set last, so a status with both keeps Loop's settings, IOB, COB and predictions over OpenAPS's
	if source.OpenAPS != nil {
		setOpenAPSColumns(&createParams, source.OpenAPS, *source.OpenAPSRaw)
	}
	if source.Loop != nil {
		setLoopColumns(&createParams, source.Loop, *source.LoopRaw)
	}

	return createParams
}

// setOpenAPSColumns keeps an OpenAPS loop's status, with its predictions, and its IOB and COB
func setOpenAPSColumns(params *sitedb.CreateDeviceReportParams, openAPS *nightscoutmodels.OpenAPSStatus, raw json.RawMessage) {
	params.OtherAttributes.LoopSettings = raw
	params.OtherAttributes.LoopName = ptr("openaps")
	if iob := openAPS.IOB; iob != nil {
		params.LoopIob = iob.IOB.Float64p()
		params.LoopIobBasal = iob.BasalIOB.Float64p()
		// The IOB not from basal is from treatments
		if iob.IOB != nil && iob.BasalIOB != nil {
			params.LoopIobTreatment = ptr(iob.IOB.Float64() - iob.BasalIOB.Float64())
		}
	}

	if suggested := openAPS.Suggested; suggested != nil {
		params.LoopCob = suggested.COB.Float64p()
	}
}

// setLoopColumns keeps Loop's status, with its predictions, and its IOB and COB
func setLoopColumns(params *sitedb.CreateDeviceReportParams, loop *nightscoutmodels.LoopStatus, raw json.RawMessage) {
	params.OtherAttributes.LoopSettings = raw
	params.OtherAttributes.LoopName = ptr("loop")
	if loop.IOB != nil {
		params.LoopIob = loop.IOB.IOB.Float64p()
	}
	if loop.COB != nil {
		params.LoopCob = loop.COB.COB.Float64p()
	}
}
