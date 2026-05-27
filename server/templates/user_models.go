package templates

import (
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
)

type UserIdentity struct {
	ID              string
	Name            string
	Created         string
	LastSeen        string
	Role            string
	CurrentIdentity bool
	HasTotpSecret   bool
	NightscoutToken string
	Enabled         bool
}

func IdentityModelToUserIdentity(identity *user.Identity, currentIdentityName string, nightscoutToken string) *UserIdentity {
	lastSeen := "Never"
	if identity.LastSeenAt != nil && *identity.LastSeenAt != 0 {
		lastSeen = time.UnixMilli(*identity.LastSeenAt).Format("2006-01-02 15:04:05")
	}

	return &UserIdentity{
		ID:              strconv.Itoa(int(identity.ID)),
		Name:            identity.Name,
		Created:         time.UnixMilli(identity.CreatedAt).Format("2006-01-02 15:04:05"),
		LastSeen:        lastSeen,
		Role:            string(identity.Role),
		CurrentIdentity: identity.Name == currentIdentityName,
		HasTotpSecret:   identity.TotpSecret != "",
		NightscoutToken: nightscoutToken,
		Enabled:         identity.Enabled,
	}
}
