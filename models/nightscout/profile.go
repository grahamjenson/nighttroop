package nightscout

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Profile struct {
	Identifier     *IdentifierString               `json:"identifier,omitempty"`
	DefaultProfile *ShortString                    `json:"defaultProfile,omitempty"`
	Store          map[ShortString]ProfileSettings `json:"store,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Date Feilds
	StartDate   *DateString `json:"startDate,omitempty"`
	Mills       *DateInt    `json:"mills,omitempty"`
	SrvCreated  *DateInt    `json:"srvCreated,omitempty"`
	SrvModified *DateInt    `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`
}

func (t *Profile) UnmarshalJSON(data []byte) error {
	type Alias Profile
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

type ProfileSettings struct {
	DIA        *FlexibleFloat64 `json:"dia,omitempty"`
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

// 00:00,
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

func (s *ScheduleTime) ToSeconds() (int, error) {
	parts := strings.Split(string(*s), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid schedule time format: %s", *s)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in schedule time: %w", err)
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in schedule time: %w", err)
	}
	return hours*3600 + minutes*60, nil
}

func (s *ScheduleTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = ScheduleTime(str)

	return nil
}
