package settings

import (
	"fmt"

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

func (m *CombinedSettingsMap) GetUnits() *consts.Units {
	return first(
		m.SiteSettingsMap.GetUnits(),
		m.IdentitySettingsMap.GetUnits(),
		m.DefaultSettingsMap.GetUnits(),
	)
}
