package controlapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// controlUser is a control admin as WebAuthn sees it, with its passkeys: named by its admin name, with a user handle
// made from its ID, so the authenticator is told nothing about who the person is
type controlUser struct {
	identity    *controldb.ControlIdentity
	passkeys    []*controldb.Passkey
	credentials []webauthn.Credential
}

func (u controlUser) WebAuthnID() []byte {
	sum := sha256.Sum256(fmt.Appendf(nil, "control:%d", u.identity.ID))
	return sum[:]
}
func (u controlUser) WebAuthnName() string                       { return u.identity.Name }
func (u controlUser) WebAuthnDisplayName() string                { return u.identity.Name }
func (u controlUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// controlOrigin is the control site's address, which its admins' passkeys are bound to
func controlOrigin(cfg *config.ServerConfig) string {
	return cfg.LinkURL("control")
}

// controlUserFor is the control admin with its passkeys, refused when adding one more than the most it may have
func controlUserFor(r *http.Request, controlDB *db.ControlDB, identity *controldb.ControlIdentity, adding bool) (controlUser, error) {
	user := controlUser{identity: identity, passkeys: []*controldb.Passkey{}}
	rows, err := controlDB.Queries.ListControlPasskeys(r.Context(), controldb.ListControlPasskeysParams{ControlIdentityID: identity.ID})
	for _, row := range rows {
		var credential webauthn.Credential
		if err := json.Unmarshal([]byte(row.CredentialJson), &credential); err != nil {
			return user, err
		}
		user.passkeys, user.credentials = append(user.passkeys, row), append(user.credentials, credential)
	}
	if err == nil && adding && len(rows) >= sitelogin.MaxPasskeys {
		err = fmt.Errorf("you have %d passkeys, the most you can add; remove one first", sitelogin.MaxPasskeys)
	}
	return user, err
}

// GETAccount is the signed in control admin's Account page: who it is, and its passkeys to add and remove
func GETAccount(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := httputils.GetControlContext(r.Context()).IdentityModel
		user, err := controlUserFor(r, controlDB, identity, false)
		if err != nil {
			slog.ErrorContext(r.Context(), "Unable to list the control admin's passkeys", "error", err)
			apiutils.ErrInternalServerError.Write(httputils.GetAcceptType(r), w)
			return
		}
		controlcollections.WriteResponse(w, r, httputils.GetAcceptType(r), http.StatusOK, user.passkeys,
			site_settings_templ.MyAccount(identity.Name, string(identity.Role), user.passkeys, sitelogin.MaxPasskeys, "/account", false, nil))
	}
}

// POSTAccountPasskeyBegin starts adding the control admin a passkey, on the control site's address, as a site's
// identity adds one: found without a name typed, checked by the person, with no attestation, and not one it has
func POSTAccountPasskeyBegin(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := controlUserFor(r, controlDB, httputils.GetControlContext(r.Context()).IdentityModel, true)
		if err != nil {
			sitelogin.PasskeyError(w, r, http.StatusForbidden, "A passkey cannot be added: "+err.Error()+".", err)
			return
		}
		sitelogin.PasskeyBegin(w, r, cfg, controlOrigin(cfg), func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
			return rp.BeginRegistration(user, webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
				webauthn.WithExclusions(webauthn.Credentials(user.credentials).CredentialDescriptors()),
				webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementRequired, UserVerification: protocol.VerificationRequired}))
		})
	}
}

// POSTAccountPasskeyFinish keeps the control admin's new passkey, never the authenticator's model, then its Account
func POSTAccountPasskeyFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := controlUserFor(r, controlDB, httputils.GetControlContext(r.Context()).IdentityModel, true)
		var rp *webauthn.WebAuthn
		var session webauthn.SessionData
		if err == nil {
			rp, session, err = sitelogin.PasskeySession(w, r, cfg, controlOrigin(cfg))
		}
		var credential *webauthn.Credential
		if err == nil {
			credential, err = rp.FinishRegistration(user, session, r)
		}
		var credentialJSON []byte
		if err == nil {
			credentialJSON, err = sitelogin.PasskeyJSON(credential)
		}
		if err == nil {
			err = controlDB.Queries.CreateControlPasskey(r.Context(), controldb.CreateControlPasskeyParams{ControlIdentityID: user.identity.ID,
				CredentialID: base64.RawURLEncoding.EncodeToString(credential.ID), CredentialJson: string(credentialJSON)})
		}
		if err != nil {
			sitelogin.PasskeyError(w, r, http.StatusBadRequest, "The passkey could not be added.", err)
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]string{"redirect": "/account"})
	}
}

