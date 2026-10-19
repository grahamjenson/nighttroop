package sitemodels

import (
	"nighttroop.t1d.nz/utils/consts"
)

type UserIdentity struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Created         string `json:"created"`
	LastSeen        string `json:"last_seen"`
	LastWrite       string `json:"last_write"`
	Role            string `json:"role"`
	Type            string `json:"type"`
	CurrentIdentity bool   `json:"current_identity"`
	HasTotpSecret   bool   `json:"has_totp_secret"`

	Enabled bool `json:"enabled"`

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
