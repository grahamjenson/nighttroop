package push_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/push"
	"nighttroop.t1d.nz/utils/testutils"
)

// Send pushes with the site's credential for the app, records it, and deletes a token Apple says is gone
func TestSend(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID, sec := controlDB.Queries, site.SiteDB.Subdomain, cfg.Security
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)

	// Apple's push service answers with the status a test sets, and keeps the alert pushed
	status, reason, pushed := http.StatusOK, "", ""
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			APS struct {
				Alert push.Alert `json:"alert"`
			} `json:"aps"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
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
	secretEnc, err := sec.Encrypt(string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})))
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

	// Refused, its error is recorded
	status, reason = http.StatusBadRequest, "TopicDisallowed"
	require.ErrorContains(t, push.Send(t.Context(), q, sec, registration, alert), "TopicDisallowed")
	require.Contains(t, *stored()[0].LastError, "TopicDisallowed")

	// A token Apple says is gone is deleted
	status, reason = http.StatusGone, "Unregistered"
	require.ErrorIs(t, push.Send(t.Context(), q, sec, registration, alert), push.ErrTokenGone)
	require.Empty(t, stored())
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
