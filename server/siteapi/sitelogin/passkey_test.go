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

	// The link from creating the site is for the admin of that site only, until it has a passkey
	token, err := testConfig.Security.GeneratePasskeyJWTToken(siteID, consts.FIRST_ADMIN_NAME)
	require.NoError(t, err)
	r.Header.Set("Passkey-Token", token)
	user, err := passkeyLinkUser(r, controlDB, testConfig)
	require.NoError(t, err)
	require.Equal(t, admin.ID, user.identity.ID)
	otherSite, err := testConfig.Security.GeneratePasskeyJWTToken("othersite", consts.FIRST_ADMIN_NAME)
	require.NoError(t, err)
	r.Header.Set("Passkey-Token", otherSite)
	_, err = passkeyLinkUser(r, controlDB, testConfig)
	require.Error(t, err)
	r.Header.Set("Passkey-Token", token)

	credential := &webauthn.Credential{ID: []byte("credential"), PublicKey: []byte("key"), Authenticator: webauthn.Authenticator{AAGUID: []byte("a phone's model"), SignCount: 1},
		Attestation: webauthn.CredentialAttestation{ClientDataJSON: []byte("client data"), AuthenticatorData: []byte("authenticator data")}}
	require.NoError(t, savePasskey(r, controlDB, credential, admin.ID, true))

	kept, err := controlDB.Queries.GetSiteIdentityPasskey(t.Context(), controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: base64.RawURLEncoding.EncodeToString([]byte("credential"))})
	require.NoError(t, err)
	require.Equal(t, admin.ID, kept.IdentityID)
	require.NotContains(t, kept.CredentialJson, base64.StdEncoding.EncodeToString([]byte("a phone's model")))
	require.False(t, strings.Contains(kept.CredentialJson, "client data") || strings.Contains(kept.CredentialJson, base64.StdEncoding.EncodeToString([]byte("client data"))))

	// The link adds no more once the admin has one
	_, err = passkeyLinkUser(r, controlDB, testConfig)
	require.ErrorContains(t, err, "already has a passkey")

	// The admin's passkeys are the one kept, named admin, with a user handle made from its site and ID
	user, err = passkeyUserFor(r, controlDB, admin)
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
