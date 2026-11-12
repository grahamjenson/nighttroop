package push_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/push"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// Send pushes with the site's credential for the app, records it, and deletes a token Apple says is gone
func TestSend(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID, sec := controlDB.Queries, site.SiteDB.Subdomain, cfg.Security
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)

	// Apple's push service answers with the status a test sets, and keeps the alert pushed and the whole payload
	status, reason, pushed := http.StatusOK, "", ""
	var pushedPayload map[string]any
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			APS struct {
				Alert push.Alert `json:"alert"`
			} `json:"aps"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &payload)
		pushedPayload = nil
		_ = json.Unmarshal(raw, &pushedPayload)
		pushed = r.URL.Path + " " + payload.APS.Alert.Body
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"reason": reason})
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)
	defaultAPNs := push.DefaultAPNs
	push.DefaultAPNs = &push.APNs{Client: server.Client(), Host: server.URL}
	t.Cleanup(func() { push.DefaultAPNs = defaultAPNs })

	token := strings.Repeat("ab", 32)
	registration, err := push.Register(t.Context(), q, sec, siteID, identities[0].ID, push.Registration{Provider: consts.PushProviderAPNs,
		App: "loop", BundleID: "com.TEAM123456.loopkit.Loop", Token: token, Production: true})
	require.NoError(t, err)
	alert := push.Alert{Title: "Nighttroop", Body: "Test"}

	// Without a credential for Loop there is nothing to push with
	require.ErrorIs(t, push.Send(t.Context(), q, sec, registration, alert), push.ErrNoCredential)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	secretEnc, err := sec.EncryptFor(security.PurposePushSecret, siteID, string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})))
	require.NoError(t, err)
	_, err = q.CreatePushCredential(t.Context(), controldb.CreatePushCredentialParams{SiteID: &siteID, Name: "My Loop",
		Provider: consts.PushProviderAPNs, TeamOrProjectID: "TEAM123456", KeyID: new("ABC123DEFG"), SecretEnc: &secretEnc,
		AppsJson: `["loop"]`, CreatedByIdentityName: "admin"})
	require.NoError(t, err)

	stored := func() []*controldb.PushRegistration {
		found, err := q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID, Limit: 10})
		require.NoError(t, err)
		return found
	}

	// Pushed to the device token, decrypted, and recorded
	require.NoError(t, push.Send(t.Context(), q, sec, registration, alert))
	require.Equal(t, "/3/device/"+token+" Test", pushed)
	require.NotNil(t, stored()[0].LastPushedAt)
	require.Nil(t, stored()[0].LastError)
	require.Equal(t, map[string]any{"alert": map[string]any{"title": "Nighttroop", "body": "Test"}, "sound": "default"}, pushedPayload["aps"])

	// An alert's data, as Loop's remote command, is at the top of the push, which wakes the app and shows in a Focus
	command := push.Alert{Title: "Remote Bolus Entry: 1 U", Data: map[string]any{"bolus-entry": 1.0, "otp": "123456"}}
	require.NoError(t, push.Send(t.Context(), q, sec, registration, command))
	require.Equal(t, map[string]any{"bolus-entry": 1.0, "otp": "123456", "aps": map[string]any{"sound": "default",
		"alert": map[string]any{"title": "Remote Bolus Entry: 1 U", "body": ""}, "content-available": 1.0,
		"interruption-level": "time-sensitive"}}, pushedPayload)

	// Refused, its error is recorded
	status, reason = http.StatusBadRequest, "TopicDisallowed"
	require.ErrorContains(t, push.Send(t.Context(), q, sec, registration, alert), "TopicDisallowed")
	require.Contains(t, *stored()[0].LastError, "TopicDisallowed")

	// A token Apple says is gone is deleted
	status, reason = http.StatusGone, "Unregistered"
	require.ErrorIs(t, push.Send(t.Context(), q, sec, registration, alert), push.ErrTokenGone)
	require.Empty(t, stored())
}

// Send pushes through Firebase with the site's service account, never sending its JWT to the token_uri in its JSON,
// records it, and deletes a token Firebase says is unregistered
func TestSendFCM(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID, sec := controlDB.Queries, site.SiteDB.Subdomain, cfg.Security
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)

	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the service account's token_uri was used")
	}))
	t.Cleanup(elsewhere.Close)
	errorCode, pushed := "", ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "access", "expires_in": 3600})
			return
		}
		var body struct {
			Message struct {
				Token        string     `json:"token"`
				Notification push.Alert `json:"notification"`
			} `json:"message"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		pushed = r.URL.Path + " " + body.Message.Token + " " + body.Message.Notification.Body
		if errorCode != "" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"status": "NOT_FOUND",
				"details": []map[string]string{{"errorCode": errorCode}}}})
		}
	}))
	t.Cleanup(server.Close)
	defaultFCM := push.DefaultFCM
	push.DefaultFCM = &push.FCM{Client: server.Client(), Host: server.URL, TokenURL: server.URL + "/token"}
	t.Cleanup(func() { push.DefaultFCM = defaultFCM })

	token := strings.Repeat("f", 40) + ":APA91b" + strings.Repeat("x", 100)
	registration, err := push.Register(t.Context(), q, sec, siteID, identities[0].ID, push.Registration{Provider: consts.PushProviderFCM,
		App: "t1dcarer", BundleID: "nz.t1d.carer", Token: token})
	require.NoError(t, err)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	account, _ := json.Marshal(map[string]string{"type": "service_account", "project_id": "t1d-carer", "private_key_id": "key1",
		"client_email": "push@t1d-carer.iam.gserviceaccount.com", "token_uri": elsewhere.URL,
		"private_key": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))})
	secretEnc, err := sec.EncryptFor(security.PurposePushSecret, siteID, string(account))
	require.NoError(t, err)
	_, err = q.CreatePushCredential(t.Context(), controldb.CreatePushCredentialParams{SiteID: &siteID, Name: "T1D Carer",
		Provider: consts.PushProviderFCM, TeamOrProjectID: "t1d-carer", SecretEnc: &secretEnc, AppsJson: `["t1dcarer"]`,
		CreatedByIdentityName: "admin"})
	require.NoError(t, err)

	alert := push.Alert{Title: "Nighttroop", Body: "Test"}
	require.NoError(t, push.Send(t.Context(), q, sec, registration, alert))
	require.Equal(t, "/v1/projects/t1d-carer/messages:send "+token+" Test", pushed)
	found, err := q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID, Limit: 10})
	require.NoError(t, err)
	require.NotNil(t, found[0].LastPushedAt)

	errorCode = "UNREGISTERED"
	require.ErrorIs(t, push.Send(t.Context(), q, sec, registration, alert), push.ErrTokenGone)
	found, err = q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, found)
}

