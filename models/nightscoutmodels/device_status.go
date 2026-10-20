package nightscoutmodels

import (
	"encoding/json"
	"log/slog"
	"nighttroop.t1d.nz/utils/types"
)

type DeviceStatus struct {
	Identifier *IdentifierString  `json:"identifier,omitempty"`
	Device     *types.ShortString `json:"device,omitempty"`

	// Uploader status
	IsCharging      *types.FlexibleBool    `json:"isCharging,omitempty"`
	Uploader        *UploaderStatus        `json:"uploader,omitempty"`
	UploaderBattery *types.FlexibleFloat64 `json:"uploaderBattery,omitempty"`

	// Pump status
	Pump *PumpStatus `json:"pump,omitempty"`

	// OpenAPS/Loop specific
	OpenAPSRaw       *json.RawMessage `json:"openaps,omitempty"`
	LoopRaw          *json.RawMessage `json:"loop,omitempty"`
	ConfigurationRaw *json.RawMessage `json:"configuration,omitempty"`

	// Date Fields
	Date      *types.DateInt    `json:"date,omitempty"` // Unix timestamp in milliseconds
	CreatedAt *types.DateString `json:"created_at,omitempty"`

	// Default String Fields
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"` // Minutes offset from UTC

	OpenAPS       *OpenAPSStatus  `json:"-"`
	Loop          *LoopStatus     `json:"-"`
	Configuration *map[string]any `json:"-"`
	RawJsonData   string          `json:"-"`
}

func (t *DeviceStatus) UnmarshalJSON(data []byte) error {
	type Alias DeviceStatus
	aux := struct {
		ID *IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
	if t.CreatedAt != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			t.Date, t.UtcOffsetMinutes = t.CreatedAt.MillisUTCOffset()
		}
	}

	// Lets do the other bits
	if t.OpenAPSRaw != nil {
		t.OpenAPS = &OpenAPSStatus{}
		err := json.Unmarshal(*t.OpenAPSRaw, t.OpenAPS)
		if err != nil {
			t.OpenAPS = nil
			slog.Warn("failed to unmarshal openaps data", "error", err)
		}
	}

	if t.LoopRaw != nil {
		t.Loop = &LoopStatus{}
		err := json.Unmarshal(*t.LoopRaw, t.Loop)
		if err != nil {
			t.Loop = nil
			slog.Warn("failed to unmarshal loop data", "error", err)
		}
	}
	if t.ConfigurationRaw != nil {
		t.Configuration = &map[string]any{}
		err := json.Unmarshal(*t.ConfigurationRaw, t.Configuration)
		if err != nil {
			t.Configuration = nil
			slog.Warn("failed to unmarshal configuration data", "error", err)
		}
	}

	return nil
}

func (t DeviceStatus) GetDevice() *string {
	return t.Device.Stringp()
}

func (t DeviceStatus) GetApp() *string {
	return t.App.Stringp()
}

func (t DeviceStatus) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t DeviceStatus) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t DeviceStatus) GetUtcOffsetMinutes() *int64 {
	return t.UtcOffsetMinutes.Int64p()
}

func (t DeviceStatus) GetIdentifier() *string {
	return t.Identifier.Stringp()
}

// /////
// LoopStatus
// /////
type LoopStatus struct {
	Predicted                   *LoopPrediction        `json:"predicted,omitempty"`
	Timestamp                   *types.DateString      `json:"timestamp,omitempty"`
	AutomaticDoseRecommendation *DoseRecommendation    `json:"automaticDoseRecommendation,omitempty"`
	IOB                         *LoopIOB               `json:"iob,omitempty"`
	Name                        *types.ShortString     `json:"name,omitempty"`
	COB                         *LoopCOB               `json:"cob,omitempty"`
	RecommendedBolus            *types.FlexibleFloat64 `json:"recommendedBolus,omitempty"`
	Enacted                     *EnactedDose           `json:"enacted,omitempty"`
	Version                     *types.ShortString     `json:"version,omitempty"`
}

type LoopPrediction struct {
	StartDate *types.DateString        `json:"startDate,omitempty"`
	Values    []*types.FlexibleFloat64 `json:"values,omitempty"`
}

type DoseRecommendation struct {
	BolusVolume *types.FlexibleFloat64 `json:"bolusVolume,omitempty"`
	Timestamp   *types.DateString      `json:"timestamp,omitempty"`
}

type LoopIOB struct {
	IOB       *types.FlexibleFloat64 `json:"iob,omitempty"`
	Timestamp *types.DateString      `json:"timestamp,omitempty"`
}

type LoopCOB struct {
	COB       *types.FlexibleFloat64 `json:"cob,omitempty"`
	Timestamp *types.DateString      `json:"timestamp,omitempty"`
}

type EnactedDose struct {
	Timestamp   *types.DateString      `json:"timestamp,omitempty"`
	Rate        *types.FlexibleFloat64 `json:"rate,omitempty"`
	Received    *types.FlexibleBool    `json:"received,omitempty"`
	Duration    *types.FlexibleFloat64 `json:"duration,omitempty"`
	BolusVolume *types.FlexibleFloat64 `json:"bolusVolume,omitempty"`
}

