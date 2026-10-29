package sitemodels

// Settings are one level of settings, the site's or an identity's. A nil value is not set at
// this level, so the level above applies: the site's for an identity, the defaults for the site.
// Units is a string, not consts.Units, so an empty value can clear it.
type Settings struct {
	Units        *string `json:"units"`
	Timezone     *string `json:"timezone"`
	Language     *string `json:"language"`
	DateFormat   *string `json:"date_format"`
	TimeFormat   *string `json:"time_format"`
	WeekStartDay *string `json:"week_start_day"`

	// The alarms, each turned on with true or off with false, with the glucose it is under or over, or rises or
	// drops by, in the settings' units, and the minutes it must last or rise or drop over
	AlarmUrgentLowEnabled  *string `json:"alarm_urgent_low_enabled"`
	AlarmUrgentLow         *string `json:"alarm_urgent_low"`
	AlarmLowEnabled        *string `json:"alarm_low_enabled"`
	AlarmLow               *string `json:"alarm_low"`
	AlarmLowMinutes        *string `json:"alarm_low_minutes"`
	AlarmHighEnabled       *string `json:"alarm_high_enabled"`
	AlarmHigh              *string `json:"alarm_high"`
	AlarmHighMinutes       *string `json:"alarm_high_minutes"`
	AlarmUrgentHighEnabled *string `json:"alarm_urgent_high_enabled"`
	AlarmUrgentHigh        *string `json:"alarm_urgent_high"`
	AlarmFastRiseEnabled   *string `json:"alarm_fast_rise_enabled"`
	AlarmFastRise          *string `json:"alarm_fast_rise"`
	AlarmFastRiseMinutes   *string `json:"alarm_fast_rise_minutes"`
	AlarmFastDropEnabled   *string `json:"alarm_fast_drop_enabled"`
	AlarmFastDrop          *string `json:"alarm_fast_drop"`
	AlarmFastDropMinutes   *string `json:"alarm_fast_drop_minutes"`
	AlarmNoDataEnabled     *string `json:"alarm_no_data_enabled"`
	AlarmNoDataMinutes     *string `json:"alarm_no_data_minutes"`

	// Inherited are the values from the level above, by setting name
	Inherited map[string]string `json:"inherited,omitempty"`
	// Site is the site's settings, shown beside an admin's own on the settings page
	Site *Settings `json:"-"`
}

// Fields are the settings by name, to read or set
func (s *Settings) Fields() map[string]**string {
	return map[string]**string{
		"units": &s.Units, "timezone": &s.Timezone, "language": &s.Language,
		"date_format": &s.DateFormat, "time_format": &s.TimeFormat, "week_start_day": &s.WeekStartDay,
		"alarm_urgent_low_enabled": &s.AlarmUrgentLowEnabled, "alarm_urgent_low": &s.AlarmUrgentLow,
		"alarm_low_enabled": &s.AlarmLowEnabled, "alarm_low": &s.AlarmLow, "alarm_low_minutes": &s.AlarmLowMinutes,
		"alarm_high_enabled": &s.AlarmHighEnabled, "alarm_high": &s.AlarmHigh, "alarm_high_minutes": &s.AlarmHighMinutes,
		"alarm_urgent_high_enabled": &s.AlarmUrgentHighEnabled, "alarm_urgent_high": &s.AlarmUrgentHigh,
		"alarm_fast_rise_enabled": &s.AlarmFastRiseEnabled, "alarm_fast_rise": &s.AlarmFastRise, "alarm_fast_rise_minutes": &s.AlarmFastRiseMinutes,
		"alarm_fast_drop_enabled": &s.AlarmFastDropEnabled, "alarm_fast_drop": &s.AlarmFastDrop, "alarm_fast_drop_minutes": &s.AlarmFastDropMinutes,
		"alarm_no_data_enabled": &s.AlarmNoDataEnabled, "alarm_no_data_minutes": &s.AlarmNoDataMinutes,
	}
}
