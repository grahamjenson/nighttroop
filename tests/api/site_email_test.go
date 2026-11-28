package api_test

import (
	"context"
	"encoding/json"
	"io"
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

// A site's admin gives the site an email on My Account. It is the site's, verified, once the link sent to it is
// opened; a site is sent one link an hour; and only an admin, on a server that sends email, can give one.
func TestSiteEmailIsVerifiedByItsLink(t *testing.T) {
	type email struct {
		From, Subject, Text string
		To                  []string
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

	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "", "", "" })
	router := server.GetMainHandler(controlDB, cfg)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	host, nsHost := site.SiteDB.Subdomain+".localhost", site.SiteDB.Subdomain+".ns.localhost"

	// call sends a request to the site, as JSON, and is its status and its body
	call := func(method, host, path, token, body string) (int, map[string]any) {
		req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Host = host
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		res, err := client.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		var answer map[string]any
		_ = json.Unmarshal(raw, &answer)
		return res.StatusCode, answer
	}
	status, signedIn := call("POST", host, "/login", "", `{"name": "admin", "password": "`+site.AdminPassword+`", "totp": "123123"}`)
	require.Equal(t, http.StatusOK, status)
	admin := signedIn["token"].(string)
	status, phone := call("GET", nsHost, "/api/v2/authorization/request/token="+site.PwdNSToken, "", "")
	require.Equal(t, http.StatusOK, status)
	stored := func() *controldb.Site {
		row, err := controlDB.Queries.GetSite(t.Context(), controldb.GetSiteParams{ID: site.SiteDB.Subdomain})
		require.NoError(t, err)
		return row
	}

	// A server that sends no email takes none
	status, _ = call("POST", host, "/settings/account/email", admin, `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusBadRequest, status)
	require.Nil(t, stored().EmailEnc)

	cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "re_key", "Nighttroop <hello@nt.example.com>", resend.URL

	// Only an admin gives the site's email, and it is one address
	status, _ = call("POST", host, "/settings/account/email", phone["token"].(string), `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusForbidden, status)
	for _, bad := range []string{"", "sam", "Sam <sam@example.com>", "sam@example.com, pat@example.com", strings.Repeat("a", 250) + "@example.com"} {
		status, _ = call("POST", host, "/settings/account/email", admin, `{"email": "`+bad+`"}`)
		require.Equal(t, http.StatusBadRequest, status, bad)
	}
	require.Empty(t, sent)

	// An email Resend did not take does not count, so the link can be sent again at once
	refuse.Store(true)
	status, _ = call("POST", host, "/settings/account/email", admin, `{"email": "sam@example.com"}`)
	require.Equal(t, http.StatusBadRequest, status)
	refuse.Store(false)

	// Requests at once send one email: a save that comes while an email is being sent is refused, and keeps nothing
	ctx, goAway := context.WithCancel(t.Context())
	req := httptest.NewRequest("POST", "http://"+host+"/settings/account/email", strings.NewReader(`{"email": "sam@example.com"}`)).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
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
		status, _ = call("POST", host, "/settings/account/email", admin, `{"email": "`+other+`"}`)
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
	require.Equal(t, "Nighttroop <hello@nt.example.com>", sent[0].From)
	require.NotContains(t, *stored().EmailEnc, "sam@example.com")
	require.Nil(t, stored().EmailVerifiedAt)
	link := regexp.MustCompile(`/settings/account/email/verify\?key=[0-9a-f]{64}`).FindString(sent[0].Text)
	require.NotEmpty(t, link, sent[0].Text)

	// A site is sent one link an hour, so it cannot be used to send email to someone again and again
	status, _ = call("POST", host, "/settings/account/email", admin, `{"email": "pat@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Len(t, sent, 1)

	// A link with another key, or another site's, verifies nothing; the link verifies the email, once, with no sign in
	status, _ = call("GET", host, "/settings/account/email/verify?key="+strings.Repeat("0", 64), "", "")
	require.Equal(t, http.StatusBadRequest, status)
	require.Nil(t, stored().EmailVerifiedAt)
	status, _ = call("GET", host, link, "", "")
	require.Equal(t, http.StatusSeeOther, status)
	require.NotNil(t, stored().EmailVerifiedAt)
	status, _ = call("GET", host, link, "", "")
	require.Equal(t, http.StatusBadRequest, status)

	// The admin takes the email away
	status, removed := call("DELETE", host, "/settings/account/email", admin, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, map[string]any{"email": "", "verified": false}, removed["result"])
	require.Nil(t, stored().EmailEnc)
	require.Nil(t, stored().EmailVerifiedAt)
}
