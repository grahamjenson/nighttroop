package nightscout

import (
	"encoding/json"
)

type DeviceStatus struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Device     *ShortString      `json:"device,omitempty"`

	// Uploader status
	IsCharging      *FlexibleBool    `json:"isCharging,omitempty"`
	Uploader        *UploaderStatus  `json:"uploader,omitempty"`
	UploaderBattery *FlexibleFloat64 `json:"uploaderBattery,omitempty"`

	// Pump status
	Pump *PumpStatus `json:"pump,omitempty"`

	// OpenAPS/Loop specific
	OpenAPS *OpenAPSStatus `json:"openaps,omitempty"`
	Loop    *LoopStatus    `json:"loop,omitempty"`

	// Date Fields
	Date      *DateInt    `json:"date,omitempty"` // Unix timestamp in milliseconds
	CreatedAt *DateString `json:"created_at,omitempty"`

	// Default String Fields
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	UtcOffsetMinutes *FlexibleInt64 `json:"utcOffset,omitempty"` // Minutes offset from UTC
	RawJsonData      string         `json:"-"`
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
	Predicted                   *LoopPrediction     `json:"predicted,omitempty"`
	Timestamp                   *DateString         `json:"timestamp,omitempty"`
	AutomaticDoseRecommendation *DoseRecommendation `json:"automaticDoseRecommendation,omitempty"`
	IOB                         *LoopIOB            `json:"iob,omitempty"`
	Name                        *ShortString        `json:"name,omitempty"`
	COB                         *LoopCOB            `json:"cob,omitempty"`
	RecommendedBolus            *FlexibleFloat64    `json:"recommendedBolus,omitempty"`
	Enacted                     *EnactedDose        `json:"enacted,omitempty"`
	Version                     *ShortString        `json:"version,omitempty"`
}

type LoopPrediction struct {
	StartDate *DateString        `json:"startDate,omitempty"`
	Values    []*FlexibleFloat64 `json:"values,omitempty"`
}

type DoseRecommendation struct {
	BolusVolume *FlexibleFloat64 `json:"bolusVolume,omitempty"`
	Timestamp   *DateString      `json:"timestamp,omitempty"`
}

type LoopIOB struct {
	IOB       *FlexibleFloat64 `json:"iob,omitempty"`
	Timestamp *DateString      `json:"timestamp,omitempty"`
}

type LoopCOB struct {
	COB       *FlexibleFloat64 `json:"cob,omitempty"`
	Timestamp *DateString      `json:"timestamp,omitempty"`
}

type EnactedDose struct {
	Timestamp   *DateString      `json:"timestamp,omitempty"`
	Rate        *FlexibleFloat64 `json:"rate,omitempty"`
	Received    *FlexibleBool    `json:"received,omitempty"`
	Duration    *FlexibleFloat64 `json:"duration,omitempty"`
	BolusVolume *FlexibleFloat64 `json:"bolusVolume,omitempty"`
}

/////
// OpenAPS
/////

type OpenAPSStatus struct {
	IOB       *OpenAPSIob       `json:"iob,omitempty"`
	Suggested *OpenAPSSuggested `json:"suggested,omitempty"`
}

type OpenAPSIob struct {
	IOB      *FlexibleFloat64 `json:"iob,omitempty"`
	BasalIOB *FlexibleFloat64 `json:"basaliob,omitempty"`
	Activity *FlexibleFloat64 `json:"activity,omitempty"`
	Time     *DateString      `json:"time,omitempty"`
}

type OpenAPSSuggested struct {
	Algorithm         *ShortString     `json:"algorithm,omitempty"`
	RunningDynamicISF *FlexibleBool    `json:"runningDynamicIsf,omitempty"`
	Timestamp         *DateString      `json:"timestamp,omitempty"`
	BG                *FlexibleFloat64 `json:"bg,omitempty"`
	Tick              *ShortString     `json:"tick,omitempty"`
	EventualBG        *FlexibleFloat64 `json:"eventualBG,omitempty"`
	TargetBG          *FlexibleFloat64 `json:"targetBG,omitempty"`
	InsulinReq        *FlexibleFloat64 `json:"insulinReq,omitempty"`
	DeliverAt         *DateString      `json:"deliverAt,omitempty"`
	SensitivityRatio  *FlexibleFloat64 `json:"sensitivityRatio,omitempty"`
	Reason            *ShortString     `json:"reason,omitempty"`
	Duration          *FlexibleInt64   `json:"duration,omitempty"`
	Rate              *FlexibleFloat64 `json:"rate,omitempty"`
	PredBGs           *OpenAPSPredBGs  `json:"predBGs,omitempty"`
	COB               *FlexibleFloat64 `json:"COB,omitempty"`
	IOB               *FlexibleFloat64 `json:"IOB,omitempty"`
	VariableSens      *FlexibleFloat64 `json:"variable_sens,omitempty"`
	ConsoleLog        []*ShortString   `json:"consoleLog,omitempty"`
	ConsoleError      []*ShortString   `json:"consoleError,omitempty"`
	IsfMgdlForCarbs   *FlexibleFloat64 `json:"isfMgdlForCarbs,omitempty"`
}

type OpenAPSPredBGs struct {
	IOB []*FlexibleFloat64 `json:"IOB,omitempty"`
	ZT  []*FlexibleFloat64 `json:"ZT,omitempty"`
}

// ////
// Pump
// ////

type PumpStatus struct {
	Battery        *PumpBattery     `json:"battery,omitempty"`
	Clock          *DateString      `json:"clock,omitempty"`
	Extended       *PumpExtended    `json:"extended,omitempty"`
	Reservoir      *FlexibleFloat64 `json:"reservoir,omitempty"`
	Status         *PumpStatusInfo  `json:"status,omitempty"`
	Bolusing       *FlexibleBool    `json:"bolusing,omitempty"`
	SecondsFromGMT *FlexibleInt64   `json:"secondsFromGMT,omitempty"`
	Manufacturer   *ShortString     `json:"manufacturer,omitempty"`
	PumpID         *ShortString     `json:"pumpID,omitempty"`
	Model          *ShortString     `json:"model,omitempty"`
	Suspended      *FlexibleBool    `json:"suspended,omitempty"`
}

type PumpBattery struct {
	Percent *FlexibleFloat64 `json:"percent,omitempty"`
	Voltage *FlexibleFloat64 `json:"voltage,omitempty"`
}

type PumpExtended struct {
	Version               *ShortString     `json:"Version,omitempty"`
	ActiveProfile         *ShortString     `json:"ActiveProfile,omitempty"`
	TempBasalAbsoluteRate *FlexibleFloat64 `json:"TempBasalAbsoluteRate,omitempty"`
	TempBasalStart        *ShortString     `json:"TempBasalStart,omitempty"`
	TempBasalRemaining    *FlexibleInt64   `json:"TempBasalRemaining,omitempty"`
	LastBolus             *ShortString     `json:"LastBolus,omitempty"`
	LastBolusAmount       *FlexibleFloat64 `json:"LastBolusAmount,omitempty"`
	BaseBasalRate         *FlexibleFloat64 `json:"BaseBasalRate,omitempty"`
}

type PumpStatusInfo struct {
	Status    *ShortString `json:"status,omitempty"`
	Timestamp *DateString  `json:"timestamp,omitempty"`
}

// ///
// Uploader
// ///
type UploaderStatus struct {
	Battery   *FlexibleInt64 `json:"battery,omitempty"`
	Name      *ShortString   `json:"name,omitempty"`
	Timestamp *DateString    `json:"timestamp,omitempty"`
}
