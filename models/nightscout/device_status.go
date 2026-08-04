package nightscout

import "encoding/json"

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
	MergedLoopStatus *LoopStatus    `json:"-"`
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

	if t.Loop != nil && t.OpenAPS != nil {
		t.MergedLoopStatus = t.OpenAPS
		if t.Loop.Timestamp != nil {
			t.MergedLoopStatus.Timestamp = t.Loop.Timestamp
		}
		if t.Loop.Name != nil {
			t.MergedLoopStatus.Name = t.Loop.Name
		}
		if t.Loop.Version != nil {
			t.MergedLoopStatus.Version = t.Loop.Version
		}
		if t.Loop.Predicted != nil {
			t.MergedLoopStatus.Predicted = t.Loop.Predicted
		}
		if t.Loop.Enacted != nil {
			t.MergedLoopStatus.Enacted = t.Loop.Enacted
		}
		if t.Loop.Suggested != nil {
			t.MergedLoopStatus.Suggested = t.Loop.Suggested
		}
		if t.Loop.AutomaticDoseRecommendation != nil {
			t.MergedLoopStatus.AutomaticDoseRecommendation = t.Loop.AutomaticDoseRecommendation
		}
		if t.Loop.RecommendedBolus != nil {
			t.MergedLoopStatus.RecommendedBolus = t.Loop.RecommendedBolus
		}
		if t.Loop.FailureReason != nil {
			t.MergedLoopStatus.FailureReason = t.Loop.FailureReason
		}

		if t.Loop.IOB != nil {
			// not nil merge
			if t.MergedLoopStatus.IOB == nil {
				t.MergedLoopStatus.IOB = t.Loop.IOB
			} else {
				if t.Loop.IOB.IOB != nil {
					t.MergedLoopStatus.IOB.IOB = t.Loop.IOB.IOB
				}
				if t.Loop.IOB.BasalIOB != nil {
					t.MergedLoopStatus.IOB.BasalIOB = t.Loop.IOB.BasalIOB
				}
				if t.Loop.IOB.Activity != nil {
					t.MergedLoopStatus.IOB.Activity = t.Loop.IOB.Activity
				}
				if t.Loop.IOB.Time != nil {
					t.MergedLoopStatus.IOB.Time = t.Loop.IOB.Time
				}
				if t.Loop.IOB.Timestamp != nil {
					t.MergedLoopStatus.IOB.Timestamp = t.Loop.IOB.Timestamp
				}
			}

		}
		if t.Loop.COB != nil {
			if t.MergedLoopStatus.COB == nil {
				t.MergedLoopStatus.COB = t.Loop.COB
			} else {
				// not nil merge
				if t.Loop.COB.COB != nil {
					t.MergedLoopStatus.COB.COB = t.Loop.COB.COB
				}
				if t.Loop.COB.Timestamp != nil {
					t.MergedLoopStatus.COB.Timestamp = t.Loop.COB.Timestamp
				}
			}
		}

	} else if t.Loop != nil {
		t.MergedLoopStatus = t.Loop
	} else if t.OpenAPS != nil {
		t.MergedLoopStatus = t.OpenAPS
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
