package sitemodels

import "nighttroop.t1d.nz/utils/consts"

// Summary is what happened on the site in the last Hours, each list newest first
type Summary struct {
	Hours int64 `json:"hours"`
	// Since is when the summary starts, in milliseconds
	Since int64         `json:"since"`
	Units *consts.Units `json:"units"`
	// LowAlarm and HighAlarm are the alarm settings' low and high, in Units: a reading below or above is out of range
	LowAlarm  float64 `json:"low_alarm"`
	HighAlarm float64 `json:"high_alarm"`

	GlucoseReadings []*SiteGlucoseReading `json:"glucose_readings"`
	InsulinDoses    []*SiteInsulinDose    `json:"insulin_doses"`
	FoodIntakes     []*SiteFoodIntake     `json:"food_intakes"`
	TempTargets     []*SiteTempTarget     `json:"temp_targets"`
	TempBasals      []*SiteTempBasal      `json:"temp_basals"`

	// DeviceStatuses is each device's latest status
	DeviceStatuses []*SiteDeviceStatus `json:"device_statuses"`
	// Loop is the latest status from the PWD's phone with loop information, for its IOB and COB
	Loop *SiteDeviceStatus `json:"loop"`

	// Minutes are the targets and basal in force each minute from Since, the first at Since and each a minute after
	Minutes []SummaryMinute `json:"minutes"`
}
