package sitemodels

import "nighttroop.t1d.nz/utils/consts"

// PushCredential is a credential pushes are sent with to apps on the site's devices: an Apple APNs key (.p8) or a
// Firebase service account. Its secret is only sent to add one, and never returned. Apps are the apps it pushes to,
// and BundleIDs, when there are any, the only bundle IDs or package names it pushes to.
type PushCredential struct {
	ID                    int64               `json:"id"`
	Name                  string              `json:"name"`
	Provider              consts.PushProvider `json:"provider"`
	TeamOrProjectID       string              `json:"team_or_project_id"`
	KeyID                 *string             `json:"key_id"`
	Secret                *string             `json:"secret,omitempty"`
	Apps                  []string            `json:"apps"`
	BundleIDs             []string            `json:"bundle_ids"`
	CreatedAt             int64               `json:"created_at"`
	UpdatedAt             int64               `json:"updated_at"`
	CreatedByIdentityName string              `json:"created_by_identity_name"`
}
