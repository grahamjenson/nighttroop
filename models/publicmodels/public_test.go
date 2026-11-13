package publicmodels

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The copy has every credential a site is created with, and a recovery's leaves out the devices', which it does not change
func TestAllInfo(t *testing.T) {
	created := NewSiteInfo{URL: "https://abcdefgh.example", AdminPassword: "adminpass", AdminTotpSecret: "TOTPSECRET",
		PwdNSToken: "phonetoken", PwdNSURL: "https://abcdefgh.ns.example", ReadOnlyNSToken: "readtoken", BackupKey: "abcdefgh:backup"}
	recovered := NewSiteInfo{URL: created.URL, AdminPassword: "newpass", AdminTotpSecret: "NEWSECRET", BackupKey: "abcdefgh:newbackup"}
	tests := []struct {
		name      string
		info      NewSiteInfo
		has       []string
		hasNoneOf []string
	}{
		{name: "a new site", info: created, has: []string{"Nighttroop site: https://abcdefgh.example", "Site ID: abcdefgh",
			"Backup key: abcdefgh:backup", "Username: admin", "Password: adminpass", "TOTP secret: TOTPSECRET", "pwd_phone",
			"Nightscout URL: https://abcdefgh.ns.example", "Nightscout token: phonetoken", "read_only", "Nightscout token: readtoken"}},
		{name: "a recovered site", info: recovered, has: []string{"Site ID: abcdefgh", "Backup key: abcdefgh:newbackup",
			"Password: newpass", "TOTP secret: NEWSECRET"}, hasNoneOf: []string{"Nightscout URL", "Nightscout token", "pwd_phone", "read_only"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.info.AllInfo()
			for _, want := range tt.has {
				require.Contains(t, info, want)
			}
			for _, unwanted := range tt.hasNoneOf {
				require.NotContains(t, info, unwanted)
			}
		})
	}
}
