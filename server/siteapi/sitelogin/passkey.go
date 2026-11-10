package sitelogin

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// passkeyUser is an identity as WebAuthn sees it: named by its identity name, as admin, never the person's, with a user
// handle made from its site and ID, so the authenticator is told nothing about who the person is
type passkeyUser struct {
	identity    *controldb.SiteIdentity
	credentials []webauthn.Credential
}

func (u passkeyUser) WebAuthnID() []byte {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s:%d", u.identity.SiteID, u.identity.ID))
	return sum[:]
}
func (u passkeyUser) WebAuthnName() string                       { return u.identity.Name }
func (u passkeyUser) WebAuthnDisplayName() string                { return u.identity.Name }
func (u passkeyUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// relyingParty is WebAuthn for the site or control at origin, its passkeys bound to that address so each works only there
func relyingParty(origin string) (*webauthn.WebAuthn, error) {
	site, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{RPID: site.Hostname(), RPDisplayName: "Nighttroop", RPOrigins: []string{site.Scheme + "://" + site.Host}})
}

// passkeyUserFor is the identity with its passkeys
func passkeyUserFor(r *http.Request, controlDB *db.ControlDB, identity *controldb.SiteIdentity) (passkeyUser, error) {
	user := passkeyUser{identity: identity}
	rows, err := controlDB.Queries.ListSiteIdentityPasskeys(r.Context(), controldb.ListSiteIdentityPasskeysParams{SiteID: identity.SiteID, IdentityID: identity.ID})
	for _, row := range rows {
		var credential webauthn.Credential
		if err := json.Unmarshal([]byte(row.CredentialJson), &credential); err != nil {
			return user, err
		}
		user.credentials = append(user.credentials, credential)
	}
	return user, err
}

// MaxPasskeys is how many passkeys an identity may have, one a device or password manager, so none fills the table
const MaxPasskeys = 10

// accountUser is the signed in identity, one that signs in to the website, with its passkeys, refused when adding one
// more than the most it may have
func accountUser(r *http.Request, controlDB *db.ControlDB, adding bool) (passkeyUser, error) {
	identity := httputils.GetSiteContext(r.Context()).IdentityModel
	if identity == nil || !can.LoginToWebsite(identity.Role) {
		return passkeyUser{}, fmt.Errorf("only an identity that signs in to the website has passkeys")
	}
	user, err := passkeyUserFor(r, controlDB, identity)
	if err == nil && adding && len(user.credentials) >= MaxPasskeys {
		err = fmt.Errorf("you have %d passkeys, the most you can add; remove one first", MaxPasskeys)
	}
	return user, err
}

// GETAccount is the signed in identity's My Account page: who it is signed in as, and its passkeys to add and remove,
// and for an admin its site's storage
func GETAccount(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())
		user, err := accountUser(r, controlDB, false)
		var passkeys []*controldb.Passkey
		if err == nil {
			passkeys, err = controlDB.Queries.ListSiteIdentityPasskeys(r.Context(), controldb.ListSiteIdentityPasskeysParams{SiteID: siteContext.SiteID, IdentityID: user.identity.ID})
		}
		if err != nil {
			slog.InfoContext(r.Context(), "My Account refused", "error", err)
			http.Error(w, "Only a user who signs in to the website has an account page", http.StatusForbidden)
			return
		}
		sitecollections.WriteResponse(w, r, siteContext, http.StatusOK, passkeys, site_settings_templ.MyAccount(user.identity.Name, string(user.identity.Role), passkeys, MaxPasskeys, "/settings/account", true, sitecollections.AccountStorage(r)), true)
	}
}

// POSTAccountPasskeyBegin starts adding the signed in identity a passkey
func POSTAccountPasskeyBegin(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := accountUser(r, controlDB, true)
		if err != nil {
			PasskeyError(w, r, http.StatusForbidden, "A passkey cannot be added: "+err.Error()+".", err)
			return
		}
		beginAdding(w, r, cfg, user)
	}
}

// beginAdding starts adding the identity a passkey, found without a name typed, checked by the person's fingerprint,
// face or PIN, with no attestation (the authenticator's model), and not one it already has
func beginAdding(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, user passkeyUser) {
	PasskeyBegin(w, r, cfg, siteOrigin(cfg, r), func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
		return rp.BeginRegistration(user, webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
			webauthn.WithExclusions(webauthn.Credentials(user.credentials).CredentialDescriptors()),
			webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementRequired, UserVerification: protocol.VerificationRequired}))
	})
}

