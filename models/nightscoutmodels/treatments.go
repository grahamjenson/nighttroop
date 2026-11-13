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
	GlucoseType *GlucoseType           `json:"glucoseType,omitempty"` // "Sensor", "Finger", "Manual", or as sent

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

	// Duration is read from and written as duration and durationInMilliseconds
	Duration     *types.Duration    `json:"-"`
	DurationType *types.ShortString `json:"durationType,omitempty"` // Loop's "indefinite" override

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

	// PumpID is AAPS's id for a pump's record, which it finds its own by
	PumpID     *types.ClientID    `json:"pumpId,omitempty"`
	PumpSerial *types.ShortString `json:"pumpSerial,omitempty"`
	PumpType   *types.ShortString `json:"pumpType,omitempty"`

	Notes     *types.SanitizedString `json:"notes,omitempty"`
	EnteredBy *types.ShortString     `json:"enteredBy,omitempty"`
	Reason    *types.ShortString     `json:"reason,omitempty"`
	// UUID is xDrip+'s id, which it matches its own treatments by
	UUID *types.ClientID `json:"uuid,omitempty"`
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

	// Date is when it happened, read from and written as its TimeAliases
	Date             *types.DateInt       `json:"-"`
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"` // Minutes offset from UTC

	// Given back for Loop's NightscoutKit, which drops a bolus without programmed and a temp basal without temp
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
	date, offset, err := treatmentTimes.Read(data)
	if err != nil {
		return err
	}
	type Alias Treatment
	// We here extract some fields used by other models as a unique identfier
	aux := struct {
		LegacyID *types.IdentifierString `json:"_id"`
		// Trio's own id, which it deletes by; one that is not an identifier is left out rather than refuse the upload
		ID             json.RawMessage `json:"id"`
		SyncIdentifier json.RawMessage `json:"syncIdentifier"`
		// Shadows EventType, to tell one sent as "<none>"
		EventType json.RawMessage `json:"eventType"`
		types.SentDuration
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.SentJSON = string(data)

	if len(aux.EventType) > 0 && string(aux.EventType) != "null" {
		var eventType TreatmentEventType
		if err := json.Unmarshal(aux.EventType, &eventType); err != nil {
			return err
		}
		var sent string
		if json.Unmarshal(aux.EventType, &sent) == nil && strings.EqualFold(strings.TrimSpace(sent), "<none>") {
			eventType = t.noEventType()
		}
		t.EventType = &eventType
	}

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

	t.Duration = aux.SentDuration.Duration()

	t.Date, t.UtcOffsetMinutes = date, offset
	return nil
}

// treatmentTimes are when a treatment happened: eventTime is its created_at, as Nightscout stores a careportal's or
// xDrip4iOS's; a timestamp, a string or ms, is when it was with no other time, and xDrip+'s ms one gives its
// created_at of whole seconds its milliseconds. It is given back on API v1 for Loop's NightscoutKit, which drops a
// treatment without one, in UTC, as LoopFollow reads it before created_at and drops its offset, but not on API v3,
// as AAPS drops a treatment whose timestamp is not a number.
var treatmentTimes = types.TimeAliases{{Name: "date", Millis: true}, {Name: "eventTime", Write: types.WriteNever},
	{Name: "created_at"}, {Name: "timestamp", Write: types.WriteV1}}

func (t Treatment) TimeAliases() types.TimeAliases {
	return treatmentTimes
}

// noEventType is the event type of a treatment sent with none, "<none>", as xDrip+ sends every bolus, carbs and note
// and Juggluco its amounts: one with insulin and carbs is a Meal Bolus, one with carbs a Carb Correction and one with
// insulin a Correction Bolus, as LoopFollow and NightscoutKit only read amounts of those, and one with neither a Note
func (t *Treatment) noEventType() TreatmentEventType {
	insulin, carbs := t.Insulin != nil && *t.Insulin > 0, t.Carbs != nil && *t.Carbs > 0
	switch {
	case insulin && carbs:
		return MealBolus
	case carbs:
		return CarbCorrection
	case insulin:
		return CorrectionBolus
	}
	return Note
}

