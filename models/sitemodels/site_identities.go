package sitemodels

import (
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
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

func IdentityModelToUserIdentity(identity *controldb.SiteIdentity, currentIdentityName string) *UserIdentity {
	lastSeen := "Never"
	if identity.LastSeenAt != nil && *identity.LastSeenAt != 0 {
		lastSeen = time.UnixMilli(*identity.LastSeenAt).Format("2006-01-02 15:04:05")
	}

	lastWrite := "Never"
	if identity.LastWriteAt != nil && *identity.LastWriteAt != 0 {
		lastWrite = time.UnixMilli(*identity.LastWriteAt).Format("2006-01-02 15:04:05")
	}
	return &UserIdentity{
		ID:        strconv.Itoa(int(identity.ID)),
		Name:      identity.Name,
		Created:   time.UnixMilli(identity.CreatedAt).Format("2006-01-02 15:04:05"),
		LastSeen:  lastSeen,
		LastWrite: lastWrite,
		Role:      string(identity.Role),

		CurrentIdentity: identity.Name == currentIdentityName,
		HasTotpSecret:   identity.TotpSecretEnc != "",
		Enabled:         identity.Enabled,
	}
}
