package publicmodels

import (
	"strings"

	"nighttroop.t1d.nz/utils/consts"
)

type NewSiteInfo struct {
	URL             string `json:"url"`
	AdminPassword   string `json:"admin_password"`
	AdminTotpSecret string `json:"admin_totp_secret"`
	PwdNSToken      string `json:"pwd_ns_token"`
	PwdNSURL        string `json:"pwd_ns_url"`
	ReadOnlyNSToken string `json:"read_only_ns_token"`
	BackupKey       string `json:"backup_key"`
	// RecoverURL is the page the backup key gets a new admin login on
	RecoverURL string `json:"-"`
}

// SiteID is the start of the backup key, which is its site's ID then a secret
func (n NewSiteInfo) SiteID() string {
	siteID, _, _ := strings.Cut(n.BackupKey, ":")
	return siteID
}

// AllInfo is the credentials to copy: the site's, then each of its first identities', leaving out the devices' when
// there are none, as on a recovery
func (n NewSiteInfo) AllInfo() string {
	info := "Nighttroop site: " + n.URL + "\nSite ID: " + n.SiteID() + "\nBackup key: " + n.BackupKey +
		"\n\nAdmin login\nUsername: " + consts.FIRST_ADMIN_NAME + "\nPassword: " + n.AdminPassword + "\nTOTP secret: " + n.AdminTotpSecret
	if n.PwdNSToken != "" {
		info += "\n\nPhone setup (AAPS / Trio), the " + consts.FIRST_PWD_NAME + " device\nNS URL: " + n.PwdNSURL + "\nNS token: " + n.PwdNSToken
	}
	if n.ReadOnlyNSToken != "" {
		info += "\n\nRead only, the " + consts.FIRST_READ_ONLY_NAME + " device\nNS URL: " + n.PwdNSURL + "\nNS token: " + n.ReadOnlyNSToken
	}
	return info
}
