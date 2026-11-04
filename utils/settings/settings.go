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

// GetRangeLowMgdl is the low of the range in mg/dL, the identity's, then the site's, then the default, whether or not its alarm is on
func (m *CombinedSettingsMap) GetRangeLowMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetRangeLowMgdl(), m.SiteSettingsMap.GetRangeLowMgdl(), m.DefaultSettingsMap.GetRangeLowMgdl())
}

// GetRangeHighMgdl is the high of the range in mg/dL, the identity's, then the site's, then the default, whether or not its alarm is on
func (m *CombinedSettingsMap) GetRangeHighMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetRangeHighMgdl(), m.SiteSettingsMap.GetRangeHighMgdl(), m.DefaultSettingsMap.GetRangeHighMgdl())
}

// GetUnitsLabel is how the Nightscout API writes units, the identity's, then the site's, then Nighttroop's own
func (m *CombinedSettingsMap) GetUnitsLabel(units consts.Units) string {
	if units == consts.UnitsMMOLL {
		return *first(m.IdentitySettingsMap.GetUnitsMmolLabel(), m.SiteSettingsMap.GetUnitsMmolLabel(), m.DefaultSettingsMap.GetUnitsMmolLabel())
	}
	return *first(m.IdentitySettingsMap.GetUnitsMgdlLabel(), m.SiteSettingsMap.GetUnitsMgdlLabel(), m.DefaultSettingsMap.GetUnitsMgdlLabel())
}

// GetTimezone is the identity's timezone, then the site's, then the default UTC
func (m *CombinedSettingsMap) GetTimezone() *string {
	return first(
		m.IdentitySettingsMap.GetTimezone(),
		m.SiteSettingsMap.GetTimezone(),
		m.DefaultSettingsMap.GetTimezone(),
	)
}

// GetWeekStartDay is the identity's week start day, then the site's, then the default monday
func (m *CombinedSettingsMap) GetWeekStartDay() *string {
	return first(m.IdentitySettingsMap.GetWeekStartDay(), m.SiteSettingsMap.GetWeekStartDay(), m.DefaultSettingsMap.GetWeekStartDay())
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
