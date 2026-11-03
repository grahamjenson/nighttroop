package db_test

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A staging server starts with its configured admin instead of being set up, and keeps it when it starts again
func TestSeedStagingAdmin(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.StagingAdminPassword, cfg.StagingAdminTOTPSecret = "a long staging password", "jbswy3dpehpk3pxp"
	controlDB, err := db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	require.NotEmpty(t, controlDB.SetupCode(), "a new server is set up with its code")

	for range 2 {
		require.NoError(t, controlDB.SeedStagingAdmin(t.Context()))
		require.True(t, controlDB.SetupDone())
	}

	admin, err := controlDB.Queries.GetControlIdentity(t.Context(), controldb.GetControlIdentityParams{Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	require.Equal(t, consts.ControlRoleAdmin, admin.Role)
	require.True(t, cfg.Security.VerifyPassword("a long staging password", admin.PasswordHash))
	require.False(t, cfg.Security.VerifyPassword("password", admin.PasswordHash))
	// Its TOTP is the configured secret, so an authenticator app given it signs in
	secret, err := cfg.Security.Decrypt(admin.TotpSecretEnc)
	require.NoError(t, err)
	code, err := totp.GenerateCode("JBSWY3DPEHPK3PXP", time.Now())
	require.NoError(t, err)
	require.True(t, cfg.Security.TOTPValidate(code, secret))
}
