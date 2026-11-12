package nightscoutmodels

import (
	"bytes"
	"cmp"
	"encoding/json"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/types"
)

type ProfileStore struct {
	Identifier     *types.IdentifierString       `json:"identifier,omitempty"`
	DefaultProfile *types.ShortString            `json:"defaultProfile,omitempty"`
	Store          map[types.ShortString]Profile `json:"store,omitempty"`
	Units          *consts.Units                 `json:"units,omitempty"`

	LoopSettings *LoopSettings `json:"loopSettings,omitempty"`
	// Trio's remote control settings are at the top of its profile
	*RemoteSettings

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity
	// EnteredBy is who uploaded it, as for treatments; Trio, iAPS and Loop need it to import or read a profile
	EnteredBy *types.ShortString `json:"enteredBy,omitempty"`

	// Date Feilds
	StartDate *types.DateString `json:"startDate,omitempty"`
	Mills     *types.DateInt    `json:"mills,omitempty"`

	// Non-JSON values to make mapping easier
	Date             *types.DateInt       `json:"date,omitempty"`      // Unix timestamp in milliseconds
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"` // Minutes offset from UTC
}

func (t *ProfileStore) UnmarshalJSON(data []byte) error {
	type Alias ProfileStore
	aux := struct {
		ID *types.IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}


	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	if t.StartDate != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			date, offset := t.StartDate.MillisSentOffset()
			t.Date, t.UtcOffsetMinutes = date, cmp.Or(t.UtcOffsetMinutes, offset)
		}
	}

	return nil
}

type Profile struct {
	Units    *consts.Units          `json:"units,omitempty"`
	Dia      *types.FlexibleFloat64 `json:"dia,omitempty"`
	CarbsHr  *types.FlexibleFloat64 `json:"carbs_hr,omitempty"`
	Delay    *types.FlexibleFloat64 `json:"delay,omitempty"`
	Timezone *types.Timezone        `json:"timezone,omitempty"`

	Sens       []*ScheduleValue `json:"sens,omitempty"`
	CarbRatio  []*ScheduleValue `json:"carbratio,omitempty"`
	Basal      []*ScheduleValue `json:"basal,omitempty"`
	TargetLow  []*ScheduleValue `json:"target_low,omitempty"`
	TargetHigh []*ScheduleValue `json:"target_high,omitempty"`
}

type ScheduleValue struct {
	Time          *ScheduleTime          `json:"time,omitempty"`
	TimeAsSeconds *types.FlexibleInt64   `json:"timeAsSeconds,omitempty"`
	Value         *types.FlexibleFloat64 `json:"value,omitempty"`
}

