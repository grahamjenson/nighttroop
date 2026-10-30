package controlmodels

import (
	"time"

	"nighttroop.t1d.nz/utils/usage"
)

// Usage is how many use the server and how much, and its resources, at Now in milliseconds. The active counts are
// from the control DB, so they last across restarts; the rest is counted since the server started.
type Usage struct {
	Now int64 `json:"now"`
	// ActiveSitesDay and ActiveSitesWeek are the live sites seen in the last day and week, and likewise identities
	ActiveSitesDay       int64          `json:"active_sites_day"`
	ActiveSitesWeek      int64          `json:"active_sites_week"`
	ActiveIdentitiesDay  int64          `json:"active_identities_day"`
	ActiveIdentitiesWeek int64          `json:"active_identities_week"`
	SitesByState         map[string]int64 `json:"sites_by_state"`
	SinceStart           usage.Snapshot `json:"since_start"`
	Server               usage.Server   `json:"server"`
}

// NowTime is Now as a time
func (u Usage) NowTime() time.Time {
	return time.UnixMilli(u.Now)
}
