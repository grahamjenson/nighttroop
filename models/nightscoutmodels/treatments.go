package nightscoutmodels

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
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
	// additional: a bolus's type, e.g. SMB, or AAPS's temp basal's, e.g. PUMP_SUSPEND
	Type *types.ShortString `json:"type,omitempty"`
	// Automatic is a bolus a loop gave, given back as sent, as LoopFollow shows one as an SMB
	Automatic *bool `json:"automatic,omitempty"`
	// IsBasalInsulin is AAPS's bolus of basal insulin, and ICfg the insulin AAPS says a bolus or profile switch was
	// given with, given back as sent, as AAPS takes a bolus without them as not basal, of the insulin in its settings
	IsBasalInsulin *bool          `json:"isBasalInsulin,omitempty"`
	ICfg           *InsulinConfig `json:"icfg,omitempty"`
	// BolusCalculatorResult is AAPS's Bolus Wizard calculation, JSON text given back as sent, as AAPS restores the
	// record from it
	BolusCalculatorResult *types.JSONTextString `json:"bolusCalculatorResult,omitempty"`

	// Exercise
	ActivityType *types.ShortString   `json:"activityType,omitempty"` // Type of activity
	Steps        *types.FlexibleInt64 `json:"steps,omitempty"`        // Step count
	Calories     *types.FlexibleInt64 `json:"calories,omitempty"`     // Calories burned
	HeartRate    *types.FlexibleInt64 `json:"heartRate,omitempty"`    // Heart rate

	DurationMinutes        *types.FlexibleInt64 `json:"duration,omitempty"`               // minutes
	DurationInMilliseconds *types.FlexibleInt64 `json:"durationInMilliseconds,omitempty"` // milliseconds
	DurationType           *types.ShortString   `json:"durationType,omitempty"`           // Loop's "indefinite" override

	// Treatment-specific fields
	Profile             *types.ShortString     `json:"profile,omitempty"`
	OriginalProfileName *types.ShortString     `json:"originalProfileName,omitempty"`
	ProfileJSON         *types.LongJSONString  `json:"profileJson,omitempty"`
	Percent             *types.FlexibleFloat64 `json:"percentage,omitempty"`
	Timeshift           *types.FlexibleInt64   `json:"timeshift,omitempty"`
	// The AAPS switch a profile switch, or an effective one (a Note), was made from; its times are in milliseconds
	OriginalCustomizedName *types.ShortString   `json:"originalCustomizedName,omitempty"`
	OriginalTimeshiftMs    *types.FlexibleInt64 `json:"originalTimeshift,omitempty"`
	OriginalPercentage     *types.FlexibleInt64 `json:"originalPercentage,omitempty"`
	OriginalDurationMs     *types.FlexibleInt64 `json:"originalDuration,omitempty"`
	OriginalEnd            *types.FlexibleInt64 `json:"originalEnd,omitempty"`

	// AAPS's running mode, e.g. DISCONNECTED_PUMP, whether the loop forced it and why; how long it was to last, in ms,
	// is OriginalDurationMs above, which a profile switch has too
	Mode       *types.ShortString     `json:"mode,omitempty"`
	AutoForced *bool                  `json:"autoForced,omitempty"`
	Reasons    *types.SanitizedString `json:"reasons,omitempty"`

	Rate     *types.FlexibleFloat64 `json:"rate,omitempty"`
	Absolute *types.FlexibleFloat64 `json:"absolute,omitempty"`
	// TempBasalPercent is a temp basal's change from the basal in percent, as Nightscout has it, so -50 is half of it and
	// -100 a zero temp, where a profile switch's is percentage
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
	// UUID is given back as sent, as xDrip+ matches its own treatments by it
	UUID *types.ShortString `json:"uuid,omitempty"`
	// Loop's id for a carb entry or dose, which it finds its own by, given back as sent; the identifier is made from it
	SyncIdentifier *string `json:"syncIdentifier,omitempty"`
	// Loop's insulin of a dose, a bolus's insulin not absorbed, and when a carb entry was entered, given back as sent
	InsulinType   *types.ShortString     `json:"insulinType,omitempty"`
	Unabsorbed    *types.FlexibleFloat64 `json:"unabsorbed,omitempty"`
	UserEnteredAt *types.DateString      `json:"userEnteredAt,omitempty"`

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Date Fields
	CreatedAt        *types.DateString    `json:"created_at,omitempty"` // ISO8601 -> time.Time
	Date             *types.DateInt       `json:"date,omitempty"`       // Unix timestamp in milliseconds
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"`  // Minutes offset from UTC

	// Given back for Loop's NightscoutKit, which drops a treatment without a timestamp, a bolus without programmed
	// and a temp basal without temp; a timestamp sent is not read, as created_at is when it was
	Timestamp  *types.DateString      `json:"timestamp,omitempty"`
	Programmed *types.FlexibleFloat64 `json:"programmed,omitempty"`
	Temp       *types.ShortString     `json:"temp,omitempty"`

	// SentJSON is the JSON it was sent as, to tell one sent again with changes from one sent the same
	SentJSON string `json:"-"`
}

