package sitemodels

import (
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// SiteProfilePreset is one named profile of a profile store a loop uploaded, with its schedules through the day.
// Sensitivities and targets are in the settings' units.
type SiteProfilePreset struct {
	ID                    string               `json:"id"`
	Name                  *string              `json:"name"`
	ProfileStoreID        string               `json:"profile_store_id"`
	IsDefault             bool                 `json:"is_default"`
	Dia                   *float64             `json:"dia"`
	Timezone              *string              `json:"timezone"`
	CarbsHr               *float64             `json:"carbs_hr"`
	Delay                 *float64             `json:"delay"`
	Units                 *consts.Units        `json:"units"`
	CarbRatios            []*SiteScheduleValue `json:"carb_ratios"`
	Sensitivities         []*SiteScheduleValue `json:"sensitivities"`
	BasalRates            []*SiteScheduleValue `json:"basal_rates"`
	LowTargets            []*SiteScheduleValue `json:"low_targets"`
	HighTargets           []*SiteScheduleValue `json:"high_targets"`
	UtcOffsetMinutes      *int64               `json:"utc_offset_minutes"`
	OccurredAt            int64                `json:"occurred_at"`
	CreatedAt             int64                `json:"created_at"`
	UpdatedAt             int64                `json:"updated_at"`
	CreatedByAppName      *string              `json:"created_by_app_name"`
	CreatedByDeviceName   *string              `json:"created_by_device_name"`
	CreatedByIdentityName string               `json:"created_by_identity_name"`
	UpdatedByIdentityName *string              `json:"updated_by_identity_name"`
}

// SiteScheduleValue is a profile's value from a time of day, in seconds from midnight, until the next one
type SiteScheduleValue struct {
	Seconds int64   `json:"seconds"`
	Value   float64 `json:"value"`
}

// scheduleFromJSON is a schedule stored as Nightscout's, each value converted by convert
func scheduleFromJSON(jsonStr *string, convert func(float64) float64) []*SiteScheduleValue {
	schedule := []*SiteScheduleValue{}
	for _, sv := range nightscoutmodels.JSONToScheduleValueSlice(jsonStr) {
		if sv == nil || sv.Value == nil {
			continue
		}
		value := &SiteScheduleValue{Value: convert(sv.Value.Float64())}
		if sv.TimeAsSeconds != nil {
			value.Seconds = sv.TimeAsSeconds.Int64()
		}
		schedule = append(schedule, value)
	}
	return schedule
}

// ScheduleFromJSON is a schedule of carb ratios or basal rates, as stored
func ScheduleFromJSON(jsonStr *string) []*SiteScheduleValue {
	return scheduleFromJSON(jsonStr, func(value float64) float64 { return value })
}

// GlucoseScheduleFromJSON is a schedule of targets, stored in mg/dL, in the settings' units
// goverter:context siteContext
func GlucoseScheduleFromJSON(jsonStr *string, siteContext *httputils.SiteContext) []*SiteScheduleValue {
	return scheduleFromJSON(jsonStr, func(mgdl float64) float64 { return MgdlToSettingsUnits(mgdl, siteContext) })
}

// ProfileSensitivities are a profile's sensitivities, stored in its own units, in the settings' units
// goverter:context siteContext
func ProfileSensitivities(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) []*SiteScheduleValue {
	return SensitivityScheduleFromJSON(source.InsulinSensitivitiesJson, source.Units, siteContext)
}

// SensitivityScheduleFromJSON is a schedule of sensitivities, stored in the profile's units, in the settings' units
func SensitivityScheduleFromJSON(jsonStr *string, units *consts.Units, siteContext *httputils.SiteContext) []*SiteScheduleValue {
	return scheduleFromJSON(jsonStr, func(value float64) float64 { return MgdlToSettingsUnits(units.ToMGDL(value), siteContext) })
}
