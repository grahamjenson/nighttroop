package settings

import (
	"fmt"

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
