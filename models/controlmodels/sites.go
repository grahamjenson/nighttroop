package controlmodels

import (
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
)

// Site is what the control API shows of a site: its state and activity, not its settings or data
type Site struct {
	ID                string `json:"id"`
	Batch             string `json:"batch"`
	State             string `json:"state"`
	MaxDbSizeKbLimit  *int64 `json:"max_db_size_kb_limit"`
	EstimatedDbSizeKb *int64 `json:"estimated_db_size_kb"`
	// AbuseScore is how much the site may abuse the server, from 0 to 1, as the resource check last scored it
	AbuseScore float64 `json:"abuse_score"`
	CreatedAt         int64  `json:"created_at"`
	LastSeenAt        *int64 `json:"last_seen_at"`
	LastWriteAt       *int64 `json:"last_write_at"`
	DisabledAt        *int64 `json:"disabled_at"`
	ArchivedAt        *int64 `json:"archived_at"`
	DeletedAt         *int64 `json:"deleted_at"`
}

// Inactive is whether the site has not been written to, or not been seen, in the days before now; one never written
// to or seen counts from when it was made
func (s *Site) Inactive(days int, now time.Time) bool {
	cutoff := now.AddDate(0, 0, -days).UnixMilli()
	lastWrite, lastSeen := s.CreatedAt, s.CreatedAt
	if s.LastWriteAt != nil {
		lastWrite = *s.LastWriteAt
	}
	if s.LastSeenAt != nil {
		lastSeen = *s.LastSeenAt
	}
	return lastWrite <= cutoff || lastSeen <= cutoff
}

// Oversized is whether the site's DB, while it is kept here, is over its size limit, so it can be truncated
func (s *Site) Oversized() bool {
	kept := s.State == "live" || s.State == "disabled"
	return kept && s.EstimatedDbSizeKb != nil && s.MaxDbSizeKbLimit != nil && *s.EstimatedDbSizeKb > *s.MaxDbSizeKbLimit
}

// AbuseLevel is the level of the site's abuse score
func (s *Site) AbuseLevel() controldb.AbuseLevel { return controldb.AbuseLevelOf(s.AbuseScore) }
