package controlapi

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// A TOTP reset signs out a control admin signed in the same millisecond before it, and not one signed in the same
// millisecond after it, so however fast a sign in and a reset are, the sign in from before it stops working
func TestCredentialsResetSignsOutASignInTheSameMillisecondBefore(t *testing.T) {
	controlDB, err := db.NewControlDB(t.Context(), testConfig)
	require.NoError(t, err)
	defer controlDB.Close()
	_, err = controlDB.CreateControlIdentity(t.Context(), "samems", consts.ControlRoleAdmin, nil)
	require.NoError(t, err)
	getAdmin := func() *controldb.ControlIdentity {
		admin, err := controlDB.Queries.GetControlIdentity(t.Context(), controldb.GetControlIdentityParams{Name: "samems"})
		require.NoError(t, err)
		return admin
	}
	// As the login makes it, for the admin as it is then
	signIn := func() string {
		admin := getAdmin()
		token, _, err := testConfig.Security.GenerateControlJWTToken(admin, admin.Name, string(admin.Role))
		require.NoError(t, err)
		return token
	}
	signedIn := func(token string) bool {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		_, err := tryLoginWithJWTToken(controlDB, testConfig, r)
		return err == nil
	}

	// Tried until a sign in, the reset and a sign in after it all land in one millisecond
	for attempt := 0; attempt < 1000; attempt++ {
		start := time.Now().UnixMilli()
		before := signIn()
		admin := getAdmin()
		_, err := controlDB.ResetControlIdentityTOTP(t.Context(), admin, admin)
		require.NoError(t, err)
		after := signIn()
		if time.Now().UnixMilli() != start {
			continue
		}
		require.False(t, signedIn(before), "a sign in the same millisecond before the reset still works")
		require.True(t, signedIn(after), "a sign in the same millisecond after the reset is refused")
		return
	}
	t.Fatal("never signed in and reset in one millisecond")
}
