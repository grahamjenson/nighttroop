package settings

import (
	"fmt"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
)

// Settings are the settings of a request, or of a site for a background job: every setting with its value, at each
// level, built once from the site's row and the identity's.
type Settings struct {
	// Default is every setting at its default
	Default Values
	// Site is the site's own settings over the defaults: what is the site's whoever is signed in, such as the
	// timezone its days are in
	Site Values
	// Values are the settings in effect: the identity's own over the site's over the defaults. Without an identity
	// they are the site's.
	Values

	// SiteLevel and IdentityLevel are what the site and the identity have each set of their own. After one is
	// changed, Rebuild makes Site and Values follow.
	SiteLevel, IdentityLevel *Level
}

// SettingAnyType is one setting as the APIs give it: its value in effect, and the level it is from
type SettingAnyType struct {
	Identifier string
	TypeString string
	Value      any
	CreatedAt  int64
	Subject    string
}

// started is when the defaults are from, as each setting has a time
var started = time.Now().UnixMilli()

// New are the settings of a site, and of an identity of it or nil for the site alone
func New(site *controldb.Site, identity *controldb.SiteIdentity) *Settings {
	siteLevel, identityLevel := &Level{Subject: "site", own: map[string]any{}}, &Level{Subject: "identity", own: map[string]any{}}
	if site != nil {
		siteLevel = LevelFromJSON(site.SettingsJson, "site", site.CreatedAt)
	}
	if identity != nil {
		identityLevel = LevelFromJSON(identity.SettingsJson, "identity", identity.CreatedAt)
	}
	return FromLevels(siteLevel, identityLevel)
}

// FromLevels are the settings with what the site and the identity have each set, nil for nothing
func FromLevels(site, identity *Level) *Settings {
	s := &Settings{SiteLevel: site, IdentityLevel: identity}
	if s.SiteLevel == nil {
		s.SiteLevel = &Level{Subject: "site", own: map[string]any{}}
	}
	if s.IdentityLevel == nil {
		s.IdentityLevel = &Level{Subject: "identity", own: map[string]any{}}
	}
	return s.Rebuild()
}

// Rebuild makes Site and Values follow the levels, after a level is changed
func (s *Settings) Rebuild() *Settings {
	s.Default = DefaultValues()
	s.Site = DefaultValues()
	s.SiteLevel.applyTo(&s.Site)
	s.Values = s.Site
	// A list is the level's own, not shared with the site's
	s.Values.Alarms.DeviceBatteryLow.Devices = append(s.Values.Alarms.DeviceBatteryLow.Devices[:0:0], s.Site.Alarms.DeviceBatteryLow.Devices...)
	s.IdentityLevel.applyTo(&s.Values)
	return s
}

// Get is a setting in effect by its identifier, with the level it is from: the identity's, then the site's, then
// the default
func (s *Settings) Get(identifier string) (*SettingAnyType, error) {
	f, ok := byName[identifier]
	if !ok {
		return nil, fmt.Errorf("setting with identifier %s not found", identifier)
	}
	setting := &SettingAnyType{Identifier: identifier, TypeString: f.typeString, Value: f.fallback, CreatedAt: started, Subject: "default"}
	for _, level := range []*Level{s.SiteLevel, s.IdentityLevel} {
		if value, ok := level.Get(identifier); ok {
			setting.Value, setting.CreatedAt, setting.Subject = value, level.At, level.Subject
		}
	}
	return setting, nil
}

// All is every setting in effect, in the order the settings are listed
func (s *Settings) All() []*SettingAnyType {
	all := make([]*SettingAnyType, 0, len(names))
	for _, name := range names {
		if setting, err := s.Get(name); err == nil {
			all = append(all, setting)
		}
	}
	return all
}

// locations are the timezones loaded, by name, as a page shows many times in one
var locations sync.Map

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