// MaxSyncIdentifierLength is the longest syncIdentifier given back, as an Omnipod dose's is 60 to 70 characters
const MaxSyncIdentifierLength = 256

// InsulinConfig is AAPS's insulin, its times in milliseconds
type InsulinConfig struct {
	InsulinLabel    *types.ShortString     `json:"insulinLabel,omitempty"`
	InsulinEndTime  *types.FlexibleInt64   `json:"insulinEndTime,omitempty"`
	InsulinPeakTime *types.FlexibleInt64   `json:"insulinPeakTime,omitempty"`
	Concentration   *types.FlexibleFloat64 `json:"concentration,omitempty"`
}

func (t *Treatment) UnmarshalJSON(data []byte) error {
	type Alias Treatment
	// We here extract some fields used by other models as a unique identfier
	aux := struct {
		LegacyID *types.IdentifierString `json:"_id"`
		// Trio's own id, which it deletes by; one that is not an identifier is left out rather than refuse the upload
		ID             json.RawMessage `json:"id"`
		SyncIdentifier json.RawMessage `json:"syncIdentifier"`
		// Nightscout stores a careportal's or xDrip4iOS's eventTime as created_at
		EventTime *types.DateString `json:"eventTime"`
		// Shadows Timestamp, so a timestamp sent is not read
		Timestamp json.RawMessage `json:"timestamp"`
		// Shadows DurationMinutes, as Loop sends minutes with a fraction, e.g. a bolus's 1.6666666666666667
		Duration *types.FlexibleFloat64 `json:"duration"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.SentJSON = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = aux.LegacyID
	}
	if t.Identifier == nil {
		t.Identifier = types.IdentifierFromJSON(aux.ID)
	}
	if t.Identifier == nil {
		t.Identifier = types.IdentifierFromJSON(aux.SyncIdentifier)
	}
	var syncIdentifier string
	if json.Unmarshal(aux.SyncIdentifier, &syncIdentifier) == nil && syncIdentifier != "" && len(syncIdentifier) <= MaxSyncIdentifierLength {
		t.SyncIdentifier = &syncIdentifier
	}
	if t.Identifier == nil && t.UUID != nil {
		t.Identifier = types.ToIdentifierString(string(*t.UUID))
	}

	// A duration of minutes with a fraction is kept in whole minutes, as its events keep them, and exactly in
	// milliseconds, when none were sent, so it is given back as sent
	if aux.Duration != nil {
		minutes := aux.Duration.Float64()
		t.DurationMinutes = new(types.FlexibleInt64(int64(minutes)))
		if t.DurationInMilliseconds == nil && minutes != math.Trunc(minutes) {
			t.DurationInMilliseconds = new(types.FlexibleInt64(math.Round(minutes * 60_000)))
		}
	}

	// A duration sent only in milliseconds is in minutes too, the nearest, as its events keep minutes
	if t.DurationMinutes == nil && t.DurationInMilliseconds != nil {
		t.DurationMinutes = new(types.FlexibleInt64((*t.DurationInMilliseconds + 30_000) / 60_000))
	}

	if aux.EventTime != nil {
		t.CreatedAt = aux.EventTime
	}

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
	if t.CreatedAt != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			date, offset := t.CreatedAt.MillisSentOffset()
			t.Date, t.UtcOffsetMinutes = date, cmp.Or(t.UtcOffsetMinutes, offset)
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
	// Trio and iAPS send an automatic bolus as SMB, and insulin not from the pump as External Insulin
	SMB             TreatmentEventType = "SMB"
	ExternalInsulin TreatmentEventType = "External Insulin"

	// Activity
	Exercise TreatmentEventType = "Exercise"

	// Algorithms
	ProfileSwitch   TreatmentEventType = "Profile Switch"
	TempBasal       TreatmentEventType = "Temp Basal"
	TemporaryTarget TreatmentEventType = "Temporary Target"
	// TemporaryOverride is Loop's override, kept as a temp target with its insulin needs scale factor
	TemporaryOverride TreatmentEventType = "Temporary Override"
	// Loop's and Trio's pump suspends and resumes
	SuspendPump TreatmentEventType = "Suspend Pump"
	ResumePump  TreatmentEventType = "Resume Pump"

	// Device Events
	OpenAPSOffline    TreatmentEventType = "OpenAPS Offline"
	PumpBatteryChange TreatmentEventType = "Pump Battery Change"
	SensorChange      TreatmentEventType = "Sensor Change"
	SensorStart       TreatmentEventType = "Sensor Start"
	SensorStop        TreatmentEventType = "Sensor Stop"
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
	SensorStop,
	SiteChange,
	SnackBolus,
	SMB,
	ExternalInsulin,
	TempBasal,
	TemporaryTarget,
	TemporaryOverride,
	SuspendPump,
	ResumePump,
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

	// xDrip4iOS sends carbs on their own as "Carbs"
	if strings.TrimSpace(s) == "Carbs" {
		*t = CarbCorrection
		return nil
	}
	for _, et := range TreatmentEventTypes {
		if strings.TrimSpace(s) == string(et) {
			*t = et
			return nil
		}
	}
	// "<none>" is Nightscout's careportal's treatment with no event type, which Juggluco sends its amounts as and xDrip+
	// gives every bolus, carbs and note, so it is a note keeping any insulin and carbs
	if strings.EqualFold(strings.TrimSpace(s), "<none>") {
		*t = Note
		return nil
	}
	// AAPS records exporting its settings as a "SettingsExport", which has nothing to dose, so it is a note
	if strings.TrimSpace(s) == "SettingsExport" {
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

	// xDrip4iOS adds the reading after a colon, e.g. "Finger: 6.7 mmol/L"
	kind, _, _ := strings.Cut(s, ":")
	switch strings.ToLower(strings.TrimSpace(kind)) {
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

// MarshalJSON writes a duration kept in milliseconds that are not whole minutes as the minutes they are, with their
// fraction, while its minutes are theirs, as Loop's NightscoutKit and LoopFollow read the minutes
func (t Treatment) MarshalJSON() ([]byte, error) {
	type Alias Treatment
	out := struct {
		Alias
		Duration any `json:"duration,omitempty"`
	}{Alias: Alias(t)}
	if t.DurationMinutes != nil {
		out.Duration = t.DurationMinutes
	}
	if ms, minutes := t.DurationInMilliseconds, t.DurationMinutes; ms != nil && minutes != nil && *ms%60_000 != 0 && int64(*ms)/60_000 == int64(*minutes) {
		out.Duration = float64(*ms) / 60_000
	}
	return json.Marshal(out)
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

// V1OnlyFields are given back only through API v1 and its sockets, not API v3 and its storage socket: Loop's
// NightscoutKit reads v1 and drops a treatment without a timestamp, but AAPS reads v3 and drops one whose timestamp is
// not a number
func (t Treatment) V1OnlyFields() []string {
	return []string{"timestamp"}
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
