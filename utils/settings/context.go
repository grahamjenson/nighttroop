package settings

import (
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// ContextSettings are the settings of a request, or of a site for a background job: what the identity has set, what
// the site has set, and the defaults. A setting is read with its getter, which is the identity's, then the site's,
// then the default, so it is never "not set".
type ContextSettings struct {
	DefaultSettings SettingsValues
	Identity        Settings
	Site            Settings

	// SiteSince and IdentitySince are when the site and the identity were made, the time each of their settings has
	SiteSince, IdentitySince int64
}

// NewContextSettings are the settings of a site, and of an identity of it or nil for the site alone
func NewContextSettings(site *controldb.Site, identity *controldb.SiteIdentity) *ContextSettings {
	c := &ContextSettings{DefaultSettings: DefaultValues()}
	if site != nil {
		c.Site, c.SiteSince = SettingsFromJSON(site.SettingsJson), site.CreatedAt
	}
	if identity != nil {
		c.Identity, c.IdentitySince = SettingsFromJSON(identity.SettingsJson), identity.CreatedAt
	}
	return c
}

// first is the first of the levels that is set, or the fallback
func first[T any](fallback T, levels ...*T) T {
	for _, level := range levels {
		if level != nil {
			return *level
		}
	}
	return fallback
}

// GetUnits are the identity's units, then the site's, then the default, as every getter here
func (c *ContextSettings) GetUnits() consts.Units {
	return first(c.DefaultSettings.Units, c.Identity.Units, c.Site.Units)
}

func (c *ContextSettings) GetTimezone() string {
	return first(c.DefaultSettings.Timezone, c.Identity.Timezone, c.Site.Timezone)
}

func (c *ContextSettings) GetLanguage() string {
	return first(c.DefaultSettings.Language, c.Identity.Language, c.Site.Language)
}

func (c *ContextSettings) GetDateFormat() string {
	return first(c.DefaultSettings.DateFormat, c.Identity.DateFormat, c.Site.DateFormat)
}

func (c *ContextSettings) GetTimeFormat() string {
	return first(c.DefaultSettings.TimeFormat, c.Identity.TimeFormat, c.Site.TimeFormat)
}

func (c *ContextSettings) GetWeekStartDay() string {
	return first(c.DefaultSettings.WeekStartDay, c.Identity.WeekStartDay, c.Site.WeekStartDay)
}

func (c *ContextSettings) GetNumberFormat() string {
	return first(c.DefaultSettings.NumberFormat, c.Identity.NumberFormat, c.Site.NumberFormat)
}

// GetRange is the glucose range in mg/dL, each of its four the identity's, then the site's, then the default
func (c *ContextSettings) GetRange() GlucoseRangeValues {
	fallback, identity, site := c.DefaultSettings.Range, c.Identity.Range, c.Site.Range
	return GlucoseRangeValues{
		UrgentLowMgdl:  first(fallback.UrgentLowMgdl, identity.UrgentLowMgdl, site.UrgentLowMgdl),
		LowMgdl:        first(fallback.LowMgdl, identity.LowMgdl, site.LowMgdl),
		HighMgdl:       first(fallback.HighMgdl, identity.HighMgdl, site.HighMgdl),
		UrgentHighMgdl: first(fallback.UrgentHighMgdl, identity.UrgentHighMgdl, site.UrgentHighMgdl),
	}
}

// GetAlarmNoDataMinutes is how long without a reading is no data, whether or not its alarm is on
func (c *ContextSettings) GetAlarmNoDataMinutes() int64 {
	return first(c.DefaultSettings.Alarms.NoData.Minutes, c.Identity.Alarms.NoData.Minutes, c.Site.Alarms.NoData.Minutes)
}

// Location is the timezone times are shown in, or UTC if it cannot be loaded
func (c *ContextSettings) Location() *time.Location {
	return SettingsValues{Timezone: c.GetTimezone()}.Location()
}

// DateLayout and TimeLayout are the Go layouts of the date and time formats
func (c *ContextSettings) DateLayout() string {
	return SettingsValues{DateFormat: c.GetDateFormat()}.DateLayout()
}

func (c *ContextSettings) TimeLayout() string {
	return SettingsValues{TimeFormat: c.GetTimeFormat()}.TimeLayout()
}

// The site's own settings over the defaults, whoever is signed in: for what is the site's, such as the timezone its
// days are in

func (c *ContextSettings) GetSiteTimezone() string {
	return first(c.DefaultSettings.Timezone, c.Site.Timezone)
}

func (c *ContextSettings) GetSiteWeekStartDay() string {
	return first(c.DefaultSettings.WeekStartDay, c.Site.WeekStartDay)
}

// GetSiteRange is the site's own glucose range, low and high in mg/dL
func (c *ContextSettings) GetSiteRange() (low, high int64) {
	return first(c.DefaultSettings.Range.LowMgdl, c.Site.Range.LowMgdl), first(c.DefaultSettings.Range.HighMgdl, c.Site.Range.HighMgdl)
}

// SetIdentityUnits sets the identity's own units
func (c *ContextSettings) SetIdentityUnits(units consts.Units) {
	c.Identity.Units = &units
}

// Values is every setting in effect: the identity's, then the site's, then the default
func (c *ContextSettings) Values() SettingsValues {
	v := c.DefaultSettings
	for _, f := range allSettings {
		for _, level := range []*Settings{&c.Site, &c.Identity} {
			if value, ok := f.own(level); ok {
				f.setValue(&v, value)
			}
		}
	}
	return v
}

// SettingAnyType is one setting as the APIs give it: its value, and the level it is from
type SettingAnyType struct {
	Identifier string
	TypeString string
	Value      any
	CreatedAt  int64
	Subject    string
}

// started is when the defaults are from, as each setting has a time
var started = time.Now().UnixMilli()

// Get is a setting in effect by its identifier, with the level it is from: the identity's, then the site's, then
// the default
func (c *ContextSettings) Get(identifier string) (*SettingAnyType, error) {
	f, ok := byName[identifier]
	if !ok {
		return nil, fmt.Errorf("setting with identifier %s not found", identifier)
	}
	if setting := c.Own(&c.Identity, identifier); setting != nil {
		return setting, nil
	}
	if setting := c.Own(&c.Site, identifier); setting != nil {
		return setting, nil
	}
	return &SettingAnyType{Identifier: identifier, TypeString: f.typeString, Value: f.fallback, CreatedAt: started, Subject: "default"}, nil
}

// Own is a setting as one level, c's Site or Identity, has set it, or nil when it has not
func (c *ContextSettings) Own(level *Settings, identifier string) *SettingAnyType {
	value, ok := level.Get(identifier)
	if !ok {
		return nil
	}
	setting := &SettingAnyType{Identifier: identifier, TypeString: byName[identifier].typeString, Value: value, CreatedAt: c.IdentitySince, Subject: "identity"}
	if level == &c.Site {
		setting.CreatedAt, setting.Subject = c.SiteSince, "site"
	}
	return setting
}

// All is every setting in effect, in the order the settings are listed
func (c *ContextSettings) All() []*SettingAnyType {
	all := make([]*SettingAnyType, 0, len(names))
	for _, name := range names {
		if setting, err := c.Get(name); err == nil {
			all = append(all, setting)
		}
	}
	return all
}
