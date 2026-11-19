package nscollections

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"strings"

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

// Rows are its report, so a battery or reservoir level out of its column's range, as a phone sends when it does not
// know it, like the uploader battery -100 Trio sends when iOS does not know the phone's, is left out
func (c DeviceStatusCollection) Rows(nsModel nightscoutmodels.DeviceStatus) []sitedb.TableRow {
	return []sitedb.TableRow{{Table: "device_reports", Row: NightscoutToDBCreateDeviceReport(nsModel, nightscoutmodels.ConverterContext{})}}
}

func (c DeviceStatusCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.DeviceReport, error) {
	createParams, err := deviceReportParams(q, ctx, nsModel, cc)
	if err != nil {
		return nil, err
	}
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateDeviceReport(ctx, createParams)
}

// deviceReportParams are a device status's columns, with its configuration recorded
func deviceReportParams(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, cc nightscoutmodels.ConverterContext) (sitedb.CreateDeviceReportParams, error) {
	createParams := NightscoutToDBCreateDeviceReport(nsModel, cc)
	// AAPS's configuration is kept once in loop_settings, and the report refers to it
	if nsModel.Configuration != nil {
		id, err := recordLoopSettings(q, ctx, aapsApp, fromDeviceStatus, nsModel.Configuration, createParams.OccurredAt, createParams.CreatedByDeviceName)
		if err != nil {
			return createParams, err
		}
		createParams.ConfigurationID = id
	}
	return createParams, nil
}

// FallbackKey is a device status's date and device
func (c DeviceStatusCollection) FallbackKey() FallbackKey[nightscoutmodels.DeviceStatus, sitedb.DeviceReport] {
	return FallbackKey[nightscoutmodels.DeviceStatus, sitedb.DeviceReport]{
		search: func(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, limit int64) ([]*sitedb.DeviceReport, error) {
			return q.SearchDeviceReports(ctx, sitedb.SearchDeviceReportsParams{DateEQ: nsModel.Date.Int64p(), DeviceEq: nsModel.Device.Stringp(), Limit: limit})
		},
		fields: []KeyField[nightscoutmodels.DeviceStatus, sitedb.DeviceReport]{
			sameDate[nightscoutmodels.DeviceStatus, sitedb.DeviceReport](),
			sameDevice[nightscoutmodels.DeviceStatus, sitedb.DeviceReport](),
		},
	}
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

// RewriteModel keeps what the report's document cannot give back unless it is sent: its pump and loop status, whose
// JSON is kept only on its device's newest report, its configuration, kept in loop_settings and given as the identity
// may see it, and the pump's charging, which only a connector sets
func (c DeviceStatusCollection) RewriteModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.DeviceStatus, sent map[string]any, stored *sitedb.DeviceReport, cc nightscoutmodels.ConverterContext) (*sitedb.DeviceReport, error) {
	params, err := deviceReportParams(q, ctx, nsModel, cc)
	if err != nil {
		return nil, err
	}
	if _, ok := sent["pump"]; !ok {
		params.PumpReservoir, params.PumpBatteryPercent, params.PumpClock = stored.PumpReservoir, stored.PumpBatteryPercent, stored.PumpClock
		params.PumpManufacturer, params.PumpModel, params.PumpID = stored.PumpManufacturer, stored.PumpModel, stored.PumpID
		params.PumpSuspended, params.PumpBolusing, params.PumpJson = stored.PumpSuspended, stored.PumpBolusing, stored.PumpJson
		params.PumpState = stored.PumpState
	}
	_, sentOpenAPS := sent["openaps"]
	if _, sentLoop := sent["loop"]; !sentOpenAPS && !sentLoop {
		params.LoopName, params.LoopStatusJson, params.LoopIob = stored.LoopName, stored.LoopStatusJson, stored.LoopIob
		params.LoopIobBasal, params.LoopIobTreatment, params.LoopCob = stored.LoopIobBasal, stored.LoopIobTreatment, stored.LoopCob
	}
	if _, ok := sent["configuration"]; !ok {
		params.ConfigurationID = stored.ConfigurationID
	}
	params.ID, params.PumpChargingState = stored.ID, stored.PumpChargingState
	return q.RewriteDeviceReport(ctx, sitedb.RewriteDeviceReportParams(params))
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

// DBToNightscoutDeviceStatus is a device report as Nightscout's device status: its columns, then its uploader, pump,
// configuration and loop, each read from what was kept as sent and filled in from the columns
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
	if source.OverrideJson != nil {
		nsModel.OverrideRaw = ptr(json.RawMessage(*source.OverrideJson))
	}
	nsModel.Uploader = uploaderStatus(source)
	nsModel.Pump = pumpStatus(source)
	// Its configuration is the loop_settings it refers to, left out when they were deleted
	if configuration := loopSettingsJSON(context.Background(), siteContext, source.ConfigurationID); configuration != nil {
		nsModel.ConfigurationRaw = &configuration
	}
	if status := loopStatus(source); status != nil && *source.LoopName == "openaps" {
		nsModel.OpenAPSRaw = &status
	} else if status != nil && *source.LoopName == "loop" {
		nsModel.LoopRaw = &status
	}
	return nsModel
}