// finishAdding is the identity's new passkey from the browser's answer to beginAdding, kept, never the authenticator's
// model, or an error already written
func finishAdding(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, controlDB *db.ControlDB, user passkeyUser) bool {
	rp, session, err := PasskeySession(w, r, cfg, siteOrigin(cfg, r))
	var credential *webauthn.Credential
	if err == nil {
		credential, err = rp.FinishRegistration(user, session, r)
	}
	if err != nil {
		PasskeyError(w, r, http.StatusBadRequest, "The passkey could not be added.", err)
		return false
	}
	if err := savePasskey(r, controlDB, credential, user.identity.ID, true); err != nil {
		PasskeyError(w, r, http.StatusInternalServerError, "The passkey could not be kept.", err)
		return false
	}
	slog.InfoContext(r.Context(), "Passkey added", "identity", user.identity.Name)
	return true
}

// POSTAccountPasskeyFinish keeps the signed in identity's new passkey, never the authenticator's model, then My Account
func POSTAccountPasskeyFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := accountUser(r, controlDB, true)
		if err != nil {
			PasskeyError(w, r, http.StatusBadRequest, "The passkey could not be added.", err)
			return
		}
		if finishAdding(w, r, cfg, controlDB, user) {
			apiutils.WriteJSON(w, http.StatusOK, map[string]string{"redirect": "/settings/account"})
		}
	}
}

// welcomeAdmin is the new site's admin, with its passkeys, when this browser was shown the site's credentials on its
// welcome page in the last 15 minutes
func welcomeAdmin(r *http.Request, controlDB *db.ControlDB, cfg *config.ServerConfig) (passkeyUser, error) {
	siteID := httputils.GetSiteContext(r.Context()).SiteID
	welcomed, err := cfg.Security.WelcomePasskeySite(r)
	if err == nil && welcomed != siteID {
		err = fmt.Errorf("the credentials were shown for another site")
	}
	var identity *controldb.SiteIdentity
	if err == nil {
		identity, err = controlDB.Queries.GetSiteIdentity(r.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	}
	if err != nil {
		return passkeyUser{}, err
	}
	return passkeyUserFor(r, controlDB, identity)
}

// POSTWelcomePasskeyBegin starts adding the new site's admin a passkey from its welcome page, by the browser it showed
// the site's credentials to
func POSTWelcomePasskeyBegin(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := welcomeAdmin(r, controlDB, cfg)
		if err != nil {
			PasskeyError(w, r, http.StatusForbidden, "This page can no longer add a passkey; log in, then add one on My Account.", err)
			return
		}
		beginAdding(w, r, cfg, user)
	}
}

// POSTWelcomePasskeyFinish keeps the new site's admin's passkey, then signs the admin in with it, as a passkey sign in
// does; the welcome adds only the one
func POSTWelcomePasskeyFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := welcomeAdmin(r, controlDB, cfg)
		if err != nil {
			PasskeyError(w, r, http.StatusForbidden, "This page can no longer add a passkey; log in, then add one on My Account.", err)
			return
		}
		if finishAdding(w, r, cfg, controlDB, user) {
			cfg.Security.ClearWelcomePasskeyCookie(w)
			passkeySignIn(w, r, cfg, user.identity)
		}
	}
}

// DELETEAccountPasskey removes one of the signed in identity's passkeys, as for a lost device, showing My Account again
func DELETEAccountPasskey(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := accountUser(r, controlDB, false)
		var id int64
		if err == nil {
			id, err = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		}
		if err == nil {
			err = controlDB.Queries.DeleteSiteIdentityPasskey(r.Context(), controldb.DeleteSiteIdentityPasskeyParams{SiteID: user.identity.SiteID, IdentityID: user.identity.ID, ID: id})
		}
		if err != nil {
			slog.InfoContext(r.Context(), "Passkey not removed", "error", err)
			http.Error(w, "The passkey could not be removed", http.StatusBadRequest)
			return
		}
		GETAccount(controlDB)(w, r)
	}
}

// POSTPasskeyLoginBegin starts signing in with whichever of the site's passkeys the browser has, checked by the
// person's fingerprint, face or PIN
func POSTPasskeyLoginBegin(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		PasskeyBegin(w, r, cfg, siteOrigin(cfg, r), func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
			return rp.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
		})
	}
}

// POSTPasskeyLoginFinish signs in the enabled identity whose passkey answered, keeping the passkey's new sign count.
// Attempts are limited by IP, as before signing in nothing else is known.
func POSTPasskeyLoginFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteID := httputils.GetSiteContext(r.Context()).SiteID
		if allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(cfg, r, siteID, "passkey"); !allowed {
			PasskeyError(w, r, http.StatusTooManyRequests, "Too many sign in attempts, try again later.", nil)
			return
		}
		var user passkeyUser
		rp, session, err := PasskeySession(w, r, cfg, siteOrigin(cfg, r))
		var credential *webauthn.Credential
		if err == nil {
			credential, err = rp.FinishDiscoverableLogin(func(rawID, _ []byte) (webauthn.User, error) {
				passkey, err := controlDB.Queries.GetSiteIdentityPasskey(r.Context(), controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: base64.RawURLEncoding.EncodeToString(rawID)})
				if err != nil {
					return nil, err
				}
				identity, err := controlDB.Queries.GetSiteIdentityByID(r.Context(), controldb.GetSiteIdentityByIDParams{SiteID: siteID, ID: *passkey.IdentityID})
				if err == nil && !identity.Enabled {
					err = fmt.Errorf("identity disabled")
				}
				if err != nil {
					return nil, err
				}
				user, err = passkeyUserFor(r, controlDB, identity)
				return user, err
			}, session, r)
		}
		if err != nil {
			PasskeyError(w, r, http.StatusUnauthorized, "That passkey could not sign you in.", err)
			return
		}
		if err := savePasskey(r, controlDB, credential, user.identity.ID, false); err != nil {
			slog.ErrorContext(r.Context(), "Unable to keep a passkey's sign count", "error", err)
		}
		passkeySignIn(w, r, cfg, user.identity)
	}
}

