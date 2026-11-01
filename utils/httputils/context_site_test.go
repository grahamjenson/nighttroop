package httputils

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
)

// Settings asked for before the identity is known are logged as an error without the site's or identity's secrets,
// as a site's logs are shown to its admin and kept in the logs DB
func TestGetSettingsLogsNoSecrets(t *testing.T) {
	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	defer slog.SetDefault(previous)

	site := &SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh", BackupKeyHash: "backup-key-hash"}}
	require.NotNil(t, site.GetSettings())
	identity := &SiteContext{SiteID: "abcdefgh", IdentityModel: &controldb.SiteIdentity{PasswordHash: "password-hash", TotpSecretEnc: "totp-secret"}}
	require.NotNil(t, identity.GetSettings())

	require.Contains(t, logged.String(), "abcdefgh")
	for _, secret := range []string{"backup-key-hash", "password-hash", "totp-secret"} {
		require.NotContains(t, logged.String(), secret)
	}
}
