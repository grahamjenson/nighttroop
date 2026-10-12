package site

import (
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/consts"
)

type UserIdentity struct {
	ID              string
	Name            string
	Created         string
	LastSeen        string
	LastWrite       string
	Role            string
	Type            string
	CurrentIdentity bool
	HasTotpSecret   bool

	Enabled bool
}

type RoleOption struct {
	Label string
	Value consts.IdentityRole
}

func IdentityModelToUserIdentity(identity *control.SiteIdentity, currentIdentityName string) *UserIdentity {
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
		HasTotpSecret:   identity.TotpSecret != "",
		Enabled:         identity.Enabled,
	}
}
