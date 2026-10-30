package publicmodels

import "strings"

type NewSiteInfo struct {
	URL             string `json:"url"`
	AdminPassword   string `json:"admin_password"`
	AdminTotpSecret string `json:"admin_totp_secret"`
	PwdNSToken      string `json:"pwd_ns_token"`
	PwdNSURL        string `json:"pwd_ns_url"`
	BackupKey       string `json:"backup_key"`
	// RecoverURL is the page the backup key gets a new admin login on
	RecoverURL string `json:"-"`
}

// SiteID is the start of the backup key, which is its site's ID then a secret
func (n NewSiteInfo) SiteID() string {
	siteID, _, _ := strings.Cut(n.BackupKey, ":")
	return siteID
}

// AllInfo is the credentials to copy, leaving out the phone's when there are none, as on a recovery
func (n NewSiteInfo) AllInfo() string {
	info := "Site: " + n.URL + "\nSite ID: " + n.SiteID() + "\nUsername: admin\nPassword: " + n.AdminPassword + "\nTOTP Secret: " + n.AdminTotpSecret
	if n.PwdNSToken != "" {
		info += "\nPWD NS Token: " + n.PwdNSToken + "\nPWD NS URL: " + n.PwdNSURL
	}
	return info + "\nBackup Key: " + n.BackupKey
}
