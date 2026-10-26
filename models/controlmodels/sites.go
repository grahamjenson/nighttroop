package controlmodels

import "time"

// Site is what the control API shows of a site: its state and activity, not its settings or data
type Site struct {
	ID                string `json:"id"`
	Batch             string `json:"batch"`
	State             string `json:"state"`
	MaxDbSizeKbLimit  *int64 `json:"max_db_size_kb_limit"`
	EstimatedDbSizeKb *int64 `json:"estimated_db_size_kb"`
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
