package controlmodels

import "nighttroop.t1d.nz/utils/consts"

// SyncSource is a site's sync source and its reports; never its credentials
type SyncSource struct {
	SiteID                     string                `json:"site_id"`
	Type                       consts.SyncSourceType `json:"type"`
	Region                     *string               `json:"region"`
	Enabled                    bool                  `json:"enabled"`
	CreatedAt                  int64                 `json:"created_at"`
	UpdatedAt                  int64                 `json:"updated_at"`
	FirstSuccessfulSyncTime    *int64                `json:"first_successful_sync_time"`
	SyncIntervalSecondsReason  string                `json:"sync_interval_seconds_reason"`
	SyncIntervalSeconds        int64                 `json:"sync_interval_seconds"`
	LastAttemptedSync          *int64                `json:"last_attempted_sync"`
	LastSuccessfulSync         *int64                `json:"last_successful_sync"`
	LastFailedSync             *int64                `json:"last_failed_sync"`
	TotalSyncs                 *int64                `json:"total_syncs"`
	TotalSuccessfulSyncs       *int64                `json:"total_successful_syncs"`
	TotalFailedSyncs           *int64                `json:"total_failed_syncs"`
	ConsecutiveSuccessfulSyncs *int64                `json:"consecutive_successful_syncs"`
	ConsecutiveFailedSyncs     *int64                `json:"consecutive_failed_syncs"`
	PreviousSyncErrors         []string              `json:"previous_sync_errors"`
	TotalEma                   *float64              `json:"total_ema"`
	SuccessEma                 *float64              `json:"success_ema"`
	FailureEma                 *float64              `json:"failure_ema"`
	LastSyncDuration           *int64                `json:"last_sync_duration"`
}

// ID is the site and type that pick out a sync source, as site::type
func (s *SyncSource) ID() string {
	return s.SiteID + "::" + string(s.Type)
}
