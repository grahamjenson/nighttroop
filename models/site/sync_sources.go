package site

import (
	"nighttroop.t1d.nz/utils/consts"
)

type SyncSource struct {
	ID            int64
	SiteID        string
	Type          consts.SyncSourceType
	Region        string
	UrlOrUsername string
	Password      string
	Enabled       bool
}

func (n *SyncSource) GetDexcomRegion() consts.DexcomRegion {
	if n == nil {
		return consts.DexcomRegionOther
	}
	return consts.NewDexcomRegionFromString(n.Region)
}

func (n *SyncSource) IsUSARegion() bool {
	if n == nil {
		return false
	}
	return n.GetDexcomRegion() == consts.DexcomRegionUS
}

func (n *SyncSource) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled
}
func (n *SyncSource) GetUsername() string {
	if n == nil {
		return ""
	}
	return n.UrlOrUsername
}
func (n *SyncSource) GetURL() string {
	if n == nil {
		return ""
	}
	return n.UrlOrUsername
}

func (n *SyncSource) GetPassword() string {
	if n == nil {
		return ""
	}
	return n.Password
}

func (n *SyncSource) GetPasswordGarbage() string {
	if n == nil {
		return ""
	}
	return consts.PASSWORD_GARBAGE_VALUE
}