// DELETEAccountPasskey removes one of the control admin's own passkeys, as for a lost device, showing its Account again
func DELETEAccountPasskey(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err == nil {
			err = controlDB.Queries.DeleteControlPasskey(r.Context(), controldb.DeleteControlPasskeyParams{ControlIdentityID: httputils.GetControlContext(r.Context()).IdentityModel.ID, ID: id})
		}
		if err != nil {
			http.Error(w, "The passkey could not be removed", http.StatusBadRequest)
			return
		}
		GETAccount(controlDB)(w, r)
	}
}

// POSTPasskeyLoginBegin starts signing in to control with whichever of its admins' passkeys the browser has
func POSTPasskeyLoginBegin(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sitelogin.PasskeyBegin(w, r, cfg, controlOrigin(cfg), func(rp *webauthn.WebAuthn) (any, *webauthn.SessionData, error) {
			return rp.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
		})
	}
}

// POSTPasskeyLoginFinish signs in the control admin whose passkey answered, on its own, as a password and TOTP do,
// keeping the passkey's new sign count. Attempts are limited by IP, as before signing in nothing else is known.
func POSTPasskeyLoginFinish(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(cfg, r, "passkey"); !allowed {
			sitelogin.PasskeyError(w, r, http.StatusTooManyRequests, "Too many sign in attempts, try again later.", nil)
			return
		}
		var user controlUser
		rp, session, err := sitelogin.PasskeySession(w, r, cfg, controlOrigin(cfg))
		var credential *webauthn.Credential
		if err == nil {
			credential, err = rp.FinishDiscoverableLogin(func(rawID, _ []byte) (webauthn.User, error) {
				passkey, err := controlDB.Queries.GetControlPasskey(r.Context(), controldb.GetControlPasskeyParams{CredentialID: base64.RawURLEncoding.EncodeToString(rawID)})
				if err != nil {
					return nil, err
				}
				identity, err := controlDB.Queries.GetControlIdentityByID(r.Context(), controldb.GetControlIdentityByIDParams{ID: *passkey.ControlIdentityID})
				if err == nil {
					user, err = controlUserFor(r, controlDB, identity, false)
				}
				return user, err
			}, session, r)
		}
		from := cfg.GeoIP.Locate(httputils.GetIPFromHTTPContext(r.Context()))
		var credentialJSON []byte
		if err == nil {
			credentialJSON, err = sitelogin.PasskeyJSON(credential)
		}
		if err != nil {
			controlDB.Logs.SignInFailedEvent(r.Context(), "", from)
			sitelogin.PasskeyError(w, r, http.StatusUnauthorized, "That passkey could not sign you in.", err)
			return
		}
		if err := controlDB.Queries.UpdateControlPasskeyUsed(r.Context(), controldb.UpdateControlPasskeyUsedParams{CredentialJson: string(credentialJSON), CredentialID: base64.RawURLEncoding.EncodeToString(credential.ID)}); err != nil {
			slog.ErrorContext(r.Context(), "Unable to keep a control passkey's sign count", "error", err)
		}
		token, expiresAt, err := cfg.Security.GenerateControlJWTToken(user.identity.Name, string(user.identity.Role))
		if err != nil {
			sitelogin.PasskeyError(w, r, http.StatusInternalServerError, "You could not be signed in.", err)
			return
		}
		controlDB.Logs.SignedInEvent(r.Context(), "", user.identity.Name, from)
		cfg.Security.SetControlLoginCookie(w, token, expiresAt)
		slog.InfoContext(r.Context(), "Admin logged in with a passkey", "admin", user.identity.Name, "role", user.identity.Role)
		apiutils.WriteJSON(w, http.StatusOK, map[string]string{"redirect": "/"})
	}
}
