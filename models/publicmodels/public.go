package publicmodels

type NewSiteInfo struct {
	URL             string
	AdminPassword   string
	AdminTotpSecret string
	PwdNSToken      string
	PwdNSURL        string
	BackupKey       string
}

func (n NewSiteInfo) AllInfo() string {
	return "Site: " + n.URL + "\nUsername: admin\nPassword: " + n.AdminPassword + "\nTOTP Secret: " + n.AdminTotpSecret + "\nPWD NS Token: " + n.PwdNSToken + "\nPWD NS URL: " + n.PwdNSURL + "\nBackup Key: " + n.BackupKey
}
