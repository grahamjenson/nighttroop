package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/types"
)

////
// Models
////

type Treatment struct {
	Identifier *types.IdentifierString `json:"identifier,omitempty"`

	EventType *TreatmentEventType `json:"eventType,omitempty"`

	Name *types.ShortString `json:"name,omitempty"` // Name of the activity

	// Glucose
	Glucose     *types.FlexibleFloat64 `json:"glucose,omitempty"`
	GlucoseType *GlucoseType           `json:"glucoseType,omitempty"` // "Sensor" or "Finger"

	// Nutrition
	Carbs                 *types.FlexibleFloat64 `json:"carbs,omitempty"`
	Protein               *types.FlexibleFloat64 `json:"protein,omitempty"`
	Fat                   *types.FlexibleFloat64 `json:"fat,omitempty"`
	FoodType              *types.ShortString     `json:"foodType,omitempty"` // Type of food
	AbsorptionTimeMinutes *types.FlexibleInt64   `json:"absorptionTime,omitempty"`
	Category              *types.ShortString     `json:"category,omitempty"`
	Subcategory           *types.ShortString     `json:"subcategory,omitempty"`
	PreBolus              *types.FlexibleFloat64 `json:"preBolus,omitempty"`
	// additional
	Fibre    *types.FlexibleFloat64 `json:"fibre,omitempty"`
	Energy   *types.FlexibleFloat64 `json:"energy,omitempty"`
	Portions *types.FlexibleFloat64 `json:"portions,omitempty"`

	// Insulin
	Insulin  *types.FlexibleFloat64 `json:"insulin,omitempty"`
	SplitNow *types.FlexibleFloat64 `json:"splitNow,omitempty"`
	SplitExt *types.FlexibleFloat64 `json:"splitExt,omitempty"`
	// additional
	Type *types.ShortString `json:"type,omitempty"`

	// Exercise
	ActivityType *types.ShortString   `json:"activityType,omitempty"` // Type of activity
	Steps        *types.FlexibleInt64 `json:"steps,omitempty"`        // Step count
	Calories     *types.FlexibleInt64 `json:"calories,omitempty"`     // Calories burned
	HeartRate    *types.FlexibleInt64 `json:"heartRate,omitempty"`    // Heart rate

	DurationMinutes        *types.FlexibleInt64 `json:"duration,omitempty"`               // minutes
	DurationInMilliseconds *types.FlexibleInt64 `json:"durationInMilliseconds,omitempty"` // minutes

	// Treatment-specific fields
	Profile             *types.ShortString     `json:"profile,omitempty"`
	OriginalProfileName *types.ShortString     `json:"originalProfileName,omitempty"`
	ProfileJSON         *types.LongJSONString  `json:"profileJson,omitempty"`
	Percent             *types.FlexibleFloat64 `json:"percentage,omitempty"`
	Timeshift           *types.FlexibleInt64   `json:"timeshift,omitempty"`

	Rate     *types.FlexibleFloat64 `json:"rate,omitempty"`
	Absolute *types.FlexibleFloat64 `json:"absolute,omitempty"`
	// TempBasalPercent is a temp basal's percent of the basal, so 50 is half of it, where a profile switch's is percentage
	TempBasalPercent *types.FlexibleFloat64 `json:"percent,omitempty"`

	HighTarget              *types.FlexibleFloat64   `json:"targetTop,omitempty"`
	LowTarget               *types.FlexibleFloat64   `json:"targetBottom,omitempty"`
	CorrectionRange         []*types.FlexibleFloat64 `json:"correctionRange,omitempty"` // [low, high]
	InsulinNeedsScaleFactor *types.FlexibleFloat64   `json:"insulinNeedsScaleFactor,omitempty"`

	Units *consts.Units `json:"units,omitempty"`

	SensorCode    *types.ShortString `json:"sensorCode,omitempty"`
	TransmitterID *types.ShortString `json:"transmitterId,omitempty"`

	PumpID     *types.ShortString `json:"pumpId,omitempty"`
	PumpSerial *types.ShortString `json:"pumpSerial,omitempty"`
	PumpType   *types.ShortString `json:"pumpType,omitempty"`

	Notes     *types.SanitizedString `json:"notes,omitempty"`
	EnteredBy *types.ShortString     `json:"enteredBy,omitempty"`
	Reason    *types.ShortString     `json:"reason,omitempty"`

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Date Fields
	CreatedAt        *types.DateString    `json:"created_at,omitempty"` // ISO8601 -> time.Time
	Date             *types.DateInt       `json:"date,omitempty"`       // Unix timestamp in milliseconds
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"`  // Minutes offset from UTC

	// Non-JSON values to make mapping easier
	RawJsonData string `json:"-"`
}

