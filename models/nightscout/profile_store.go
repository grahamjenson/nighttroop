package nightscout

import (
	"encoding/json"
)

type ProfileStore struct {
	Identifier     *IdentifierString       `json:"identifier,omitempty"`
	DefaultProfile *ShortString            `json:"defaultProfile,omitempty"`
	Store          map[ShortString]Profile `json:"store,omitempty"`
	Units          *Units                  `json:"units,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Date Feilds
	StartDate *DateString `json:"startDate,omitempty"`
	Mills     *DateInt    `json:"mills,omitempty"`

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`
}

func (t *ProfileStore) UnmarshalJSON(data []byte) error {
	type Alias ProfileStore
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

	// Set default StartDate (leave UTCOffset nil to be later replaces by user timezone)
	if t.StartDate == nil {
		// This is UTC not User timezone, will have to set UTC offset minutes at API level
		ds, now := DateStringNow()
		t.StartDate = ds
		t.Date = ptr(DateIntFromTime(now))
	} else {
		t.Date, t.UtcOffsetMinutes = t.StartDate.MillisUTCOffset()
	}

	return nil
}

type Profile struct {
	Dia        *FlexibleFloat64 `json:"dia,omitempty"`
	CarbRatio  []*ScheduleValue `json:"carbratio,omitempty"`
	Sens       []*ScheduleValue `json:"sens,omitempty"`
	Basal      []*ScheduleValue `json:"basal,omitempty"`
	TargetLow  []*ScheduleValue `json:"target_low,omitempty"`
	TargetHigh []*ScheduleValue `json:"target_high,omitempty"`

	Timezone *Timezone `json:"timezone,omitempty"`
	Units    *Units    `json:"units,omitempty"`
}

type ScheduleValue struct {
	Time          *ScheduleTime    `json:"time,omitempty"`
	TimeAsSeconds *FlexibleInt64   `json:"timeAsSeconds,omitempty"`
	Value         *FlexibleFloat64 `json:"value,omitempty"`
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

func (s *ScheduleTime) ToSeconds() *FlexibleInt64 {
	switch *s {
	case T0000:
		return ptr(FlexibleInt64(0))
	case T0030:
		return ptr(FlexibleInt64(1800))
	case T0100:
		return ptr(FlexibleInt64(1800 * 2))
	case T0130:
		return ptr(FlexibleInt64(1800 * 3))
	case T0200:
		return ptr(FlexibleInt64(1800 * 4))
	case T0230:
		return ptr(FlexibleInt64(1800 * 5))
	case T0300:
		return ptr(FlexibleInt64(1800 * 6))
	case T0330:
		return ptr(FlexibleInt64(1800 * 7))
	case T0400:
		return ptr(FlexibleInt64(1800 * 8))
	case T0430:
		return ptr(FlexibleInt64(1800 * 9))
	case T0500:
		return ptr(FlexibleInt64(1800 * 10))
	case T0530:
		return ptr(FlexibleInt64(1800 * 11))
	case T0600:
		return ptr(FlexibleInt64(1800 * 12))
	case T0630:
		return ptr(FlexibleInt64(1800 * 13))
	case T0700:
		return ptr(FlexibleInt64(1800 * 14))
	case T0730:
		return ptr(FlexibleInt64(1800 * 15))
	case T0800:
		return ptr(FlexibleInt64(1800 * 16))
	case T0830:
		return ptr(FlexibleInt64(1800 * 17))
	case T0900:
		return ptr(FlexibleInt64(1800 * 18))
	case T0930:
		return ptr(FlexibleInt64(1800 * 19))
	case T1000:
		return ptr(FlexibleInt64(1800 * 20))
	case T1030:
		return ptr(FlexibleInt64(1800 * 21))
	case T1100:
		return ptr(FlexibleInt64(1800 * 22))
	case T1130:
		return ptr(FlexibleInt64(1800 * 23))
	case T1200:
		return ptr(FlexibleInt64(1800 * 24))
	case T1230:
		return ptr(FlexibleInt64(1800 * 25))
	case T1300:
		return ptr(FlexibleInt64(1800 * 26))
	case T1330:
		return ptr(FlexibleInt64(1800 * 27))
	case T1400:
		return ptr(FlexibleInt64(1800 * 28))
	case T1430:
		return ptr(FlexibleInt64(1800 * 29))
	case T1500:
		return ptr(FlexibleInt64(1800 * 30))
	case T1530:
		return ptr(FlexibleInt64(1800 * 31))
	case T1600:
		return ptr(FlexibleInt64(1800 * 32))
	case T1630:
		return ptr(FlexibleInt64(1800 * 33))
	case T1700:
		return ptr(FlexibleInt64(1800 * 34))
	case T1730:
		return ptr(FlexibleInt64(1800 * 35))
	case T1800:
		return ptr(FlexibleInt64(1800 * 36))
	case T1830:
		return ptr(FlexibleInt64(1800 * 37))
	case T1900:
		return ptr(FlexibleInt64(1800 * 38))
	case T1930:
		return ptr(FlexibleInt64(1800 * 39))
	case T2000:
		return ptr(FlexibleInt64(1800 * 40))
	case T2030:
		return ptr(FlexibleInt64(1800 * 41))
	case T2100:
		return ptr(FlexibleInt64(1800 * 42))
	case T2130:
		return ptr(FlexibleInt64(1800 * 43))
	case T2200:
		return ptr(FlexibleInt64(1800 * 44))
	case T2230:
		return ptr(FlexibleInt64(1800 * 45))
	case T2300:
		return ptr(FlexibleInt64(1800 * 46))
	case T2330:
		return ptr(FlexibleInt64(1800 * 47))
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