// SplitCarbs is a treatment with both insulin and carbs as two, as AAPS reads only the insulin of one with both
// (RemoteTreatment.toTreatment makes one record of a document, insulin first): the treatment without its carbs, and its
// carbs as a Carb Correction then. ok is false for one without both.
func (t Treatment) SplitCarbs() (insulin, carbs Treatment, ok bool) {
	if t.Insulin == nil || *t.Insulin <= 0 || t.Carbs == nil || *t.Carbs <= 0 {
		return t, Treatment{}, false
	}
	carbs = Treatment{EventType: new(CarbCorrection), Carbs: t.Carbs, Protein: t.Protein, Fat: t.Fat, FoodType: t.FoodType,
		AbsorptionTimeMinutes: t.AbsorptionTimeMinutes, EnteredBy: t.EnteredBy, App: t.App, Date: t.Date,
		UtcOffsetMinutes: t.UtcOffsetMinutes}
	insulin = t
	insulin.Carbs, insulin.Protein, insulin.Fat, insulin.FoodType, insulin.AbsorptionTimeMinutes = nil, nil, nil, nil, nil
	return insulin, carbs, true
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
	// gives every bolus, carbs and note; a treatment sent so takes its type from its amounts (noEventType), and is
	// otherwise a note
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
	// Manual is a finger stick typed in, as AAPS and iAPS send it
	Manual GlucoseType = "Manual"
)

// maxGlucoseTypeLength is the most of a glucose type kept, as its column holds 32 characters
const maxGlucoseTypeLength = 32

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
	case "manual":
		*t = Manual
	default:
		// Another type is kept as sent, as Nightscout does, rather than refuse the treatment
		if len(s) > maxGlucoseTypeLength {
			s = strings.ToValidUTF8(s[:maxGlucoseTypeLength], "")
		}
		*t = GlucoseType(s)
	}

	return nil
}

// MarshalJSON writes its duration as its fields
func (t Treatment) MarshalJSON() ([]byte, error) {
	type Alias Treatment
	return json.Marshal(struct {
		Alias
		types.DurationFields
	}{Alias(t), t.Duration.Fields()})
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

// V3 is the treatment as API v3 and its storage socket write it: its duration in whole minutes, as AAPS reads it as a
// whole number (RemoteTreatment.duration) and drops a treatment, or a page of them, whose duration has a fraction. Its
// durationInMilliseconds, which AAPS reads first, is exact; API v1 gives Loop the minutes with their fraction.
//
// A profile switch's timeshift is in milliseconds too, as AAPS reads it so: one AAPS sent, with its
// originalProfileName, is already, and another, from the site or the careportal, is kept in minutes.
func (t Treatment) V3() any {
	t.Duration = t.Duration.InWholeMinutes()
	if t.Timeshift != nil && t.OriginalProfileName == nil {
		t.Timeshift = new(*t.Timeshift * 60_000)
	}
	return t
}

// MergeAliases are the fields a merge takes out of the stored treatment when one is sent, as they say the same: a
// duration sent is no longer indefinite, its targets are its correction range, and a temp basal's rate is its percent
func (t Treatment) MergeAliases() map[string][]string {
	return map[string][]string{
		"duration":               {"durationInMilliseconds", "durationType"},
		"durationInMilliseconds": {"duration", "durationType"},
		"targetTop":              {"correctionRange"},
		"targetBottom":           {"correctionRange"},
		"correctionRange":        {"targetTop", "targetBottom"},
		"absolute":               {"rate", "percent"},
		"rate":                   {"absolute", "percent"},
		"percent":                {"absolute", "rate"},
	}
}

// GlucoseFields are the fields whose glucose is in the treatment's units
func (t Treatment) GlucoseFields() []string {
	return []string{"glucose", "targetTop", "targetBottom", "correctionRange"}
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
