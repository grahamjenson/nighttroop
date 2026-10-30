package publicmodels

type NewSiteInfo struct {
	URL             string `json:"url"`
	AdminPassword   string `json:"admin_password"`
	AdminTotpSecret string `json:"admin_totp_secret"`
	PwdNSToken      string `json:"pwd_ns_token"`
	PwdNSURL        string `json:"pwd_ns_url"`
	BackupKey       string `json:"backup_key"`
}

func (n NewSiteInfo) AllInfo() string {
	return "Site: " + n.URL + "\nUsername: admin\nPassword: " + n.AdminPassword + "\nTOTP Secret: " + n.AdminTotpSecret + "\nPWD NS Token: " + n.PwdNSToken + "\nPWD NS URL: " + n.PwdNSURL + "\nBackup Key: " + n.BackupKey
}