// uploaderStatus is the uploader kept as sent, through the model so only known fields go out, with the phone's battery
// and charging from their columns, as Loop and Trio send them in uploader, where LoopFollow reads them. It is nil when
// it has none of these.
func uploaderStatus(source *sitedb.DeviceReport) *nightscoutmodels.UploaderStatus {
	uploader := nightscoutmodels.UploaderStatus{}
	if source.UploaderJson != nil {
		_ = json.Unmarshal([]byte(*source.UploaderJson), &uploader)
	}
	if uploader.Battery == nil && source.DeviceBatteryPercent != nil {
		uploader.Battery = ptr(types.FlexibleInt64(*source.DeviceBatteryPercent))
	}
	if uploader.IsCharging == nil && source.IsDeviceCharging != nil {
		uploader.IsCharging = ptr(types.FlexibleBool(*source.IsDeviceCharging))
	}
	if uploader == (nightscoutmodels.UploaderStatus{}) {
		return nil
	}
	return &uploader
}

// pumpStatus is the pump kept as sent, through the model so only known fields go out, such as AAPS's extended with the
// profile and temp basal Nightguard shows. The columns fill in the fields it does not have, as for a pump kept before
// its JSON was. It is nil when it has none of these.
func pumpStatus(source *sitedb.DeviceReport) *nightscoutmodels.PumpStatus {
	pump := nightscoutmodels.PumpStatus{}
	if source.PumpJson != nil {
		_ = json.Unmarshal([]byte(*source.PumpJson), &pump)
	}
	pump.Manufacturer = cmp.Or(pump.Manufacturer, (*types.ShortString)(source.PumpManufacturer))
	pump.Model = cmp.Or(pump.Model, (*types.ShortString)(source.PumpModel))
	pump.PumpID = cmp.Or(pump.PumpID, (*types.ShortString)(source.PumpID))
	// Followers show a suspended pump's insulin delivery as suspended
	pump.Suspended = cmp.Or(pump.Suspended, (*types.FlexibleBool)(source.PumpSuspended))
	pump.Bolusing = cmp.Or(pump.Bolusing, (*types.FlexibleBool)(source.PumpBolusing))
	if pump.Battery == nil && source.PumpBatteryPercent != nil {
		pump.Battery = &nightscoutmodels.PumpBattery{Percent: ptr(types.FlexibleFloat64(*source.PumpBatteryPercent))}
	}
	if pump.Reservoir == nil && source.PumpReservoir != nil {
		pump.Reservoir = ptr(types.FlexibleFloat64(*source.PumpReservoir))
	}
	if pump.Clock == nil && source.PumpClock != nil {
		pump.Clock = ptr(types.DateStringFromInt64AndUTCOffset(*source.PumpClock, nil))
	}
	if pump == (nightscoutmodels.PumpStatus{}) {
		return nil
	}
	return &pump
}