// A token in upper case is the same token, and an identity with the most registrations can still register one of
// them again, as Loop does with each profile it uploads
func TestRegisterAgain(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	register := func(token string) (*controldb.PushRegistration, error) {
		return push.Register(t.Context(), q, cfg.Security, siteID, identities[0].ID, push.Registration{Provider: consts.PushProviderAPNs,
			App: "loop", BundleID: "com.TEAM123456.loopkit.Loop", Token: token, Production: true})
	}
	first, err := register(strings.Repeat("ab", 32))
	require.NoError(t, err)
	again, err := register(strings.Repeat("AB", 32))
	require.NoError(t, err)
	require.Equal(t, first.ID, again.ID)

	for i := 1; i < push.MaxRegistrations; i++ {
		_, err := register(fmt.Sprintf("%064x", i))
		require.NoError(t, err)
	}
	_, err = register(strings.Repeat("cd", 32))
	require.ErrorIs(t, err, push.ErrTooManyRegistrations)
	_, err = register(strings.Repeat("ab", 32))
	require.NoError(t, err)
}

// Profiles uploaded at once register their apps one at a time, so together they cannot pass an identity's most. The
// control DB is on disk, with a connection for each, as the server's is.
func TestRegisterAloneKeepsTheMost(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain
	identities, err := controlDB.Queries.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)

	var uploads sync.WaitGroup
	for i := range 3 * push.MaxRegistrations {
		uploads.Go(func() {
			_, _ = push.RegisterAlone(t.Context(), controlDB, cfg.Security, siteID, identities[0].ID, push.Registration{
				Provider: consts.PushProviderAPNs, App: "loop", BundleID: "com.TEAM123456.loopkit.Loop",
				Token: fmt.Sprintf("%064x", i+1), Production: true})
		})
	}
	uploads.Wait()

	registered, err := controlDB.Queries.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID,
		IdentityID: &identities[0].ID, Limit: 100})
	require.NoError(t, err)
	require.Len(t, registered, push.MaxRegistrations)
}