// siteOrigin is the address of the request's site, which its passkeys are bound to
func siteOrigin(cfg *config.ServerConfig, r *http.Request) string {
	return cfg.SiteURL(httputils.GetSiteContext(r.Context()).SiteID)
}

// PasskeyBegin starts a ceremony at origin: its options to the browser, and its session kept in a cookie for the finish
func PasskeyBegin(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, origin string, begin func(*webauthn.WebAuthn) (any, *webauthn.SessionData, error)) {
	var options any
	var session *webauthn.SessionData
	rp, err := relyingParty(origin)
	if err == nil {
		options, session, err = begin(rp)
	}
	var sessionJSON []byte
	if err == nil {
		sessionJSON, err = json.Marshal(session)
	}
	if err == nil {
		err = cfg.Security.SetPasskeySessionCookie(w, sessionJSON)
	}
	if err != nil {
		PasskeyError(w, r, http.StatusInternalServerError, "A passkey could not be asked for.", err)
		return
	}
	apiutils.WriteJSON(w, http.StatusOK, options)
}

// PasskeySession is the session PasskeyBegin kept, and WebAuthn at origin to finish it with
func PasskeySession(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, origin string) (*webauthn.WebAuthn, webauthn.SessionData, error) {
	var session webauthn.SessionData
	rp, err := relyingParty(origin)
	if err != nil {
		return nil, session, err
	}
	sessionJSON, err := cfg.Security.PasskeySession(w, r)
	if err == nil {
		err = json.Unmarshal(sessionJSON, &session)
	}
	return rp, session, err
}

// PasskeyJSON is what is kept of a passkey, never the authenticator's model or attestation
func PasskeyJSON(credential *webauthn.Credential) ([]byte, error) {
	credential.Authenticator.AAGUID, credential.Attestation = nil, webauthn.CredentialAttestation{}
	return json.Marshal(credential)
}

// savePasskey keeps a site identity's new passkey, or a used one's new sign count
func savePasskey(r *http.Request, controlDB *db.ControlDB, credential *webauthn.Credential, identityID int64, isNew bool) error {
	credentialJSON, err := PasskeyJSON(credential)
	if err != nil {
		return err
	}
	siteID, credentialID := httputils.GetSiteContext(r.Context()).SiteID, base64.RawURLEncoding.EncodeToString(credential.ID)
	if isNew {
		return controlDB.Queries.CreateSiteIdentityPasskey(r.Context(), controldb.CreateSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: identityID, CredentialID: credentialID, CredentialJson: string(credentialJSON)})
	}
	return controlDB.Queries.UpdateSiteIdentityPasskeyUsed(r.Context(), controldb.UpdateSiteIdentityPasskeyUsedParams{SiteID: siteID, CredentialID: credentialID, CredentialJson: string(credentialJSON)})
}

// passkeySignIn signs the identity in as a password and TOTP do; a passkey is already two factors, the device and the
// person's fingerprint, face or PIN, so it is enough on its own
func passkeySignIn(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, identity *controldb.SiteIdentity) {
	if !can.LoginToWebsite(identity.Role) {
		PasskeyError(w, r, http.StatusForbidden, "This identity cannot sign in to the website.", nil)
		return
	}
	token, expiresAt, err := cfg.Security.GenerateUserJWTToken(identity.SiteID, identity.Name, string(identity.Role))
	if err != nil {
		PasskeyError(w, r, http.StatusInternalServerError, "You could not be signed in.", err)
		return
	}
	cfg.Security.SetUserLoginCookie(w, token, expiresAt)
	slog.InfoContext(r.Context(), "User logged in with a passkey", "identity", identity.Name, "role", identity.Role)
	httputils.RecordAuthClient(r, identity)
	recordSignIn(r, identity)
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"redirect": "/"})
}

// PasskeyError tells the page's script why a passkey did not work, logging what went wrong
func PasskeyError(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	slog.InfoContext(r.Context(), "Passkey failed", "message", message, "error", err)
	apiutils.WriteJSON(w, status, map[string]string{"error": message})
}
