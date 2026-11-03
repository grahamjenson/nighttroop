package controlmodels

import (
	"time"

	"nighttroop.t1d.nz/data/sql/sqltiming"
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
	// SitesCreatedToday, Week and Month are the sites created since midnight, Monday and the 1st, in UTC
	SitesCreatedToday int64 `json:"sites_created_today"`
	SitesCreatedWeek  int64 `json:"sites_created_week"`
	SitesCreatedMonth int64 `json:"sites_created_month"`
	SinceStart           usage.Snapshot `json:"since_start"`
	Server               usage.Server   `json:"server"`
	// ResourceReadings are the checks of the server's resources since ResourcesSince, oldest first, kept while it
	// runs, and the levels memory and disk warn at
	ResourceReadings   []ResourceReading `json:"resource_readings"`
	ResourcesSince     int64             `json:"resources_since"`
	MemoryAlertPercent int               `json:"memory_alert_percent"`
	DiskAlertPercent   int               `json:"disk_alert_percent"`

	// SlowestQueries are the DB queries whose sampled runs took the longest in all since the server started, of the
	// QuerySamplePercent kept
	SlowestQueries     []sqltiming.QueryStat `json:"slowest_queries"`
	QuerySamplePercent int                   `json:"query_sample_percent"`
}

// ResourceReading is a check of the server's resources: its mean CPU since the check before and the longest it was at
// 100%, and the memory and disk used, unset where they could not be read
type ResourceReading struct {
	At               int64    `json:"at"`
	CPUPercent       float64  `json:"cpu_percent"`
	CPUPeggedSeconds int64    `json:"cpu_pegged_seconds"`
	MemoryPercent    *float64 `json:"memory_percent"`
	DiskPercent      *float64 `json:"disk_percent"`
}

// NowTime is Now as a time
func (u Usage) NowTime() time.Time {
	return time.UnixMilli(u.Now)
}
