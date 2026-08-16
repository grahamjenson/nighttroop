package nightscout

import (
	"encoding/json"
	"fmt"
)

type DeviceStatus struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Device     *ShortString      `json:"device,omitempty"`

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
	CreatedAt *DateString `json:"created_at,omitempty"`

	// Default String Fields
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`

	// Data Extraction
	LoopIob          *FlexibleFloat64 `json:"-"`
	LoopIobBasal     *FlexibleFloat64 `json:"-"`
	LoopIobTreatment *FlexibleFloat64 `json:"-"`
	LoopCob          *FlexibleFloat64 `json:"-"`
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
		t.Date, t.UtcOffsetMinutes = t.CreatedAt.MillisUTCOffset()
	}

	fmt.Println("FDSJfjiodsjfiodsfiodsjiofjiodsjfio")
	fmt.Println(t.Loop, t.OpenAPS)

	fmt.Println("FDSJfjiodsjfiodsfiodsjiofjiodsjfio")
	if t.Loop != nil {
		if t.Loop.IOB != nil {
			if t.Loop.IOB.IOB != nil {
				t.LoopIob = t.Loop.IOB.IOB
			}
			if t.Loop.IOB.BasalIOB != nil {
				t.LoopIobBasal = t.Loop.IOB.BasalIOB
			}

		}
		if t.Loop.COB != nil {
			if t.Loop.COB.COB != nil {
				t.LoopCob = t.Loop.COB.COB
			}
		}
	}

	if t.OpenAPS != nil {
		if t.OpenAPS.IOB != nil {
			if t.OpenAPS.IOB.IOB != nil {
				t.LoopIob = t.OpenAPS.IOB.IOB
			}
			if t.OpenAPS.IOB.BasalIOB != nil {
				t.LoopIobBasal = t.OpenAPS.IOB.BasalIOB
			}

		}
		if t.OpenAPS.COB != nil {
			if t.OpenAPS.COB.COB != nil {
				t.LoopCob = t.OpenAPS.COB.COB
			}
		}
	}

	if t.LoopIobTreatment == nil && t.LoopIobBasal != nil && t.LoopIob != nil {
		t.LoopIobTreatment = ptr(FlexibleFloat64(*t.LoopIob - *t.LoopIobBasal))
	}

	return nil
}

type UploaderStatus struct {
	Battery        *FlexibleFloat64 `json:"battery,omitempty"`
	BatteryVoltage *FlexibleFloat64 `json:"batteryVoltage,omitempty"`

	Name      *ShortString `json:"name,omitempty"`
	Timestamp *DateString  `json:"timestamp,omitempty"`
}

type LoopStatus struct {
	Timestamp *DateString  `json:"timestamp,omitempty"`
	Name      *ShortString `json:"name,omitempty"`
	Version   *ShortString `json:"version,omitempty"`

	IOB *IOBStatus `json:"iob,omitempty"`
	COB *COBStatus `json:"cob,omitempty"`

	Predicted map[string]any `json:"predicted,omitempty"`

	Enacted   map[string]any `json:"enacted,omitempty"`
	Suggested map[string]any `json:"suggested,omitempty"`

	AutomaticDoseRecommendation map[string]any `json:"automaticDoseRecommendation,omitempty"`

	RecommendedBolus *FlexibleFloat64 `json:"recommendedBolus,omitempty"`

	FailureReason *ShortString `json:"failureReason,omitempty"`
}

type PumpStatus struct {
	Clock     *DateString      `json:"clock,omitempty"`
	Reservoir *FlexibleFloat64 `json:"reservoir,omitempty"`

	Battery map[string]any `json:"battery,omitempty"`
	Status  map[string]any `json:"status,omitempty"`

	Extended map[string]any `json:"extended,omitempty"`

	Bolusing       *FlexibleBool  `json:"bolusing,omitempty"`
	Suspended      *FlexibleBool  `json:"suspended,omitempty"`
	Manufacturer   *ShortString   `json:"manufacturer,omitempty"`
	Model          *ShortString   `json:"model,omitempty"`
	PumpID         *ShortString   `json:"pumpID,omitempty"`
	SecondsFromGMT *FlexibleInt64 `json:"secondsFromGMT,omitempty"`
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
