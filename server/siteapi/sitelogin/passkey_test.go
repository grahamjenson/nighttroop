package sitelogin

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A passkey is kept with only what checking a sign in needs: never the authenticator's model or attestation, and its
// user handle and names say nothing about the person
func TestPasskeyKeepsNothingAboutThePerson(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	admin, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	r := httptest.NewRequest("POST", "/passkey/new/finish", nil)
	r = r.WithContext(httputils.CreateSiteContext(r.Context(), testConfig, controlDB, site.SiteDB, nil, siteID, "203.0.113.9", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON))

	// Only an identity that signs in to the website has passkeys, its own
	_, err = accountUser(r, controlDB, false)
	require.Error(t, err)
	r = r.WithContext(httputils.SetIdentityModelInSiteContext(r.Context(), admin))
	user, err := accountUser(r, controlDB, false)
	require.NoError(t, err)
	require.Equal(t, admin.ID, user.identity.ID)
	require.Empty(t, user.WebAuthnCredentials())

	credential := &webauthn.Credential{ID: []byte("credential"), PublicKey: []byte("key"), Authenticator: webauthn.Authenticator{AAGUID: []byte("a phone's model"), SignCount: 1},
		Attestation: webauthn.CredentialAttestation{ClientDataJSON: []byte("client data"), AuthenticatorData: []byte("authenticator data")}}
	require.NoError(t, savePasskey(r, controlDB, credential, admin.ID, true))

	kept, err := controlDB.Queries.GetSiteIdentityPasskey(t.Context(), controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: base64.RawURLEncoding.EncodeToString([]byte("credential"))})
	require.NoError(t, err)
	require.Equal(t, admin.ID, *kept.IdentityID)
	require.NotContains(t, kept.CredentialJson, base64.StdEncoding.EncodeToString([]byte("a phone's model")))
	require.False(t, strings.Contains(kept.CredentialJson, "client data") || strings.Contains(kept.CredentialJson, base64.StdEncoding.EncodeToString([]byte("client data"))))

	// The admin's passkeys are the one kept, named admin, with a user handle made from its site and ID
	user, err = accountUser(r, controlDB, false)
	require.NoError(t, err)
	require.Len(t, user.WebAuthnCredentials(), 1)
	require.Equal(t, "admin", user.WebAuthnName())
	require.Equal(t, "admin", user.WebAuthnDisplayName())
	require.Len(t, user.WebAuthnID(), 32)

	// Signing in again keeps its new sign count, still without the model
	credential.Authenticator = webauthn.Authenticator{AAGUID: []byte("a phone's model"), SignCount: 2}
	require.NoError(t, savePasskey(r, controlDB, credential, admin.ID, false))
	kept, err = controlDB.Queries.GetSiteIdentityPasskey(t.Context(), controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: kept.CredentialID})
	require.NoError(t, err)
	require.Contains(t, kept.CredentialJson, `"signCount":2`)
	require.NotContains(t, kept.CredentialJson, base64.StdEncoding.EncodeToString([]byte("a phone's model")))
	require.NotNil(t, kept.LastUsedAt)
}

// The browser a site's credentials were shown to adds its admin one passkey from the welcome, once: its cookie still
// lasts 15 minutes, but sent again once the admin has a passkey it adds no more, past the most an identity may have
func TestWelcomeAddsOnePasskey(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	admin, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	shown := httptest.NewRecorder()
	require.NoError(t, cfg.Security.SetWelcomePasskeyCookie(shown, siteID))
	r := httptest.NewRequest("POST", "/welcome/passkey/finish", nil)
	for _, cookie := range shown.Result().Cookies() {
		r.AddCookie(cookie)
	}
	r = r.WithContext(httputils.CreateSiteContext(r.Context(), cfg, controlDB, site.SiteDB, nil, siteID, "203.0.113.9", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON))

	user, err := welcomeAdmin(r, controlDB, cfg)
	require.NoError(t, err)
	require.Equal(t, admin.ID, user.identity.ID)
	require.NoError(t, savePasskey(r, controlDB, &webauthn.Credential{ID: []byte("welcome"), PublicKey: []byte("key")}, admin.ID, true))

	// The same cookie sent again, within its 15 minutes, adds the admin no more
	_, err = welcomeAdmin(r, controlDB, cfg)
	require.Error(t, err)
}

// An identity removes only its own passkeys
func TestPasskeyRemovedByItsIdentityOnly(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	admin, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	require.NoError(t, controlDB.Queries.CreateSiteIdentityPasskey(t.Context(), controldb.CreateSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: admin.ID, CredentialID: "one", CredentialJson: "{}"}))
	passkeys, err := controlDB.Queries.ListSiteIdentityPasskeys(t.Context(), controldb.ListSiteIdentityPasskeysParams{SiteID: siteID, IdentityID: admin.ID})
	require.NoError(t, err)
	require.Len(t, passkeys, 1)

	// Another identity's ID removes nothing
	require.NoError(t, controlDB.Queries.DeleteSiteIdentityPasskey(t.Context(), controldb.DeleteSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: admin.ID + 1000, ID: passkeys[0].ID}))
	passkeys, err = controlDB.Queries.ListSiteIdentityPasskeys(t.Context(), controldb.ListSiteIdentityPasskeysParams{SiteID: siteID, IdentityID: admin.ID})
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	require.NoError(t, controlDB.Queries.DeleteSiteIdentityPasskey(t.Context(), controldb.DeleteSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: admin.ID, ID: passkeys[0].ID}))
	passkeys, err = controlDB.Queries.ListSiteIdentityPasskeys(t.Context(), controldb.ListSiteIdentityPasskeysParams{SiteID: siteID, IdentityID: admin.ID})
	require.NoError(t, err)
	require.Empty(t, passkeys)
}
