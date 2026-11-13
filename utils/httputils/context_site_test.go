package httputils

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
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

// Without a signed in identity, as for a live update sent to a site's sockets, the site's own settings are in effect,
// and that is not an error
func TestGetSettingsWithoutAnIdentityIsTheSites(t *testing.T) {
	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	defer slog.SetDefault(previous)

	site := &SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh", Units: new(consts.UnitsMGDL)}}
	require.Equal(t, consts.UnitsMGDL, *site.GetSettings().GetUnits())
	require.NotContains(t, logged.String(), "ERROR")
}

// A JWT issued in the same millisecond as its identity's credentials changed is refused too, as the reset may have come
// after it: a sign in, a request and a reset can all be within one millisecond
func TestCheckIssuedAfterCredentialsChanged(t *testing.T) {
	identity := controldb.SiteIdentity{CredentialsChangedAt: 1000}
	tests := []struct {
		name       string
		issuedAtMs int64
		refused    bool
	}{
		{"issued before", 999, true},
		{"issued in the same millisecond", 1000, true},
		{"issued after", 1001, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckIssuedAfterCredentialsChanged(identity, tt.issuedAtMs)
			require.Equal(t, tt.refused, err != nil)
		})
	}
}
