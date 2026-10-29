package settings

import (
	"fmt"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

type CombinedSettingsMap struct {
	SiteSettingsMap     *SettingMap
	IdentitySettingsMap *SettingMap
	DefaultSettingsMap  *SettingMap
}

func DefaultCombinedSettingsMap() *CombinedSettingsMap {
	return &CombinedSettingsMap{
		SiteSettingsMap:     DefaultSettingsMap,
		IdentitySettingsMap: DefaultSettingsMap,
		DefaultSettingsMap:  DefaultSettingsMap,
	}
}

func NewCombinedSettingsMap(site *controldb.Site, identity *controldb.SiteIdentity) *CombinedSettingsMap {
	return &CombinedSettingsMap{
		SiteSettingsMap:     NewSettingsMap(site, &site.CreatedAt, ptr("site")),
		IdentitySettingsMap: NewSettingsMap(identity, &identity.CreatedAt, ptr("identity")),
		DefaultSettingsMap:  DefaultSettingsMap,
	}
}

func (m *CombinedSettingsMap) GetSettingAnyTypeByName(identifier string) (*SettingAnyType, error) {
	setting := first(
		m.IdentitySettingsMap.GetSettingAnyTypeByName(identifier),
		m.SiteSettingsMap.GetSettingAnyTypeByName(identifier),
		m.DefaultSettingsMap.GetSettingAnyTypeByName(identifier),
	)

	if setting != nil {
		return setting, nil
	}

	return nil, fmt.Errorf("setting with identifier %s not found", identifier)
}

func (m *CombinedSettingsMap) GetAllSettingAnyTypes() []*SettingAnyType {
	allSettings := make([]*SettingAnyType, 0)

	for _, name := range allSettingNames {
		setting, err := m.GetSettingAnyTypeByName(name)
		if err == nil && setting != nil {
			allSettings = append(allSettings, setting)
		}
	}

	return allSettings
}

// GetUnits checks the identity first, like every other setting, so a user's own units win over the site's
func (m *CombinedSettingsMap) GetUnits() *consts.Units {
	return first(
		m.IdentitySettingsMap.GetUnits(),
		m.SiteSettingsMap.GetUnits(),
		m.DefaultSettingsMap.GetUnits(),
	)
}

// GetAlarmLowMgdl is the low alarm's mg/dL, the identity's, then the site's, then the default, whether or not it is on
func (m *CombinedSettingsMap) GetAlarmLowMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetAlarmLowMgdl(), m.SiteSettingsMap.GetAlarmLowMgdl(), m.DefaultSettingsMap.GetAlarmLowMgdl())
}

// GetAlarmHighMgdl is the high alarm's mg/dL, the identity's, then the site's, then the default, whether or not it is on
func (m *CombinedSettingsMap) GetAlarmHighMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetAlarmHighMgdl(), m.SiteSettingsMap.GetAlarmHighMgdl(), m.DefaultSettingsMap.GetAlarmHighMgdl())
}

// GetTimezone is the identity's timezone, then the site's, then the default UTC
func (m *CombinedSettingsMap) GetTimezone() *string {
	return first(
		m.IdentitySettingsMap.GetTimezone(),
		m.SiteSettingsMap.GetTimezone(),
		m.DefaultSettingsMap.GetTimezone(),
	)
}

// locations are the timezones loaded, by name, as a page shows many times in one
var locations sync.Map

// Location is the timezone times are shown in, or UTC if it cannot be loaded
func (m *CombinedSettingsMap) Location() *time.Location {
	name := m.GetTimezone()
	if name == nil {
		return time.UTC
	}
	if location, ok := locations.Load(*name); ok {
		return location.(*time.Location)
	}
	location, err := time.LoadLocation(*name)
	if err != nil {
		return time.UTC
	}
	locations.Store(*name, location)
	return location
}