func (t *ScheduleValue) UnmarshalJSON(data []byte) error {
	aux := struct {
		Time          *ScheduleTime        `json:"time,omitempty"`
		TimeAsSeconds *types.FlexibleInt64 `json:"timeAsSeconds,omitempty"`
		Value         json.RawMessage      `json:"value,omitempty"`
	}{
		Time:          t.Time,
		TimeAsSeconds: t.TimeAsSeconds,
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.Time = aux.Time
	t.TimeAsSeconds = aux.TimeAsSeconds

	if bytes.Equal(bytes.TrimSpace(aux.Value), []byte("false")) {
		t.Value = nil
		return nil
	}

	if len(aux.Value) == 0 || bytes.Equal(bytes.TrimSpace(aux.Value), []byte("null")) {
		t.Value = nil
		return nil
	}

	var value types.FlexibleFloat64
	if err := json.Unmarshal(aux.Value, &value); err != nil {
		return err
	}

	t.Value = &value
	return nil
}

type ScheduleTime string

const (
	T0000 ScheduleTime = "00:00"
	T0030 ScheduleTime = "00:30"
	T0100 ScheduleTime = "01:00"
	T0130 ScheduleTime = "01:30"
	T0200 ScheduleTime = "02:00"
	T0230 ScheduleTime = "02:30"
	T0300 ScheduleTime = "03:00"
	T0330 ScheduleTime = "03:30"
	T0400 ScheduleTime = "04:00"
	T0430 ScheduleTime = "04:30"
	T0500 ScheduleTime = "05:00"
	T0530 ScheduleTime = "05:30"
	T0600 ScheduleTime = "06:00"
	T0630 ScheduleTime = "06:30"
	T0700 ScheduleTime = "07:00"
	T0730 ScheduleTime = "07:30"
	T0800 ScheduleTime = "08:00"
	T0830 ScheduleTime = "08:30"
	T0900 ScheduleTime = "09:00"
	T0930 ScheduleTime = "09:30"
	T1000 ScheduleTime = "10:00"
	T1030 ScheduleTime = "10:30"
	T1100 ScheduleTime = "11:00"
	T1130 ScheduleTime = "11:30"
	T1200 ScheduleTime = "12:00"
	T1230 ScheduleTime = "12:30"
	T1300 ScheduleTime = "13:00"
	T1330 ScheduleTime = "13:30"
	T1400 ScheduleTime = "14:00"
	T1430 ScheduleTime = "14:30"
	T1500 ScheduleTime = "15:00"
	T1530 ScheduleTime = "15:30"
	T1600 ScheduleTime = "16:00"
	T1630 ScheduleTime = "16:30"
	T1700 ScheduleTime = "17:00"
	T1730 ScheduleTime = "17:30"
	T1800 ScheduleTime = "18:00"
	T1830 ScheduleTime = "18:30"
	T1900 ScheduleTime = "19:00"
	T1930 ScheduleTime = "19:30"
	T2000 ScheduleTime = "20:00"
	T2030 ScheduleTime = "20:30"
	T2100 ScheduleTime = "21:00"
	T2130 ScheduleTime = "21:30"
	T2200 ScheduleTime = "22:00"
	T2230 ScheduleTime = "22:30"
	T2300 ScheduleTime = "23:00"
	T2330 ScheduleTime = "23:30"
)

func (s *ScheduleTime) ToSeconds() *types.FlexibleInt64 {
	switch *s {
	case T0000:
		return new(types.FlexibleInt64(0))
	case T0030:
		return new(types.FlexibleInt64(1800))
	case T0100:
		return new(types.FlexibleInt64(1800 * 2))
	case T0130:
		return new(types.FlexibleInt64(1800 * 3))
	case T0200:
		return new(types.FlexibleInt64(1800 * 4))
	case T0230:
		return new(types.FlexibleInt64(1800 * 5))
	case T0300:
		return new(types.FlexibleInt64(1800 * 6))
	case T0330:
		return new(types.FlexibleInt64(1800 * 7))
	case T0400:
		return new(types.FlexibleInt64(1800 * 8))
	case T0430:
		return new(types.FlexibleInt64(1800 * 9))
	case T0500:
		return new(types.FlexibleInt64(1800 * 10))
	case T0530:
		return new(types.FlexibleInt64(1800 * 11))
	case T0600:
		return new(types.FlexibleInt64(1800 * 12))
	case T0630:
		return new(types.FlexibleInt64(1800 * 13))
	case T0700:
		return new(types.FlexibleInt64(1800 * 14))
	case T0730:
		return new(types.FlexibleInt64(1800 * 15))
	case T0800:
		return new(types.FlexibleInt64(1800 * 16))
	case T0830:
		return new(types.FlexibleInt64(1800 * 17))
	case T0900:
		return new(types.FlexibleInt64(1800 * 18))
	case T0930:
		return new(types.FlexibleInt64(1800 * 19))
	case T1000:
		return new(types.FlexibleInt64(1800 * 20))
	case T1030:
		return new(types.FlexibleInt64(1800 * 21))
	case T1100:
		return new(types.FlexibleInt64(1800 * 22))
	case T1130:
		return new(types.FlexibleInt64(1800 * 23))
	case T1200:
		return new(types.FlexibleInt64(1800 * 24))
	case T1230:
		return new(types.FlexibleInt64(1800 * 25))
	case T1300:
		return new(types.FlexibleInt64(1800 * 26))
	case T1330:
		return new(types.FlexibleInt64(1800 * 27))
	case T1400:
		return new(types.FlexibleInt64(1800 * 28))
	case T1430:
		return new(types.FlexibleInt64(1800 * 29))
	case T1500:
		return new(types.FlexibleInt64(1800 * 30))
	case T1530:
		return new(types.FlexibleInt64(1800 * 31))
	case T1600:
		return new(types.FlexibleInt64(1800 * 32))
	case T1630:
		return new(types.FlexibleInt64(1800 * 33))
	case T1700:
		return new(types.FlexibleInt64(1800 * 34))
	case T1730:
		return new(types.FlexibleInt64(1800 * 35))
	case T1800:
		return new(types.FlexibleInt64(1800 * 36))
	case T1830:
		return new(types.FlexibleInt64(1800 * 37))
	case T1900:
		return new(types.FlexibleInt64(1800 * 38))
	case T1930:
		return new(types.FlexibleInt64(1800 * 39))
	case T2000:
		return new(types.FlexibleInt64(1800 * 40))
	case T2030:
		return new(types.FlexibleInt64(1800 * 41))
	case T2100:
		return new(types.FlexibleInt64(1800 * 42))
	case T2130:
		return new(types.FlexibleInt64(1800 * 43))
	case T2200:
		return new(types.FlexibleInt64(1800 * 44))
	case T2230:
		return new(types.FlexibleInt64(1800 * 45))
	case T2300:
		return new(types.FlexibleInt64(1800 * 46))
	case T2330:
		return new(types.FlexibleInt64(1800 * 47))
	default:
		return nil
	}
}

func (s *ScheduleTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = ScheduleTime(str)

	return nil
}

type LoopSettings struct {
	MaximumBasalRatePerHour *types.FlexibleFloat64   `json:"maximumBasalRatePerHour,omitempty"`
	MaximumBolus            *types.FlexibleFloat64   `json:"maximumBolus,omitempty"`
	DosingStrategy          *types.ShortString       `json:"dosingStrategy,omitempty"`
	DosingEnabled           *types.FlexibleBool      `json:"dosingEnabled,omitempty"`
	MinimumBGGuard          *types.FlexibleFloat64   `json:"minimumBGGuard,omitempty"`
	DeviceToken             *types.DeviceTokenString `json:"deviceToken,omitempty"`
	BundleIdentifier        *types.ShortString       `json:"bundleIdentifier,omitempty"`
	PreMealTargetRange      []*types.FlexibleFloat64 `json:"preMealTargetRange,omitempty"`
	OverridePresets         []*OverridePreset        `json:"overridePresets,omitempty"`
}

// RemoteSettings are where and how LoopFollow sends Trio its remote commands, and Trio's override presets
type RemoteSettings struct {
	BundleIdentifier *types.ShortString       `json:"bundleIdentifier,omitempty"`
	DeviceToken      *types.DeviceTokenString `json:"deviceToken,omitempty"`
	IsAPNSProduction *types.FlexibleBool      `json:"isAPNSProduction,omitempty"`
	TeamID           *types.ShortString       `json:"teamID,omitempty"`
	ExpirationDate   *types.DateString        `json:"expirationDate,omitempty"`
	OverridePresets  []*TrioOverridePreset    `json:"overridePresets,omitempty"`
}

// TrioOverridePreset is one of Trio's override presets, with its target in mg/dL
type TrioOverridePreset struct {
	Name       *types.ShortString     `json:"name,omitempty"`
	Duration   *types.FlexibleFloat64 `json:"duration,omitempty"`
	Percentage *types.FlexibleFloat64 `json:"percentage,omitempty"`
	Target     *types.FlexibleFloat64 `json:"target,omitempty"`
}

type OverridePreset struct {
	Symbol                  *types.ShortString       `json:"symbol,omitempty"`
	Name                    *types.ShortString       `json:"name,omitempty"`
	TargetRange             []*types.FlexibleFloat64 `json:"targetRange,omitempty"`
	Duration                *types.FlexibleInt64     `json:"duration,omitempty"`
	InsulinNeedsScaleFactor *types.FlexibleFloat64   `json:"insulinNeedsScaleFactor,omitempty"`
}

func (t ProfileStore) GetDevice() *string {
	return t.Device.Stringp()
}

func (t ProfileStore) GetApp() *string {
	return t.App.Stringp()
}

func (t ProfileStore) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t ProfileStore) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t ProfileStore) GetUtcOffsetMinutes() *int64 {
	return t.UtcOffsetMinutes.Int64p()
}

