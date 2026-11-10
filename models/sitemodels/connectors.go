package sitemodels

import (
	"nighttroop.t1d.nz/utils/consts"
)

type Connector struct {
	SiteID        string               `json:"site_id"`
	Type          consts.ConnectorType `json:"type"`
	Region        *string              `json:"region"`
	UrlOrUsername *string              `json:"url_or_username"`
	Password      *string              `json:"password"`
	Enabled       bool                 `json:"enabled"`
	// PreviousSyncErrors are from the last failed sync, shown to the user; cleared when the source is saved
	PreviousSyncErrors []string `json:"previous_sync_errors,omitempty"`

	// How it has synced, read only: saving a source never sets these
	SyncIntervalSeconds       int64   `json:"sync_interval_seconds"`
	SyncIntervalSecondsReason string  `json:"sync_interval_seconds_reason"`
	LastAttemptedSync         *int64  `json:"last_attempted_sync"`
	LastSuccessfulSync        *int64  `json:"last_successful_sync"`
	LastFailedSync            *int64  `json:"last_failed_sync"`
	LastSyncDuration          *int64  `json:"last_sync_duration"`
	TotalSuccessfulSyncs      *int64  `json:"total_successful_syncs"`
	TotalFailedSyncs          *int64  `json:"total_failed_syncs"`
	ConsecutiveFailedSyncs    *int64  `json:"consecutive_failed_syncs"`
	LastBackfillFinishedAt    *int64  `json:"last_backfill_finished_at"`
	// LastBackfillFrom is the oldest time its last backfill fetched from, how far back its backfills have gone, as
	// each fetches from before the site's oldest reading
	LastBackfillFrom *int64 `json:"last_backfill_from"`
	LastBackfillDuration      *int64  `json:"last_backfill_duration"`
	LastBackfillRecords       *int64  `json:"last_backfill_records"`
	LastBackfillError         *string `json:"last_backfill_error"`
	TotalBackfills            int64   `json:"total_backfills"`
	// TotalRecords are the records its syncs and backfills have added in all, and LastSyncRecords those its last
	// successful sync added
	TotalRecords    int64  `json:"total_records"`
	LastSyncRecords *int64 `json:"last_sync_records"`
	// RecentEvents and RecentErrors are the newest of the source's events and errors from the logs DB
	RecentEvents []*SiteEvent `json:"recent_events,omitempty"`
	RecentErrors []*SiteError `json:"recent_errors,omitempty"`
}

func (n *Connector) GetDexcomRegion() consts.DexcomRegion {
	if n == nil || n.Region == nil {
		return consts.DexcomRegionOther
	}
	return consts.NewDexcomRegionFromString(*n.Region)
}

func (n *Connector) IsUSARegion() bool {
	if n == nil {
		return false
	}
	return n.GetDexcomRegion() == consts.DexcomRegionUS
}

// GetLibreRegion is the source's LibreLinkUp region, or Global when it has none
func (n *Connector) GetLibreRegion() string {
	if n == nil || n.Region == nil || *n.Region == "" {
		return consts.LibreRegionGlobal
	}
	return *n.Region
}

func (n *Connector) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled
}
func (n *Connector) GetUsername() string {
	if n == nil || n.UrlOrUsername == nil {
		return ""
	}
	return *n.UrlOrUsername
}
func (n *Connector) GetURL() string {
	if n == nil || n.UrlOrUsername == nil {
		return ""
	}
	return *n.UrlOrUsername
}

func (n *Connector) GetPassword() string {
	if n == nil || n.Password == nil {
		return ""
	}
	return *n.Password
}

func (n *Connector) GetPasswordGarbage() string {
	if n == nil {
		return ""
	}
	return consts.PASSWORD_GARBAGE_VALUE
}