/////
// OpenAPS
/////

type OpenAPSStatus struct {
	IOB       *OpenAPSIob       `json:"iob,omitempty"`
	Suggested *OpenAPSSuggested `json:"suggested,omitempty"`
}

type OpenAPSIob struct {
	IOB      *types.FlexibleFloat64 `json:"iob,omitempty"`
	BasalIOB *types.FlexibleFloat64 `json:"basaliob,omitempty"`
	Activity *types.FlexibleFloat64 `json:"activity,omitempty"`
	Time     *types.DateString      `json:"time,omitempty"`
}

type OpenAPSSuggested struct {
	Algorithm         *types.ShortString     `json:"algorithm,omitempty"`
	RunningDynamicISF *types.FlexibleBool    `json:"runningDynamicIsf,omitempty"`
	Timestamp         *types.DateString      `json:"timestamp,omitempty"`
	BG                *types.FlexibleFloat64 `json:"bg,omitempty"`
	Tick              *types.ShortString     `json:"tick,omitempty"`
	EventualBG        *types.FlexibleFloat64 `json:"eventualBG,omitempty"`
	TargetBG          *types.FlexibleFloat64 `json:"targetBG,omitempty"`
	InsulinReq        *types.FlexibleFloat64 `json:"insulinReq,omitempty"`
	DeliverAt         *types.DateString      `json:"deliverAt,omitempty"`
	SensitivityRatio  *types.FlexibleFloat64 `json:"sensitivityRatio,omitempty"`
	Reason            *types.ShortString     `json:"reason,omitempty"`
	Duration          *types.FlexibleInt64   `json:"duration,omitempty"`
	Rate              *types.FlexibleFloat64 `json:"rate,omitempty"`
	PredBGs           *OpenAPSPredBGs        `json:"predBGs,omitempty"`
	COB               *types.FlexibleFloat64 `json:"COB,omitempty"`
	IOB               *types.FlexibleFloat64 `json:"IOB,omitempty"`
	VariableSens      *types.FlexibleFloat64 `json:"variable_sens,omitempty"`
	ConsoleLog        []*types.ShortString   `json:"consoleLog,omitempty"`
	ConsoleError      []*types.ShortString   `json:"consoleError,omitempty"`
	IsfMgdlForCarbs   *types.FlexibleFloat64 `json:"isfMgdlForCarbs,omitempty"`
}

type OpenAPSPredBGs struct {
	IOB []*types.FlexibleFloat64 `json:"IOB,omitempty"`
	ZT  []*types.FlexibleFloat64 `json:"ZT,omitempty"`
}

// ////
// Pump
// ////

type PumpStatus struct {
	Battery        *PumpBattery           `json:"battery,omitempty"`
	Clock          *types.DateString      `json:"clock,omitempty"`
	Extended       *PumpExtended          `json:"extended,omitempty"`
	Reservoir      *types.FlexibleFloat64 `json:"reservoir,omitempty"`
	Status         *PumpStatusInfo        `json:"status,omitempty"`
	Bolusing       *types.FlexibleBool    `json:"bolusing,omitempty"`
	SecondsFromGMT *types.FlexibleInt64   `json:"secondsFromGMT,omitempty"`
	Manufacturer   *types.ShortString     `json:"manufacturer,omitempty"`
	PumpID         *types.ShortString     `json:"pumpID,omitempty"`
	Model          *types.ShortString     `json:"model,omitempty"`
	Suspended      *types.FlexibleBool    `json:"suspended,omitempty"`
}

type PumpBattery struct {
	Percent *types.FlexibleFloat64 `json:"percent,omitempty"`
	Voltage *types.FlexibleFloat64 `json:"voltage,omitempty"`
}

type PumpExtended struct {
	Version               *types.ShortString     `json:"Version,omitempty"`
	ActiveProfile         *types.ShortString     `json:"ActiveProfile,omitempty"`
	TempBasalAbsoluteRate *types.FlexibleFloat64 `json:"TempBasalAbsoluteRate,omitempty"`
	TempBasalStart        *types.ShortString     `json:"TempBasalStart,omitempty"`
	TempBasalRemaining    *types.FlexibleInt64   `json:"TempBasalRemaining,omitempty"`
	LastBolus             *types.ShortString     `json:"LastBolus,omitempty"`
	LastBolusAmount       *types.FlexibleFloat64 `json:"LastBolusAmount,omitempty"`
	BaseBasalRate         *types.FlexibleFloat64 `json:"BaseBasalRate,omitempty"`
}

type PumpStatusInfo struct {
	Status    *types.ShortString `json:"status,omitempty"`
	Timestamp *types.DateString  `json:"timestamp,omitempty"`
}

// ///
// Uploader
// ///
type UploaderStatus struct {
	Battery   *types.FlexibleInt64 `json:"battery,omitempty"`
	Name      *types.ShortString   `json:"name,omitempty"`
	Timestamp *types.DateString    `json:"timestamp,omitempty"`
}
