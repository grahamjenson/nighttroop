package sitemodels

import (
	"nighttroop.t1d.nz/utils/authclients"
	"nighttroop.t1d.nz/utils/consts"
)

type UserIdentity struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	CreatedAt       int64  `json:"created_at"`
	LastSeenAt      *int64 `json:"last_seen_at"`
	LastWriteAt     *int64 `json:"last_write_at"`
	Role            string `json:"role"`
	Type            string `json:"type"`
	CurrentIdentity bool   `json:"current_identity"`
	HasTotpSecret   bool   `json:"has_totp_secret"`

	Enabled bool `json:"enabled"`
	// RecentClients are where it recently authenticated from, newest first, kept in memory only
	RecentClients []authclients.Client `json:"recent_clients"`

	// Only set when an identity is created: the credentials are shown once and never stored in the clear
	Password   string  `json:"password,omitempty"`
	TotpSecret *string `json:"totp_secret,omitempty"`
	NSToken    *string `json:"ns_token,omitempty"`
	SiteURL    string  `json:"-"`
}

type RoleOption struct {
	Label string
	Value consts.IdentityRole
}