func (t *Treatment) UnmarshalJSON(data []byte) error {
	type Alias Treatment
	// We here extract some fields used by other models as a unique identfier
	aux := struct {
		LegacyID       *types.IdentifierString `json:"_id"`
		SyncIdentifier *types.IdentifierString `json:"syncIdentifier"`
		UUID           *types.IdentifierString `json:"uuid"`
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
		t.Identifier = aux.LegacyID
	}
	if t.Identifier == nil {
		t.Identifier = aux.SyncIdentifier
	}
	if t.Identifier == nil {
		t.Identifier = aux.UUID
	}

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
	if t.CreatedAt != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			t.Date, t.UtcOffsetMinutes = t.CreatedAt.MillisUTCOffset()
		}
	}

	return nil
}

//"eventType": "Correction Bolus",
// "eventType": "Temporary Target",
// "eventType": "Meal Bolus",
// "eventType": "Temp Basal",
// "eventType": "Note",
// "eventType": "Profile Switch",
// "eventType": "Carb Correction",

//////
// Subtypes and Constants
//////

type TreatmentEventType string

const (
	// Records & alerts
	Announcement TreatmentEventType = "Announcement"
	Note         TreatmentEventType = "Note"
	Question     TreatmentEventType = "Question"
	DADAlert     TreatmentEventType = "D.A.D. Alert"

	// Glucose
	BGCheck TreatmentEventType = "BG Check"

	//Carbs
	CarbCorrection TreatmentEventType = "Carb Correction"

	// Boluses
	Bolus           TreatmentEventType = "Bolus"
	BolusWizard     TreatmentEventType = "Bolus Wizard"
	ComboBolus      TreatmentEventType = "Combo Bolus"
	CorrectionBolus TreatmentEventType = "Correction Bolus"
	MealBolus       TreatmentEventType = "Meal Bolus"
	SnackBolus      TreatmentEventType = "Snack Bolus"

	// Activity
	Exercise TreatmentEventType = "Exercise"

	// Algorithms
	ProfileSwitch   TreatmentEventType = "Profile Switch"
	TempBasal       TreatmentEventType = "Temp Basal"
	TemporaryTarget TreatmentEventType = "Temporary Target"

	// Device Events
	OpenAPSOffline    TreatmentEventType = "OpenAPS Offline"
	PumpBatteryChange TreatmentEventType = "Pump Battery Change"
	SensorChange      TreatmentEventType = "Sensor Change"
	SensorStart       TreatmentEventType = "Sensor Start"
	SiteChange        TreatmentEventType = "Site Change"
	InsulinChange     TreatmentEventType = "Insulin Change"
	CannulaFill       TreatmentEventType = "Cannula Fill"
	Prime             TreatmentEventType = "Prime"
	ReservoirChange   TreatmentEventType = "Reservoir Change"
	Calibration       TreatmentEventType = "Calibration"

	//
	Unknown TreatmentEventType = "Unknown"
)

func (t TreatmentEventType) String() string {
	return string(t)
}

func (t *TreatmentEventType) Stringp() *string {
	if t == nil {
		return nil
	}
	s := string(*t)
	return &s
}

func ToSnakeCase(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), "_"))
}

var TreatmentEventTypes = []TreatmentEventType{
	Announcement,
	BGCheck,
	Bolus,
	BolusWizard,
	CarbCorrection,
	ComboBolus,
	CorrectionBolus,
	DADAlert,
	Exercise,
	InsulinChange,
	MealBolus,
	Note,
	OpenAPSOffline,
	ProfileSwitch,
	PumpBatteryChange,
	Question,
	SensorChange,
	SensorStart,
	SiteChange,
	SnackBolus,
	TempBasal,
	TemporaryTarget,
	CannulaFill,
	Prime,
	ReservoirChange,
	Calibration,
}

func (t *TreatmentEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, et := range TreatmentEventTypes {
		if strings.TrimSpace(s) == string(et) {
			*t = et
			return nil
		}
	}
	// "<none>" is Nightscout's careportal's treatment with no event type, which Juggluco sends its amounts as, so it
	// is a note keeping any insulin and carbs
	if strings.EqualFold(strings.TrimSpace(s), "<none>") {
		*t = Note
		return nil
	}
	return fmt.Errorf("unknown treatment event type: %s", s)
}

type GlucoseType string

const (
	Sensor GlucoseType = "Sensor"
	Finger GlucoseType = "Finger"
	Meter  GlucoseType = "Meter"
)

func (t *GlucoseType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sensor":
		*t = Sensor
	case "finger":
		*t = Finger
	case "meter":
		*t = Meter
	default:
		return fmt.Errorf("unknown glucose type: %s", s)
	}

	return nil
}

func (t Treatment) GetDevice() *string {
	return t.Device.Stringp()
}

func (t Treatment) GetApp() *string {
	return t.App.Stringp()
}

func (t Treatment) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t Treatment) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t Treatment) GetUtcOffsetMinutes() *int64 {
	return t.UtcOffsetMinutes.Int64p()
}

func (t Treatment) GetIdentifier() *string {
	return t.Identifier.Stringp()
}
