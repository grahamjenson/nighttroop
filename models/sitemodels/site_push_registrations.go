package sitemodels

import "nighttroop.t1d.nz/utils/consts"

// PushRegistration is an app on a device that the signed in identity has pushes sent to. Its token addresses the app on
// that device; it is only sent to register, and never returned. Production is whether an APNs token is from a TestFlight
// or App Store build, rather than a development one.
type PushRegistration struct {
	ID           int64               `json:"id"`
	App          consts.PushApp      `json:"app"`
	Provider     consts.PushProvider `json:"provider"`
	BundleID     string              `json:"bundle_id"`
	Production   bool                `json:"production"`
	Token        *string             `json:"token,omitempty"`
	CreatedAt    int64               `json:"created_at"`
	UpdatedAt    int64               `json:"updated_at"`
	LastPushedAt *int64              `json:"last_pushed_at"`
	LastError    *string             `json:"last_error"`
}
