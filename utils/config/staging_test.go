package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// Staging runs like production, always shows the test site banner, and will not start without its admin
func TestStagingConfig(t *testing.T) {
	cfg, err := readServerConfig(consts.Staging, map[string]string{"HOST_PROTOCOL": "https"})
	require.NoError(t, err)
	require.True(t, cfg.IsStaging())
	require.False(t, cfg.IsDevOrTest())
	require.True(t, cfg.TestSite)
	require.ErrorContains(t, cfg.ValidateClean(), "STAGING_ADMIN_PASSWORD")

	cfg, err = readServerConfig(consts.Staging, map[string]string{"HOST_PROTOCOL": "https",
		"STAGING_ADMIN_PASSWORD": "a long staging password", "STAGING_ADMIN_TOTP_SECRET": "JBSWY3DPEHPK3PXP"})
	require.NoError(t, err)
	require.True(t, cfg.TestSite)
	require.NoError(t, cfg.ValidateClean())

	// TEST_SITE is no longer a setting, so a .env file that sets it is refused as unknown
	dotenv := map[string]string{"HOST_PROTOCOL": "https", "TEST_SITE": "false"}
	_, err = readServerConfig(consts.Staging, dotenv)
	require.NoError(t, err)
	require.Error(t, rejectUnknownVariables(dotenv))

	// Production has no staging admin, and needs none
	cfg, err = readServerConfig(consts.Production, map[string]string{"HOST_PROTOCOL": "https"})
	require.NoError(t, err)
	require.False(t, cfg.TestSite)
	require.NoError(t, cfg.ValidateClean())
}

// Staging needs its admin's credentials, a long password and a TOTP secret an authenticator app takes
func TestValidateStagingAdmin(t *testing.T) {
	tests := []struct {
		name, password, secret string
		ok                     bool
	}{
		{"both set", "a long staging password", "JBSWY3DPEHPK3PXP", true},
		{"a lower case and padded secret", "a long staging password", "jbswy3dpehpk3pxpjbswy3dpehpk3pxp====", true},
		{"no password", "", "JBSWY3DPEHPK3PXP", false},
		{"a short password", "short", "JBSWY3DPEHPK3PXP", false},
		{"no secret", "a long staging password", "", false},
		{"a short secret", "a long staging password", "JBSWY3DP", false},
		{"a secret that is not base32", "a long staging password", "not base32 at all 189!", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.ok, validateStagingAdmin(tt.password, tt.secret) == nil)
		})
	}
}
