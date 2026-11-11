package nightscoutapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// A legacy token is refused as a wrong secret is, with a password checked, whether its name exists, is an admin or has
// TOTP, so neither the answer nor how long it takes tells an identity's name or what it is
func TestLegacyTokenRefusalsAreLikeWrongSecret(t *testing.T) {
	cfg := testConfig
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	subdomain := site.SiteDB.Subdomain
	// The phone has TOTP, and the admin none, so each is refused for what it is
	setTOTP := func(name string, totpSecretEnc string) {
		identity, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: subdomain, Name: name})
		require.NoError(t, err)
		_, err = controlDB.Queries.UpdateSiteIdentityTOTP(t.Context(), controldb.UpdateSiteIdentityTOTPParams{TotpSecretEnc: totpSecretEnc, SiteID: subdomain, ID: identity.ID})
		require.NoError(t, err)
	}
	setTOTP(consts.FIRST_PWD_NAME, "secret")
	setTOTP(consts.FIRST_ADMIN_NAME, "")
	router := NewNightscoutRouter(controlDB, cfg)

	tests := []struct {
		name     string
		identity string
		role     consts.IdentityRole
	}{
		{"wrong secret", consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead},
		{"no such name", "nosuchname", consts.IdentityRoleDeviceRead},
		{"wrong role", consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleUser},
		{"admin", consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin},
		{"totp", consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone},
	}
	var wantStatus int
	var wantBody string
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := cfg.Security.CreateNightscoutTokenForIdentity(subdomain, tt.identity, string(tt.role), "wrong")
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/entries.json?token="+token, nil)
			// Each from its own IP, as an IP has only a few unauthenticated requests
			ip := fmt.Sprintf("10.1.0.%d", i+1)
			ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, nil, nil, subdomain, ip, httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			rec := httptest.NewRecorder()
			before := cfg.Security.PasswordsVerified()
			router.ServeHTTP(rec, req.WithContext(ctx))
			require.Equal(t, int64(1), cfg.Security.PasswordsVerified()-before, "passwords checked")

			if i == 0 {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
				wantStatus, wantBody = rec.Code, rec.Body.String()
			}
			require.Equal(t, wantStatus, rec.Code)
			require.Equal(t, wantBody, rec.Body.String())
		})
	}
}
