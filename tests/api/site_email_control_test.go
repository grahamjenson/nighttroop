package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// This file is after nightscout_test.go by name, as the test config is loaded once, with the sign in limits TestNS sets.
//
// A control admin sends one of the pre-built emails to a site whose email is verified: only on a server that sends
// email, only an email of the list, only to a verified email, and one a site a day. Control is never given the address.
func TestControlSendsAPreBuiltEmailToAVerifiedSite(t *testing.T) {
	type email struct {
		Subject, Text string
		To            []string
	}
	var sent []email
	var lock sync.Mutex
	refuse := false
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		if refuse {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var e email
		require.NoError(t, json.NewDecoder(r.Body).Decode(&e))
		sent = append(sent, e)
	}))
	t.Cleanup(resend.Close)

	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "", "", "" })
	router, siteID := server.GetMainHandler(controlDB, cfg), site.SiteDB.Subdomain
	token := controlLogin(t, router)
	call := func(method, path, body string) (int, string) {
		req := httptest.NewRequest(method, "http://control.localhost"+path, strings.NewReader(body))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr.Code, rr.Body.String()
	}
	saveEmail := func(siteID string) *string {
		emailEnc, err := cfg.Security.EncryptFor(security.PurposeSiteEmail, siteID, "sam@example.com")
		require.NoError(t, err)
		require.NoError(t, controlDB.Queries.UpdateSiteEmail(t.Context(), controldb.UpdateSiteEmailParams{ID: siteID, EmailEnc: &emailEnc}))
		return &emailEnc
	}

	// A site with no email, or one not verified, is sent nothing
	cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "re_key", "Nighttroop <hello@nt.example.com>", resend.URL
	status, body := call("GET", "/sites/"+siteID, "")
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, body, `"email_verified_at":null`)
	status, _ = call("POST", "/sites/"+siteID+"/email", `{"email": "archive"}`)
	require.Equal(t, http.StatusNotFound, status)
	emailEnc := saveEmail(siteID)
	status, _ = call("POST", "/sites/"+siteID+"/email", `{"email": "archive"}`)
	require.Equal(t, http.StatusNotFound, status)

	// Verified, the site says when, with no address
	verified, err := controlDB.Queries.VerifySiteEmail(t.Context(), controldb.VerifySiteEmailParams{ID: siteID, EmailEnc: emailEnc})
	require.NoError(t, err)
	require.EqualValues(t, 1, verified)
	status, body = call("GET", "/sites/"+siteID, "")
	require.Equal(t, http.StatusOK, status)
	require.NotContains(t, body, `"email_verified_at":null`)
	require.NotContains(t, body, "sam@example.com")
	require.NotContains(t, body, *emailEnc)

	// A server that sends no email sends none
	cfg.ResendAPIKey = ""
	status, _ = call("POST", "/sites/"+siteID+"/email", `{"email": "archive"}`)
	require.Equal(t, http.StatusBadRequest, status)
	cfg.ResendAPIKey = "re_key"

	// The admin picks an email of the list, and writes no text of their own
	for _, bad := range []string{`{}`, `{"email": "hello"}`, `{"email": "archive "}`, `{"email": "", "text": "Buy now", "subject": "Hi"}`} {
		status, _ = call("POST", "/sites/"+siteID+"/email", bad)
		require.Equal(t, http.StatusBadRequest, status, bad)
	}
	status, _ = call("POST", "/sites/nosuchsi/email", `{"email": "archive"}`)
	require.Equal(t, http.StatusNotFound, status)
	require.Empty(t, sent)

	// An email Resend did not take does not count, so it can be sent again at once
	refuse = true
	status, _ = call("POST", "/sites/"+siteID+"/email", `{"email": "archive"}`)
	require.Equal(t, http.StatusBadRequest, status)
	refuse = false

	// The email goes to the site's address, with the site's own URL, and an event says who sent which
	status, body = call("POST", "/sites/"+siteID+"/email", `{"email": "archive", "text": "Buy now"}`)
	require.Equal(t, http.StatusOK, status, body)
	require.Len(t, sent, 1)
	require.Equal(t, []string{"sam@example.com"}, sent[0].To)
	require.Equal(t, "Your Nighttroop site will be archived", sent[0].Subject)
	require.Contains(t, sent[0].Text, cfg.SiteURL(siteID))
	require.NotContains(t, sent[0].Text, "Buy now")
	events, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_emailed"), Limit: 10})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.JSONEq(t, `{"by": "admin", "email": "archive"}`, *events[0].Attributes)

	// A site is sent one email a day
	status, _ = call("POST", "/sites/"+siteID+"/email", `{"email": "storage"}`)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Len(t, sent, 1)
}

// Sent Emails lists what Resend has, each with its site when a live site has that address, and with the address
// mostly hidden. Only an admin's request for one email gives its whole address.
func TestControlListsSentEmailsWithHiddenAddresses(t *testing.T) {
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/emails":
			_, _ = w.Write([]byte(`{"object": "list", "has_more": false, "data": [
				{"id": "e2", "to": ["Sam@Example.com"], "subject": "Your Nighttroop site will be archived", "created_at": "2026-11-25 22:13:42.674981+00", "last_event": "delivered"},
				{"id": "e1", "to": ["pat@other.example.org"], "subject": "Verify the email of your Nighttroop site", "created_at": "2026-11-24T10:00:00Z", "last_event": "bounced"}]}`))
		case "/emails/e1":
			_, _ = w.Write([]byte(`{"object": "email", "id": "e1", "to": ["pat@other.example.org"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(resend.Close)

	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "", "", "" })
	router, siteID := server.GetMainHandler(controlDB, cfg), site.SiteDB.Subdomain
	token := controlLogin(t, router)
	call := func(path string) (int, string) {
		req := httptest.NewRequest("GET", "http://control.localhost"+path, nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr.Code, rr.Body.String()
	}
	emailEnc, err := cfg.Security.EncryptFor(security.PurposeSiteEmail, siteID, "sam@example.com")
	require.NoError(t, err)
	require.NoError(t, controlDB.Queries.UpdateSiteEmail(t.Context(), controldb.UpdateSiteEmailParams{ID: siteID, EmailEnc: &emailEnc}))

	// A server that sends no email lists none, and asks Resend nothing
	status, body := call("/sent-emails")
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, body, `"result":[]`)

	// Each email has when, its subject, what happened to it, its site when one has the address, and a hidden address
	cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL = "re_key", "Nighttroop <hello@nt.example.com>", resend.URL+"/emails"
	status, body = call("/sent-emails")
	require.Equal(t, http.StatusOK, status, body)
	var listed struct {
		Result []map[string]any `json:"result"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &listed))
	require.Equal(t, []map[string]any{
		{"id": "e2", "sent_at": float64(1795644822674), "subject": "Your Nighttroop site will be archived", "status": "delivered", "site_id": siteID, "to": "S•••@E•••.com"},
		{"id": "e1", "sent_at": float64(1795514400000), "subject": "Verify the email of your Nighttroop site", "status": "bounced", "site_id": nil, "to": "p•••@o•••.org"},
	}, listed.Result)
	require.NotContains(t, strings.ToLower(body), "sam@example.com")
	require.NotContains(t, body, "pat@other.example.org")

	// An admin asks for the whole address of one email; one Resend does not have is not found
	status, body = call("/sent-emails/e1/address")
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, body, `"to":"pat@other.example.org"`)
	status, _ = call("/sent-emails/e9/address")
	require.Equal(t, http.StatusNotFound, status)
	status, _ = call("/sent-emails/..%2Fdomains/address")
	require.Equal(t, http.StatusNotFound, status)
}
