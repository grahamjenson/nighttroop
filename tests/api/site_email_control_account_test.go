package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

// This file is after nightscout_test.go by name, as the test config is loaded once, with the sign in limits TestNS sets.
//
// A control admin gives an email of their own on Account. It is theirs, verified, once the link sent to it is opened;
// an admin is sent one link an hour; and only a server that sends email takes one.
func TestControlAdminEmailIsVerifiedByItsLink(t *testing.T) {
	type email struct {
		Subject, Text string
		To            []string
	}
	var sent []email
	var lock sync.Mutex
	var refuse atomic.Bool
	// The first email waits at Resend until release, and sending says it is there, so more requests come while it is sent
	sending, held := make(chan struct{}, 1), make(chan struct{})
	release := sync.OnceFunc(func() { close(held) })
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if refuse.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var e email
		require.NoError(t, json.NewDecoder(r.Body).Decode(&e))
		lock.Lock()
		sent = append(sent, e)
		first := len(sent) == 1
		lock.Unlock()
		if first {
			sending <- struct{}{}
			<-held
		}
	}))
	t.Cleanup(resend.Close)
	t.Cleanup(release)

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
	stored := func() *controldb.ControlIdentity {
		admin, err := controlDB.Queries.GetControlIdentity(t.Context(), controldb.GetControlIdentityParams{Name: "admin"})
		require.NoError(t, err)
		return admin
	}

	// A server that sends no email takes none
	status, _ := call("POST", "/account/email", token, `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusBadRequest, status)
	require.Nil(t, stored().EmailEnc)

	cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "re_key", "Nighttroop <hello@nt.example.com>", resend.URL

	// It is one address, given by a signed in admin
	status, _ = call("POST", "/account/email", "", `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusUnauthorized, status)
	for _, bad := range []string{"", "sam", "Sam <sam@example.com>", "sam@example.com, pat@example.com", strings.Repeat("a", 250) + "@example.com"} {
		status, _ = call("POST", "/account/email", token, `{"email": "`+bad+`"}`)
		require.Equal(t, http.StatusBadRequest, status, bad)
	}
	require.Empty(t, sent)

	// An email Resend did not take does not count, so the link can be sent again at once
	refuse.Store(true)
	status, _ = call("POST", "/account/email", token, `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusBadRequest, status)
	refuse.Store(false)

	// Requests at once send one email: a save that comes while an email is being sent is refused, and keeps nothing
	ctx, goAway := context.WithCancel(t.Context())
	req := httptest.NewRequest("POST", "http://control.localhost/account/email", strings.NewReader(`{"email": "sam@example.com"}`)).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	first, answered := httptest.NewRecorder(), make(chan struct{})
	go func() {
		defer close(answered)
		router.ServeHTTP(first, req)
	}()
	select {
	case <-sending:
	case <-time.After(10 * time.Second):
		require.Fail(t, "the first email did not get to Resend")
	}
	for _, other := range []string{"pat@example.com", "kim@example.com"} {
		status, _ = call("POST", "/account/email", token, `{"email": "`+other+`"}`)
		require.Equal(t, http.StatusTooManyRequests, status, other)
	}

	// A client that goes away while its email is being sent does not stop the email, so the email counts
	goAway()
	release()
	<-answered

	// The email is kept encrypted and not verified, and a link is sent to it
	require.Equal(t, http.StatusOK, first.Code)
	var saved map[string]any
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &saved))
	require.Equal(t, map[string]any{"email": "sam@example.com", "verified": false}, saved["result"])
	require.Len(t, sent, 1)
	require.Equal(t, []string{"sam@example.com"}, sent[0].To)
	require.Equal(t, "Verify your email for Nighttroop control", sent[0].Subject)
	require.NotContains(t, *stored().EmailEnc, "sam@example.com")
	require.Nil(t, stored().EmailVerifiedAt)
	link := regexp.MustCompile(`/account/email/verify\?id=[0-9]+&key=[0-9a-f]{64}`).FindString(sent[0].Text)
	require.NotEmpty(t, link, sent[0].Text)

	// An admin is sent one link an hour, so control cannot be used to send email to someone again and again
	status, _ = call("POST", "/account/email", token, `{"email": "pat@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Len(t, sent, 1)

	// A link with another key, or of another admin, verifies nothing; the link verifies the email, once, with no sign in
	key := link[strings.Index(link, "key=")+len("key="):]
	for _, wrong := range []string{"/account/email/verify?id=" + "0&key=" + key, strings.Replace(link, key, strings.Repeat("0", 64), 1)} {
		status, _ = call("GET", wrong, "", "")
		require.Equal(t, http.StatusBadRequest, status, wrong)
	}
	require.Nil(t, stored().EmailVerifiedAt)
	status, _ = call("GET", link, "", "")
	require.Equal(t, http.StatusSeeOther, status)
	require.NotNil(t, stored().EmailVerifiedAt)
	status, _ = call("GET", link, "", "")
	require.Equal(t, http.StatusBadRequest, status)

	// The admin takes the email away
	status, removed := call("DELETE", "/account/email", token, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, map[string]any{"email": "", "verified": false}, removed["result"])
	require.Nil(t, stored().EmailEnc)
	require.Nil(t, stored().EmailVerifiedAt)
}
