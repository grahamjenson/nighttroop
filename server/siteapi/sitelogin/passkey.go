package sitelogin

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_login_templ"
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

// relyingParty is WebAuthn for the site, its passkeys bound to the site's own address so each works only there
func relyingParty(cfg *config.ServerConfig, r *http.Request) (*webauthn.WebAuthn, error) {
	site, err := url.Parse(cfg.SiteURL(httputils.GetSiteContext(r.Context()).SiteID))
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

// passkeyLinkUser is the identity the link from creating its site, its token in the Passkey-Token header, adds the
// first passkey to: so a link adds only one, and none once the identity has one
func passkeyLinkUser(r *http.Request, controlDB *db.ControlDB, cfg *config.ServerConfig) (passkeyUser, error) {
	name, err := cfg.Security.VerifyPasskeyJWTToken(r.Header.Get("Passkey-Token"), httputils.GetSiteContext(r.Context()).SiteID)
	if err != nil {
		return passkeyUser{}, err
	}
	identity, err := getTheIdentityeModel(controlDB, r, name)
	if err != nil {
		return passkeyUser{}, err
	}
	user, err := passkeyUserFor(r, controlDB, identity)
	if err == nil && len(user.credentials) > 0 {
		err = fmt.Errorf("%s already has a passkey", name)
	}
	return user, err
}

// GETPasskeyNew is the page to add the admin's first passkey, which the link from creating the site opens. The link's
// token is after its #, so never sent here; the page sends it to begin and finish, which check it.
func GETPasskeyNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		site_login_templ.AddPasskey(consts.FIRST_ADMIN_NAME).Render(r.Context(), w)
	}
}

// POSTPasskeyNewBegin starts adding the passkey: one the browser finds without a name being typed, checked by the
// person's fingerprint, face or PIN, and with no attestation, so the authenticator's model is not sent
func POSTPasskeyNewBegin(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := passkeyLinkUser(r, controlDB, cfg)
		if err != nil {
			passkeyError(w, r, http.StatusForbidden, "This link to add a passkey has expired or been used.", err)
			return
		}
		passkeyBegin(w, r, cfg, func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
			return rp.BeginRegistration(user, webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
				webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementRequired, UserVerification: protocol.VerificationRequired}))
		})
	}
}

// POSTPasskeyNewFinish keeps the new passkey, without the authenticator's model, and signs the identity in with it
func POSTPasskeyNewFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := passkeyLinkUser(r, controlDB, cfg)
		if err != nil {
			passkeyError(w, r, http.StatusForbidden, "This link to add a passkey has expired or been used.", err)
			return
		}
		rp, session, err := passkeySession(w, r, cfg)
		var credential *webauthn.Credential
		if err == nil {
			credential, err = rp.FinishRegistration(user, session, r)
		}
		if err != nil {
			passkeyError(w, r, http.StatusBadRequest, "The passkey could not be added.", err)
			return
		}
		if err := savePasskey(r, controlDB, credential, user.identity.ID, true); err != nil {
			passkeyError(w, r, http.StatusInternalServerError, "The passkey could not be kept.", err)
			return
		}
		passkeySignIn(w, r, cfg, user.identity)
	}
}

// POSTPasskeyLoginBegin starts signing in with whichever of the site's passkeys the browser has, checked by the
// person's fingerprint, face or PIN
func POSTPasskeyLoginBegin(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passkeyBegin(w, r, cfg, func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
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
			passkeyError(w, r, http.StatusTooManyRequests, "Too many sign in attempts, try again later.", nil)
			return
		}
		var user passkeyUser
		rp, session, err := passkeySession(w, r, cfg)
		var credential *webauthn.Credential
		if err == nil {
			credential, err = rp.FinishDiscoverableLogin(func(rawID, _ []byte) (webauthn.User, error) {
				passkey, err := controlDB.Queries.GetSiteIdentityPasskey(r.Context(), controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: base64.RawURLEncoding.EncodeToString(rawID)})
				if err != nil {
					return nil, err
				}
				identity, err := controlDB.Queries.GetSiteIdentityByID(r.Context(), controldb.GetSiteIdentityByIDParams{SiteID: siteID, ID: passkey.IdentityID})
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
			passkeyError(w, r, http.StatusUnauthorized, "That passkey could not sign you in.", err)
			return
		}
		if err := savePasskey(r, controlDB, credential, user.identity.ID, false); err != nil {
			slog.ErrorContext(r.Context(), "Unable to keep a passkey's sign count", "error", err)
		}
		passkeySignIn(w, r, cfg, user.identity)
	}
}

// passkeyBegin starts a ceremony: its options to the browser, and its session kept in a cookie for the finish
func passkeyBegin(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig, begin func(*webauthn.WebAuthn) (any, *webauthn.SessionData, error)) {
	var options any
	var session *webauthn.SessionData
	rp, err := relyingParty(cfg, r)
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
		passkeyError(w, r, http.StatusInternalServerError, "A passkey could not be asked for.", err)
		return
	}
	apiutils.WriteJSON(w, http.StatusOK, options)
}

// passkeySession is the session passkeyBegin kept, and WebAuthn for the site to finish it with
func passkeySession(w http.ResponseWriter, r *http.Request, cfg *config.ServerConfig) (*webauthn.WebAuthn, webauthn.SessionData, error) {
	var session webauthn.SessionData
	rp, err := relyingParty(cfg, r)
	if err != nil {
		return nil, session, err
	}
	sessionJSON, err := cfg.Security.PasskeySession(w, r)
	if err == nil {
		err = json.Unmarshal(sessionJSON, &session)
	}
	return rp, session, err
}

// savePasskey keeps a new passkey, or a used one's new sign count, never the authenticator's model or attestation
func savePasskey(r *http.Request, controlDB *db.ControlDB, credential *webauthn.Credential, identityID int64, isNew bool) error {
	credential.Authenticator.AAGUID, credential.Attestation = nil, webauthn.CredentialAttestation{}
	credentialJSON, err := json.Marshal(credential)
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
		passkeyError(w, r, http.StatusForbidden, "This identity cannot sign in to the website.", nil)
		return
	}
	token, expiresAt, err := cfg.Security.GenerateUserJWTToken(identity.SiteID, identity.Name, string(identity.Role))
	if err != nil {
		passkeyError(w, r, http.StatusInternalServerError, "You could not be signed in.", err)
		return
	}
	cfg.Security.SetUserLoginCookie(w, token, expiresAt)
	slog.InfoContext(r.Context(), "User logged in with a passkey", "identity", identity.Name, "role", identity.Role)
	httputils.RecordAuthClient(r, identity)
	recordSignIn(r, identity)
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"redirect": "/"})
}

// passkeyError tells the page's script why a passkey did not work, logging what went wrong
func passkeyError(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	slog.InfoContext(r.Context(), "Passkey failed", "message", message, "error", err)
	apiutils.WriteJSON(w, status, map[string]string{"error": message})
}
