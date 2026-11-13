package settings

import (
	"fmt"
	"strings"
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

// GetRangeUrgentLowMgdl is the urgent low in mg/dL, the identity's, then the site's, then the default, whether or not its alarm is on
func (m *CombinedSettingsMap) GetRangeUrgentLowMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetRangeUrgentLowMgdl(), m.SiteSettingsMap.GetRangeUrgentLowMgdl(), m.DefaultSettingsMap.GetRangeUrgentLowMgdl())
}

// GetRangeUrgentHighMgdl is the urgent high in mg/dL, the identity's, then the site's, then the default, whether or not its alarm is on
func (m *CombinedSettingsMap) GetRangeUrgentHighMgdl() *int64 {
	return first(m.IdentitySettingsMap.GetRangeUrgentHighMgdl(), m.SiteSettingsMap.GetRangeUrgentHighMgdl(), m.DefaultSettingsMap.GetRangeUrgentHighMgdl())
}

// GetAlarmNoDataMinutes is how long without a reading is no data, the identity's, then the site's, then the default,
// whether or not its alarm is on
func (m *CombinedSettingsMap) GetAlarmNoDataMinutes() *int64 {
	return first(m.IdentitySettingsMap.GetAlarmNoDataMinutes(), m.SiteSettingsMap.GetAlarmNoDataMinutes(), m.DefaultSettingsMap.GetAlarmNoDataMinutes())
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

// GetLanguage is the identity's language, then the site's, then the default
func (m *CombinedSettingsMap) GetLanguage() *string {
	return first(m.IdentitySettingsMap.GetLanguage(), m.SiteSettingsMap.GetLanguage(), m.DefaultSettingsMap.GetLanguage())
}

// DateLayout is the Go layout of the date format, the identity's, then the site's, then the default, e.g. 02/01/2006
// for DD/MM/YYYY
func (m *CombinedSettingsMap) DateLayout() string {
	format := first(m.IdentitySettingsMap.GetDateFormat(), m.SiteSettingsMap.GetDateFormat(), m.DefaultSettingsMap.GetDateFormat())
	if format == nil {
		return time.DateOnly
	}
	return strings.NewReplacer("YYYY", "2006", "MM", "01", "DD", "02").Replace(*format)
}

// timeLayouts are the Go layouts of the time formats, e.g. 03:04 PM for hh:mm a
var timeLayouts = strings.NewReplacer("HH", "15", "hh", "03", "mm", "04", "ss", "05", "a", "PM")

// TimeLayout is the Go layout of the time format, the identity's, then the site's, then the default, e.g. 03:04 PM
// for hh:mm a
func (m *CombinedSettingsMap) TimeLayout() string {
	format := first(m.IdentitySettingsMap.GetTimeFormat(), m.SiteSettingsMap.GetTimeFormat(), m.DefaultSettingsMap.GetTimeFormat())
	if format == nil {
		return time.TimeOnly
	}
	return timeLayouts.Replace(*format)
}

// GetWeekStartDay is the identity's week start day, then the site's, then the default monday
func (m *CombinedSettingsMap) GetWeekStartDay() *string {
	return first(m.IdentitySettingsMap.GetWeekStartDay(), m.SiteSettingsMap.GetWeekStartDay(), m.DefaultSettingsMap.GetWeekStartDay())
}

// GetNumberFormat is the identity's number format, then the site's, then the default 1,234.5
func (m *CombinedSettingsMap) GetNumberFormat() *string {
	return first(m.IdentitySettingsMap.GetNumberFormat(), m.SiteSettingsMap.GetNumberFormat(), m.DefaultSettingsMap.GetNumberFormat())
}

// locations are the timezones loaded, by name, as a page shows many times in one
var locations sync.Map

// Location is the timezone times are shown in, or UTC if it cannot be loaded
func (m *CombinedSettingsMap) Location() *time.Location {
	name := m.GetTimezone()
	if name == nil {
		return time.UTC
	}
	location, err := LoadLocation(*name)
	if err != nil {
		return time.UTC
	}
	return location
}

// LoadLocation is the named timezone, loaded once and then kept, as time.LoadLocation reads its zone file each time.
// Only a timezone that loads is kept, so there are at most as many as the zone database has.
func LoadLocation(name string) (*time.Location, error) {
	if location, ok := locations.Load(name); ok {
		return location.(*time.Location), nil
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	locations.Store(name, location)
	return location, nil
}
