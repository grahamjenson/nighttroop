package nighttroop

import (
	"nighttroop.t1d.nz/utils/consts"
)

type SiteSyncSource struct {
	ID            int64
	SiteID        string
	Type          consts.SyncSourceType
	Region        string
	UrlOrUsername string
	Password      string
	Enabled       bool
}

func (n *SiteSyncSource) GetDexcomRegion() consts.DexcomRegion {
	if n == nil {
		return consts.DexcomRegionOther
	}
	return consts.NewDexcomRegionFromString(n.Region)
}

func (n *SiteSyncSource) IsUSARegion() bool {
	if n == nil {
		return false
	}
	return n.GetDexcomRegion() == consts.DexcomRegionUS
}

func (n *SiteSyncSource) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled
}
func (n *SiteSyncSource) GetUsername() string {
	if n == nil {
		return ""
	}
	return n.UrlOrUsername
}
func (n *SiteSyncSource) GetURL() string {
	if n == nil {
		return ""
	}
	return n.UrlOrUsername
}

func (n *SiteSyncSource) GetPassword() string {
	if n == nil {
		return ""
	}
	return n.Password
}

func (n *SiteSyncSource) GetPasswordGarbage() string {
	if n == nil {
		return ""
	}
	return consts.PASSWORD_GARBAGE_VALUE
}