func (t ProfileStore) GetIdentifier() *string {
	return t.Identifier.Stringp()
}

func (p *Profile) ConvertToMGDL() {
	p.TargetLow = ScheduleValuesToJSONWithUnitConversionToMGDL(p.TargetLow, p.Units)
	p.TargetHigh = ScheduleValuesToJSONWithUnitConversionToMGDL(p.TargetHigh, p.Units)
}

// WithGlucose is a copy of Loop's settings with each glucose in them, its suspend threshold, pre-meal range and
// override presets' ranges, converted, as Loop sends them in its profile's units
func (l *LoopSettings) WithGlucose(convert func(float64) float64) *LoopSettings {
	if l == nil {
		return nil
	}
	converted := *l
	glucose := func(values []*types.FlexibleFloat64) []*types.FlexibleFloat64 {
		if values == nil {
			return nil
		}
		out := make([]*types.FlexibleFloat64, len(values))
		for i, v := range values {
			if v != nil {
				out[i] = new(types.FlexibleFloat64(convert(v.Float64())))
			}
		}
		return out
	}
	if l.MinimumBGGuard != nil {
		converted.MinimumBGGuard = new(types.FlexibleFloat64(convert(l.MinimumBGGuard.Float64())))
	}
	converted.PreMealTargetRange = glucose(l.PreMealTargetRange)
	if l.OverridePresets != nil {
		converted.OverridePresets = make([]*OverridePreset, len(l.OverridePresets))
		for i, preset := range l.OverridePresets {
			if preset != nil {
				p := *preset
				p.TargetRange = glucose(preset.TargetRange)
				converted.OverridePresets[i] = &p
			}
		}
	}
	return &converted
}

func ScheduleValuesToJSONWithUnitConversionToMGDL(svs []*ScheduleValue, units *consts.Units) []*ScheduleValue {
	if len(svs) == 0 {
		return svs
	}

	for _, sv := range svs {
		// A null entry is kept as it is, as readers of the schedule skip it
		if sv != nil && sv.Value != nil {
			sv.Value = new(types.FlexibleFloat64(units.ToMGDL(sv.Value.Float64())))
		}
	}
	return svs
}
