package nightscout

import "encoding/json"

type DeviceStatus struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Device     *SanitizedString  `json:"device,omitempty"`

	// Uploader status
	Uploader *UploaderStatus `json:"uploader,omitempty"`

	// Loop status
	Loop *LoopStatus `json:"loop,omitempty"`

	// Pump status
	Pump *PumpStatus `json:"pump,omitempty"`

	// OpenAPS specific
	OpenAPS *LoopStatus `json:"openaps,omitempty"`

	//
	UploaderBattery *FlexibleFloat64 `json:"uploaderBattery,omitempty"`

	// Date Fields
	CreatedAt   *DateString `json:"created_at,omitempty"`
	SrvCreated  *DateInt    `json:"srvCreated,omitempty"`
	SrvModified *DateInt    `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`
}

func (t *DeviceStatus) UnmarshalJSON(data []byte) error {
	type Alias DeviceStatus
	aux := struct {
		ID IdentifierString `json:"_id"`
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
		t.Identifier = &aux.ID
	}

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
	if t.CreatedAt == nil {
		// This is UTC not User timezone, will have to set UTC offset minutes at API level
		ds, now := DateStringNow()
		t.CreatedAt = ds
		t.Date = ptr(DateIntFromTime(now))
	} else {
		t.Date, t.UtcOffsetMinutes = t.CreatedAt.MillisUTCOffset()
	}

	return nil
}

type UploaderStatus struct {
	Battery        *FlexibleFloat64 `json:"battery,omitempty"`
	BatteryVoltage *FlexibleFloat64 `json:"batteryVoltage,omitempty"`

	Name      *SanitizedString `json:"name,omitempty"`
	Timestamp *DateString      `json:"timestamp,omitempty"`
}

type LoopStatus struct {
	Timestamp *DateString      `json:"timestamp,omitempty"`
	Name      *SanitizedString `json:"name,omitempty"`
	Version   *SanitizedString `json:"version,omitempty"`

	IOB *IOBStatus `json:"iob,omitempty"`
	COB *COBStatus `json:"cob,omitempty"`

	Predicted map[string]any `json:"predicted,omitempty"`

	Enacted   map[string]any `json:"enacted,omitempty"`
	Suggested map[string]any `json:"suggested,omitempty"`

	AutomaticDoseRecommendation map[string]any `json:"automaticDoseRecommendation,omitempty"`

	RecommendedBolus *FlexibleFloat64 `json:"recommendedBolus,omitempty"`

	FailureReason *SanitizedString `json:"failureReason,omitempty"`
}

type PumpStatus struct {
	Clock     *DateString      `json:"clock,omitempty"`
	Reservoir *FlexibleFloat64 `json:"reservoir,omitempty"`

	Battery map[string]any `json:"battery,omitempty"`
	Status  map[string]any `json:"status,omitempty"`

	Extended map[string]any `json:"extended,omitempty"`

	Bolusing       *FlexibleBool    `json:"bolusing,omitempty"`
	Suspended      *FlexibleBool    `json:"suspended,omitempty"`
	Manufacturer   *SanitizedString `json:"manufacturer,omitempty"`
	Model          *SanitizedString `json:"model,omitempty"`
	PumpID         *SanitizedString `json:"pumpID,omitempty"`
	SecondsFromGMT *FlexibleInt64   `json:"secondsFromGMT,omitempty"`
}

type IOBStatus struct {
	IOB       *FlexibleFloat64 `json:"iob,omitempty"`
	BasalIOB  *FlexibleFloat64 `json:"basaliob,omitempty"`
	Activity  *FlexibleFloat64 `json:"activity,omitempty"`
	Time      *DateString      `json:"time,omitempty"`
	Timestamp *DateString      `json:"timestamp,omitempty"`
}

type COBStatus struct {
	COB       *FlexibleFloat64 `json:"cob,omitempty"`
	Timestamp *DateString      `json:"timestamp,omitempty"`
}
