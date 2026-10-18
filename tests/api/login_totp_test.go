package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// A pre-auth cookie from one site must not log in to the same identity name on another site
func TestPreAuthTokenIsBoundToSite(t *testing.T) {
	config, controlDB, victim, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	attacker, err := controlDB.CreateSite()
	require.NoError(t, err)
	router := server.GetMainHandler(controlDB, config)

	form := url.Values{"name": {consts.FIRST_ADMIN_NAME}, "password": {attacker.AdminPassword}}
	login := httptest.NewRequest(http.MethodPost, "http://"+attacker.SiteDB.Subdomain+".localhost/login", strings.NewReader(form.Encode()))
	login.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	login.Header.Set("HX-Request", "true")
	loginRes := httptest.NewRecorder()
	router.ServeHTTP(loginRes, login)
	var preAuth *http.Cookie
	for _, c := range loginRes.Result().Cookies() {
		if c.Name == security.PRE_AUTH_TOKEN_COOKIE {
			preAuth = c
		}
	}
	require.NotNil(t, preAuth, "attacker should get a pre-auth cookie from their own site")

	totpForm := url.Values{"totp": {*config.Security.DevelopmentDangerousHardcodedAdminTOTPCode}}
	totp := httptest.NewRequest(http.MethodPost, "http://"+victim.SiteDB.Subdomain+".localhost/login/totp", strings.NewReader(totpForm.Encode()))
	totp.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	totp.Header.Set("HX-Request", "true")
	totp.AddCookie(preAuth)
	totpRes := httptest.NewRecorder()
	router.ServeHTTP(totpRes, totp)

	for _, c := range totpRes.Result().Cookies() {
		require.NotEqual(t, security.SITE_TOKEN_COOKIE, c.Name, "victim site issued a login cookie")
	}
}
