package nightscout

type DeviceStatus struct {
	Identifier *string     `json:"identifier,omitempty"`
	Device     *string     `json:"device,omitempty"`
	CreatedAt  *DateString `json:"created_at,omitempty"`

	// Uploader status
	Uploader *UploaderStatus `json:"uploader,omitempty"`

	// Loop status
	Loop *LoopStatus `json:"loop,omitempty"`

	// Pump status
	Pump *PumpStatus `json:"pump,omitempty"`

	// OpenAPS specific
	OpenAPS *LoopStatus `json:"openaps,omitempty"`

	// API v3 additions

	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	RawJSON string `json:"-"`
}

func (e *DeviceStatus) FromMap(m map[string]any) error {
	return FromMap(m, e)
}

func (e *DeviceStatus) ToMap() (map[string]any, error) {
	return ToMap(e)
}

type UploaderStatus struct {
	Battery        *FlexibleFloat64 `json:"battery,omitempty"`
	BatteryVoltage *FlexibleFloat64 `json:"batteryVoltage,omitempty"`

	Name      *string     `json:"name,omitempty"`
	Timestamp *DateString `json:"timestamp,omitempty"`
}

type LoopStatus struct {
	Timestamp *DateString `json:"timestamp,omitempty"`
	Name      *string     `json:"name,omitempty"`
	Version   *string     `json:"version,omitempty"`

	IOB *IOBStatus `json:"iob,omitempty"`
	COB *COBStatus `json:"cob,omitempty"`

	Predicted map[string]any `json:"predicted,omitempty"`

	Enacted   map[string]any `json:"enacted,omitempty"`
	Suggested map[string]any `json:"suggested,omitempty"`

	AutomaticDoseRecommendation map[string]any `json:"automaticDoseRecommendation,omitempty"`

	RecommendedBolus *FlexibleFloat64 `json:"recommendedBolus,omitempty"`

	FailureReason *string `json:"failureReason,omitempty"`
}

type PumpStatus struct {
	Clock     *DateString      `json:"clock,omitempty"`
	Reservoir *FlexibleFloat64 `json:"reservoir,omitempty"`

	Battery map[string]any `json:"battery,omitempty"`
	Status  map[string]any `json:"status,omitempty"`

	Extended map[string]any `json:"extended,omitempty"`

	Bolusing       *bool   `json:"bolusing,omitempty"`
	Suspended      *bool   `json:"suspended,omitempty"`
	Manufacturer   *string `json:"manufacturer,omitempty"`
	Model          *string `json:"model,omitempty"`
	PumpID         *string `json:"pumpID,omitempty"`
	SecondsFromGMT *int    `json:"secondsFromGMT,omitempty"`
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
