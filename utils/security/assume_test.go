package security_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// A link to sign in as a site's identity names it and the control admin, for its own site only, and only once
func TestAssumeJWTToken(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	token, err := cfg.GenerateAssumeJWTToken("abcdefgh", "admin", "control_admin")
	require.NoError(t, err)

	_, _, err = cfg.VerifyAssumeJWTToken(token, "otherone")
	require.ErrorContains(t, err, "different site")

	name, by, err := cfg.VerifyAssumeJWTToken(token, "abcdefgh")
	require.NoError(t, err)
	require.Equal(t, "admin", name)
	require.Equal(t, "control_admin", by)

	_, _, err = cfg.VerifyAssumeJWTToken(token, "abcdefgh")
	require.ErrorContains(t, err, "has been used")

	// A sign in token is not a link to sign in
	userToken, _, err := cfg.GenerateUserJWTToken(&controldb.SiteIdentity{ID: 1}, "abcdefgh", "admin", string(consts.IdentityRoleAdmin))
	require.NoError(t, err)
	_, _, err = cfg.VerifyAssumeJWTToken(userToken, "abcdefgh")
	require.Error(t, err)
}
