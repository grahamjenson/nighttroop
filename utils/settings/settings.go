package settings

import (
	"fmt"
	"reflect"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/consts"
)

type CombinedSettingsMap struct {
	SiteSettingsMap     *SettingMap
	IdentitySettingsMap *SettingMap
	DefaultSettingsMap  *SettingMap
}

func NewCombinedSettingsMap(site *control.Site, identity *control.SiteIdentity) *CombinedSettingsMap {
	return &CombinedSettingsMap{
		SiteSettingsMap:     NewSettingsMap(site),
		IdentitySettingsMap: NewSettingsMap(identity),
		DefaultSettingsMap:  DefaultSettingsMap,
	}
}

func (m *CombinedSettingsMap) GetUnits() *consts.Units {
	return first(
		m.SiteSettingsMap.GetUnits(),
		m.IdentitySettingsMap.GetUnits(),
		m.DefaultSettingsMap.GetUnits(),
	)
}

func (m *CombinedSettingsMap) GetSettingAnyTypeByName(identifier string) (*SettingAnyType, error) {
	setting := m.SiteSettingsMap.GetSettingAnyTypeByName(identifier)
	if setting != nil {
		return setting, nil
	}

	setting = m.IdentitySettingsMap.GetSettingAnyTypeByName(identifier)
	if setting != nil {
		return setting, nil
	}

	setting = m.DefaultSettingsMap.GetSettingAnyTypeByName(identifier)
	if setting != nil {
		return setting, nil
	}

	return nil, fmt.Errorf("setting with identifier %s not found", identifier)
}

func (m *CombinedSettingsMap) GetAllSettingAnyTypes() []*SettingAnyType {
	allSettings := make([]*SettingAnyType, 0)
	allSettingNames := []string{
		"units",
		"timezone",
		"language",
		"date_format",
		"time_format",
		"week_start_day",
		"bolus_dia_hours",
		"bolus_onset_minutes",
		"bolus_peak_minutes",
		"basal_dia_hours",
		"basal_onset_minutes",
		"basal_peak_minutes",
		"default_insulin_sensitivity",
		"default_carb_ratio",
		"default_basal_rate",
		"carb_max_absorption_hours",
		"carb_min_5m_impact",
		"alarm_urgent_low_enabled",
		"alarm_urgent_low",
		"alarm_low_enabled",
		"alarm_low",
		"alarm_low_minutes",
		"alarm_low_prediction_enabled",
		"alarm_high_enabled",
		"alarm_high",
		"alarm_high_minutes",
		"alarm_urgent_high_enabled",
		"alarm_urgent_high",
		"alarm_fast_rise_enabled",
		"alarm_fast_rise",
		"alarm_fast_rise_minutes",
		"alarm_fast_drop_enabled",
		"alarm_fast_drop",
		"alarm_fast_drop_minutes",
		"alarm_no_data_enabled",
		"alarm_no_data_minutes",
		"alarm_battery_low_enabled",
		"alarm_battery_low_percent",
		"alarm_battery_low_pwd_enabled",
		"alarm_battery_low_pwd_percent",
		"alarm_low_reservoir_enabled",
		"alarm_low_reservoir_units",
		"notifications_enabled",
	}
	for _, name := range allSettingNames {
		setting, err := m.GetSettingAnyTypeByName(name)
		if err == nil && setting != nil {
			allSettings = append(allSettings, setting)
		}
	}

	return allSettings
}
func first[T any](values ...*T) *T {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
			if v.IsNil() {
				continue
			}
		}

		return value
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
