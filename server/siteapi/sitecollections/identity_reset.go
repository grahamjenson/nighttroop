package sitecollections

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// ResetIdentityPasswordHandler gives one of the site's identities a new random password, and a device the Nightscout
// token made from it, shown once. An admin can reset any of the site's identities, and anyone their own.
func ResetIdentityPasswordHandler(w http.ResponseWriter, r *http.Request) {
	resetIdentity(w, r, "password")
}

// ResetIdentityTOTPHandler gives one of the site's identities that logs in with TOTP a new TOTP secret, shown once
func ResetIdentityTOTPHandler(w http.ResponseWriter, r *http.Request) {
	resetIdentity(w, r, "totp")
}

// resetIdentity resets the credential of the identity in the URL. Its JWTs issued before stop working, so resetting
// your own signs you out, and the cookie is cleared for you to sign in again with the new one.
func resetIdentity(w http.ResponseWriter, r *http.Request, credential string) {
	siteContext := httputils.GetSiteContext(r.Context())
	you := siteContext.IdentityModel
	identityID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	if identityID != you.ID && you.Role != consts.IdentityRoleAdmin {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrForbidden)
		return
	}
	identity, err := siteContext.ControlDB.Queries.GetSiteIdentityByID(r.Context(), controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}

	if credential == "totp" && identity.TotpSecretEnc == "" {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint("only an identity that logs in with TOTP can have it reset"))
		return
	}
	// Each reset hashes a password with bcrypt or encrypts a secret, so it shares the limit on creating identities
	if allowed, seconds := httputils.IdentityCreateWithinRateLimit(siteContext); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(seconds))
		return
	}

	var values *db.IdentityValues
	if credential == "totp" {
		values, err = siteContext.ControlDB.ResetSiteIdentityTOTP(r.Context(), identity, you.Name)
	} else {
		values, err = siteContext.ControlDB.ResetSiteIdentityPassword(r.Context(), identity, you.Name)
	}
	if err != nil {
		slog.Error("failed to reset an identity's credentials", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}

	signedOut := identity.ID == you.ID
	if signedOut {
		siteContext.Config.Security.ClearUserLoginCookie(w)
	}
	model := sitemodels.IdentityFromDB(identity, siteContext)
	model.Password, model.TotpSecret, model.NSToken = values.Password, values.TotpSecret, values.NSToken
	url := siteContext.SiteDB.URL()
	WriteResponse(w, r, siteContext, http.StatusOK, model, site_settings_templ.IdentityReset(url, model, signedOut), false)
}
