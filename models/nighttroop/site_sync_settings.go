package nighttroop

import "nighttroop.t1d.nz/utils/consts"

type SiteSyncSettings struct {
	NightscoutSyncSettings  *NightscoutSyncSettings  `json:"nightscout"`
	DexcomSyncSettings      *DexcomSyncSettings      `json:"dexcom"`
	VirtualPumpSyncSettings *VirtualPumpSyncSettings `json:"virtual_pump"`
}

type VirtualPumpSyncSettings struct {
	Enabled bool `json:"enabled"`
}

type NightscoutSyncSettings struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func (n *NightscoutSyncSettings) GetURL() string {
	if n == nil {
		return ""
	}
	return n.URL
}

func (n *NightscoutSyncSettings) GetPasswordGarbage() string {
	if n == nil {
		return ""
	}
	return consts.PASSWORD_GARBAGE_VALUE
}

func (n *NightscoutSyncSettings) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled
}

type DexcomSyncSettings struct {
	Username string              `json:"username"`
	Region   consts.DexcomRegion `json:"region"`
	Enabled  bool                `json:"enabled"`
}

func (n *DexcomSyncSettings) GetUsername() string {
	if n == nil {
		return ""
	}
	return n.Username
}

func (n *DexcomSyncSettings) GetPasswordGarbage() string {
	if n == nil {
		return ""
	}
	return consts.PASSWORD_GARBAGE_VALUE
}

func (n *DexcomSyncSettings) IsUSARegion() bool {
	if n == nil {
		return false
	}
	return n.Region == consts.DexcomRegionUS
}
func (n *DexcomSyncSettings) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled
}
