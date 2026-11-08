package controlmodels

import "nighttroop.t1d.nz/utils/consts"

// SiteIdentity is what the control plane shows of one of a site's identities: who it is and when it was used, never
// its password or TOTP secret
type SiteIdentity struct {
	ID                    int64               `json:"id"`
	Name                  string              `json:"name"`
	Role                  consts.IdentityRole `json:"role"`
	Enabled               bool                `json:"enabled"`
	HasTotp               bool                `json:"has_totp"`
	CreatedAt             int64               `json:"created_at"`
	CreatedByIdentityName string              `json:"created_by_identity_name"`
	UpdatedByIdentityName *string             `json:"updated_by_identity_name"`
	LastSeenAt            *int64              `json:"last_seen_at"`
	LastWriteAt           *int64              `json:"last_write_at"`
}

// SiteIdentityCredentials are an identity's new credentials after a control admin resets them, shown only this once:
// a new password, and the Nightscout token made from it for a device, or a new TOTP secret
type SiteIdentityCredentials struct {
	Name       string  `json:"name"`
	Password   *string `json:"password,omitempty"`
	NSToken    *string `json:"ns_token,omitempty"`
	TotpSecret *string `json:"totp_secret,omitempty"`
}

// SiteIdentityAssume is a link a control admin follows, once and within a minute, to sign in to a site as its identity
type SiteIdentityAssume struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// SitePage is a site's own page in control: its details, the bytes its DB's tables take, its identities, and its
// latest events and errors
type SitePage struct {
	Site       *Site           `json:"site"`
	Tables     []*SiteTable    `json:"tables"`
	Identities []*SiteIdentity `json:"identities"`
	Events     []*Event        `json:"events"`
	Errors     []*Error        `json:"errors"`
}

// SiteTable is a table of a site's DB and the bytes it takes in the file, its indexes with it
type SiteTable struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
}

// HasSecret is whether an encrypted secret is set
func HasSecret(encrypted string) bool {
	return encrypted != ""
}
