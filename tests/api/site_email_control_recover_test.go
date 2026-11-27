package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// This file is after nightscout_test.go by name, as the test config is loaded once, with the sign in limits TestNS sets.
//
// A control admin who forgot their password gives their verified email, and the link sent to it gives them a new one,
// once. Control answers the same for an email no admin has verified, and sends that nothing.
func TestControlAdminRecoversTheirPasswordByTheirVerifiedEmail(t *testing.T) {
	type email struct {
		Subject, Text string
		To            []string
	}
	var sent []email
	var lock sync.Mutex
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e email
		require.NoError(t, json.NewDecoder(r.Body).Decode(&e))
		lock.Lock()
		sent = append(sent, e)
		lock.Unlock()
	}))
	t.Cleanup(resend.Close)
	emails := func() []email {
		lock.Lock()
		defer lock.Unlock()
		return append([]email{}, sent...)
	}

	cfg, controlDB, _ := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "", "", "" })
	router := server.GetMainHandler(controlDB, cfg)
	token := controlLogin(t, router)
	call := func(method, path, token, body string) (int, map[string]any) {
		req := httptest.NewRequest(method, "http://control.localhost"+path, strings.NewReader(body))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		var answer map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &answer)
		return rr.Code, answer
	}
	admin, err := controlDB.Queries.GetControlIdentity(t.Context(), controldb.GetControlIdentityParams{Name: "admin"})
	require.NoError(t, err)
	emailEnc, err := cfg.Security.EncryptFor(security.PurposeControlEmail, "", "sam@example.com")
	require.NoError(t, err)
	require.NoError(t, controlDB.Queries.UpdateControlIdentityEmail(t.Context(), controldb.UpdateControlIdentityEmailParams{ID: admin.ID, EmailEnc: &emailEnc}))

	// A server that sends no email has no recovery
	status, _ := call("POST", "/login/recover", "", `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusBadRequest, status)
	cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "re_key", "Nighttroop <hello@nt.example.com>", resend.URL

	// An email not verified, or no admin's, or not as it was typed, is answered the same, and sent nothing
	for _, other := range []string{"sam@example.com", "pat@example.com", "Sam@example.com"} {
		status, _ = call("POST", "/login/recover", "", `{"email": "`+other+`"}`)
		require.Equal(t, http.StatusOK, status, other)
	}
	status, _ = call("POST", "/login/recover", "", `{"email": "sam"}`)
	require.Equal(t, http.StatusBadRequest, status)

	// Verified, it is sent a link, once an hour
	verified, err := controlDB.Queries.VerifyControlIdentityEmail(t.Context(), controldb.VerifyControlIdentityEmailParams{ID: admin.ID, EmailEnc: &emailEnc})
	require.NoError(t, err)
	require.EqualValues(t, 1, verified)
	status, _ = call("POST", "/login/recover", "", `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusOK, status)
	require.Eventually(t, func() bool { return len(emails()) == 1 }, 5*time.Second, 10*time.Millisecond)
	require.Equal(t, []string{"sam@example.com"}, emails()[0].To)
	require.Equal(t, "Recover your Nighttroop control password", emails()[0].Subject)
	link := regexp.MustCompile(`/login/recover/password\?id=([0-9]+)&key=([0-9a-f]{64})`).FindStringSubmatch(emails()[0].Text)
	require.NotNil(t, link, emails()[0].Text)
	status, _ = call("POST", "/login/recover", "", `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusOK, status)

	// Opening the link changes nothing, as mail scanners open links; another key or admin gets no password
	req := httptest.NewRequest("GET", "http://control.localhost"+link[0], nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), url.QueryEscape(link[2]))
	for _, wrong := range [][2]string{{"0", link[2]}, {link[1], strings.Repeat("0", 64)}} {
		status, _ = call("POST", "/login/recover/password", "", `{"id": "`+wrong[0]+`", "key": "`+wrong[1]+`"}`)
		require.Equal(t, http.StatusBadRequest, status, wrong)
	}

	// Its button gives the admin a new password, once, and the sign ins before it stop working
	status, answer := call("POST", "/login/recover/password", "", `{"id": "`+link[1]+`", "key": "`+link[2]+`"}`)
	require.Equal(t, http.StatusOK, status)
	result := answer["result"].(map[string]any)
	require.Equal(t, "admin", result["name"])
	// A test server's new password is always "password", so the sign in it ends shows the change
	password := result["password"].(string)
	status, _ = call("POST", "/login/recover/password", "", `{"id": "`+link[1]+`", "key": "`+link[2]+`"}`)
	require.Equal(t, http.StatusBadRequest, status)
	status, _ = call("GET", "/account", token, "")
	require.Equal(t, http.StatusUnauthorized, status)
	status, _ = call("POST", "/login", "", `{"name": "admin", "password": "`+password+`", "totp": "123123"}`)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, emails(), 1)
}
