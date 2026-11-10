package controlmodels

import "nighttroop.t1d.nz/utils/consts"

// ControlAdmin is what control shows of one of its admins: who it is, its role, and who made and last changed it, never
// its password or TOTP secret
type ControlAdmin struct {
	ID                    int64              `json:"id"`
	Name                  string             `json:"name"`
	Role                  consts.ControlRole `json:"role"`
	CreatedAt             int64              `json:"created_at"`
	CreatedByIdentityName string             `json:"created_by_identity_name"`
	UpdatedByIdentityName *string            `json:"updated_by_identity_name"`
}
