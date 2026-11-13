package sitemodels

// Settings are one level of settings, the site's or an identity's. A nil value is not set at
// this level, so the level above applies: the site's for an identity, the defaults for the site.
// Units is a string, not consts.Units, so an empty value can clear it.
type Settings struct {
	Units *string `json:"units"`
	// How the Nightscout API writes each units, e.g. mg/dl for xDrip4iOS
	UnitsMgdlLabel *string `json:"units_mgdl_label"`
	UnitsMmolLabel *string `json:"units_mmol_label"`
	Timezone       *string `json:"timezone"`
	Language       *string `json:"language"`
	DateFormat     *string `json:"date_format"`
	TimeFormat     *string `json:"time_format"`
	WeekStartDay   *string `json:"week_start_day"`
	// How numbers are written: 1,234.5, 1.234,5 or 1 234,5
	NumberFormat *string `json:"number_format"`

	// The alarms, each turned on with true or off with false, with the range's glucose it is under or over, or rises or
	// drops by, in the settings' units, and the minutes it must last or rise or drop over
	AlarmUrgentLowEnabled  *string `json:"alarm_urgent_low_enabled"`
	RangeUrgentLow         *string `json:"range_urgent_low"`
	AlarmLowEnabled        *string `json:"alarm_low_enabled"`
	RangeLow               *string `json:"range_low"`
	AlarmLowMinutes        *string `json:"alarm_low_minutes"`
	AlarmHighEnabled       *string `json:"alarm_high_enabled"`
	RangeHigh              *string `json:"range_high"`
	AlarmHighMinutes       *string `json:"alarm_high_minutes"`
	AlarmUrgentHighEnabled *string `json:"alarm_urgent_high_enabled"`
	RangeUrgentHigh        *string `json:"range_urgent_high"`
	AlarmFastRiseEnabled   *string `json:"alarm_fast_rise_enabled"`
	AlarmFastRise          *string `json:"alarm_fast_rise"`
	AlarmFastRiseMinutes   *string `json:"alarm_fast_rise_minutes"`
	AlarmFastDropEnabled   *string `json:"alarm_fast_drop_enabled"`
	AlarmFastDrop          *string `json:"alarm_fast_drop"`
	AlarmFastDropMinutes   *string `json:"alarm_fast_drop_minutes"`
	AlarmNoDataEnabled     *string `json:"alarm_no_data_enabled"`
	AlarmNoDataMinutes     *string `json:"alarm_no_data_minutes"`

	// The predicted low follows the current trend for its minutes to the low alarm's glucose
	AlarmLowPredictionEnabled *string `json:"alarm_low_prediction_enabled"`
	AlarmLowPredictionMinutes *string `json:"alarm_low_prediction_minutes"`

	// Inherited are the values from the level above, by setting name
	Inherited map[string]string `json:"inherited,omitempty"`
	// Site is the site's settings, shown beside an admin's own on the settings page
	Site *Settings `json:"-"`

	// IdentityID and IdentityName are whose settings these are, when an admin is changing another identity's
	IdentityID   *int64  `json:"identity_id,omitempty"`
	IdentityName *string `json:"identity_name,omitempty"`
	// OwnerID is the identity whose own settings these are, or nil for the site's, so a change to them is streamed
	// only to whom they may be shown
	OwnerID *int64 `json:"-"`
}

// Fields are the settings by name, to read or set
func (s *Settings) Fields() map[string]**string {
	return map[string]**string{
		"units": &s.Units, "units_mgdl_label": &s.UnitsMgdlLabel, "units_mmol_label": &s.UnitsMmolLabel,
		"timezone": &s.Timezone, "language": &s.Language,
		"date_format": &s.DateFormat, "time_format": &s.TimeFormat, "week_start_day": &s.WeekStartDay,
		"number_format": &s.NumberFormat, "alarm_urgent_low_enabled": &s.AlarmUrgentLowEnabled, "range_urgent_low": &s.RangeUrgentLow,
		"alarm_low_enabled": &s.AlarmLowEnabled, "range_low": &s.RangeLow, "alarm_low_minutes": &s.AlarmLowMinutes,
		"alarm_low_prediction_enabled": &s.AlarmLowPredictionEnabled, "alarm_low_prediction_minutes": &s.AlarmLowPredictionMinutes,
		"alarm_high_enabled": &s.AlarmHighEnabled, "range_high": &s.RangeHigh, "alarm_high_minutes": &s.AlarmHighMinutes,
		"alarm_urgent_high_enabled": &s.AlarmUrgentHighEnabled, "range_urgent_high": &s.RangeUrgentHigh,
		"alarm_fast_rise_enabled": &s.AlarmFastRiseEnabled, "alarm_fast_rise": &s.AlarmFastRise, "alarm_fast_rise_minutes": &s.AlarmFastRiseMinutes,
		"alarm_fast_drop_enabled": &s.AlarmFastDropEnabled, "alarm_fast_drop": &s.AlarmFastDrop, "alarm_fast_drop_minutes": &s.AlarmFastDropMinutes,
		"alarm_no_data_enabled": &s.AlarmNoDataEnabled, "alarm_no_data_minutes": &s.AlarmNoDataMinutes,
	}
}
