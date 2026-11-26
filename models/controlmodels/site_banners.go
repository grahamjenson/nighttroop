package controlmodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/controldb"
)

// The most a control admin can make of site banners, so they cannot fill the table or a site's pages
const (
	// MaxSiteBanners is the most a server has at once, MaxSiteBannerDays the longest one lasts, and
	// MaxSiteBannerSites the most sites one can list
	MaxSiteBanners     = 20
	MaxSiteBannerDays  = 90
	MaxSiteBannerSites = 100
)

// SiteBannerLevels are how a banner is shown, the first as one made without a level
var SiteBannerLevels = []string{"info", "warning"}

// SiteBannerMessageLimit is what a banner's message takes, as its table checks
var SiteBannerMessageLimit, _ = controldb.ColumnLimitOf("site_banners", "Message")

// SiteBanner is a message a control admin shows at the top of sites' pages until it expires
type SiteBanner struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Level   string `json:"level"`
	// SiteIDs are the sites it is shown on, and none is every site
	SiteIDs               []string `json:"site_ids"`
	ExpiresAt             int64    `json:"expires_at"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	CreatedAt             int64    `json:"created_at"`
}

func SiteBannerFromDB(row *controldb.SiteBanner) *SiteBanner {
	banner := &SiteBanner{ID: row.ID, Message: row.Message, Level: row.Level, SiteIDs: []string{}, ExpiresAt: row.ExpiresAt,
		CreatedByIdentityName: row.CreatedByIdentityName, CreatedAt: row.CreatedAt}
	_ = json.Unmarshal([]byte(row.SiteIdsJson), &banner.SiteIDs)
	return banner
}