// loopStatus is the loop's status, given back as its openaps or loop object: as it was sent, without its predictions
// once a day old, or, for one cleared whole by an earlier truncate, as what was extracted of it, its IOB and COB. It is
// nil without a loop.
func loopStatus(source *sitedb.DeviceReport) json.RawMessage {
	if source.LoopName == nil {
		return nil
	}
	if source.LoopStatusJson != nil {
		return json.RawMessage(*source.LoopStatusJson)
	}
	return extractedLoopStatus(*source.LoopName, source)
}

// extractedLoopStatus is a loop's status as its IOB and COB columns keep it, where the loop sends them, nil without them
func extractedLoopStatus(loopName string, source *sitedb.DeviceReport) json.RawMessage {
	iob, status := map[string]any{}, map[string]any{}
	if source.LoopIob != nil {
		iob["iob"] = *source.LoopIob
	}
	switch loopName {
	case "openaps":
		if source.LoopIobBasal != nil {
			iob["basaliob"] = *source.LoopIobBasal
		}
		if source.LoopCob != nil {
			status["suggested"] = map[string]any{"COB": *source.LoopCob}
		}
	case "loop":
		if source.LoopCob != nil {
			status["cob"] = map[string]any{"cob": *source.LoopCob}
		}
	}
	if len(iob) > 0 {
		status["iob"] = iob
	}
	if len(status) == 0 {
		return nil
	}
	raw, _ := json.Marshal(status)
	return raw
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
		createParams.OverrideJson = ptr(string(*source.OverrideRaw))
	}
	// The pump and uploader are kept as the model reads them, so only their known fields, for those without a column
	createParams.PumpJson, createParams.UploaderJson = jsonp(source.Pump), jsonp(source.Uploader)

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
		createParams.PumpID = source.Pump.PumpID.Stringp()
		createParams.PumpSuspended = (*bool)(source.Pump.Suspended)
		createParams.PumpBolusing = (*bool)(source.Pump.Bolusing)
		createParams.PumpState = pumpState(source.Pump)
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

// pumpState is "suspended" when the pump says it is: Loop in its suspended, Trio in its status's suspended, and AAPS
// with its status's status. It is "normal" when the pump says it is not, and nil when the pump does not say.
func pumpState(pump *nightscoutmodels.PumpStatus) *string {
	suspended := pump.Suspended
	var status *string
	if pump.Status != nil {
		suspended, status = cmp.Or(suspended, pump.Status.Suspended), pump.Status.Status.Stringp()
	}
	switch {
	case (suspended != nil && bool(*suspended)) || (status != nil && strings.EqualFold(*status, "suspended")):
		return ptr("suspended")
	case suspended != nil || status != nil:
		return ptr("normal")
	}
	return nil
}

// setOpenAPSColumns keeps an OpenAPS loop's status, with its predictions, and its IOB and COB
func setOpenAPSColumns(params *sitedb.CreateDeviceReportParams, openAPS *nightscoutmodels.OpenAPSStatus, raw json.RawMessage) {
	params.LoopStatusJson = ptr(string(raw))
	params.LoopName = ptr("openaps")
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
	params.LoopStatusJson = ptr(string(raw))
	params.LoopName = ptr("loop")
	if loop.IOB != nil {
		params.LoopIob = loop.IOB.IOB.Float64p()
	}
	if loop.COB != nil {
		params.LoopCob = loop.COB.COB.Float64p()
	}
}

// jsonp is v as JSON, nil when it is
func jsonp[T any](v *T) *string {
	if v == nil {
		return nil
	}
	raw, _ := json.Marshal(v)
	return ptr(string(raw))
}
