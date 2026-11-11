package sitemodels

import "nighttroop.t1d.nz/utils/consts"

// Summary is what happened on the site in the last Hours, each list newest first
type Summary struct {
	Hours int64 `json:"hours"`
	// Since is when the summary starts, in milliseconds
	Since int64         `json:"since"`
	Units *consts.Units `json:"units"`
	// RangeLow and RangeHigh are the range settings' low and high, in Units: a reading below or above is out of range
	RangeLow  float64 `json:"range_low"`
	RangeHigh float64 `json:"range_high"`
	// RangeUrgentLow and RangeUrgentHigh are the range settings' urgent low and high, in Units
	RangeUrgentLow  float64 `json:"range_urgent_low"`
	RangeUrgentHigh float64 `json:"range_urgent_high"`

	// GlucoseReadings are cleaned, the same glucose from more than one source kept once
	GlucoseReadings []*SiteCleanGlucoseReading `json:"glucose_readings"`
	InsulinDoses    []*SiteInsulinDose         `json:"insulin_doses"`
	FoodIntakes     []*SiteFoodIntake          `json:"food_intakes"`
	TempTargets     []*SiteTempTarget          `json:"temp_targets"`
	TempBasals      []*SiteTempBasal           `json:"temp_basals"`

	// DeviceStatuses is each device's latest status
	DeviceStatuses []*SiteDeviceStatus `json:"device_statuses"`
	// Loop is the latest status from the PWD's phone with loop information, for its IOB and COB
	Loop *SiteDeviceStatus `json:"loop"`

	// Alarms are the identity's active alarms, and those it has snoozed that are not, the most urgent first
	Alarms []*SiteAlarm `json:"alarms"`

	// Minutes are the targets and basal in force each minute from Since, the first at Since and each a minute after
	Minutes []SummaryMinute `json:"minutes"`

	// Today is the day so far from midnight in the settings' timezone, and Yesterday the whole day before it
	Today     SummaryDay `json:"today"`
	Yesterday SummaryDay `json:"yesterday"`
	// LastWeek, LastFortnight and LastMonth are the mean day of the last whole week, fortnight and month before
	// today's, from the settings' week start day, their time in range and total daily dose the means of their days with
	// data
	LastWeek      SummaryDay `json:"last_week"`
	LastFortnight SummaryDay `json:"last_fortnight"`
	LastMonth     SummaryDay `json:"last_month"`
	// SensorStartedAt is when the latest CGM sensor was started or changed, and SiteChangedAt when the pump site was
	// last changed, in milliseconds, null when there is none
	SensorStartedAt *int64 `json:"sensor_started_at"`
	SiteChangedAt   *int64 `json:"site_changed_at"`
}
