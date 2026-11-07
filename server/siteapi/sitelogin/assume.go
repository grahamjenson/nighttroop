package sitelogin

import (
	"fmt"
	"log/slog"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// GETAssume signs a control admin in as the site's identity a link from control names, once and within a minute of
// the link being made, and records that they did on the site's events, as they did not sign in with its credentials
func GETAssume(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())
		if err := assume(controlDB, config, w, r, siteContext.SiteID); err != nil {
			slog.WarnContext(r.Context(), "Refused a control link to sign in", "error", err)
			http.Error(w, "This link to sign in is for another site, has been used or has expired: make another from control", http.StatusUnauthorized)
			return
		}
		// The site's own page moves on to its summary, as a redirect would be part of the navigation from control, which
		// a browser that sees control as another site does not send the Strict sign in cookie on. The link's URL, with
		// its token, is not sent on as the referrer.
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><meta http-equiv="refresh" content="0;url=/"><a href="/">Open the site</a>`))
	}
}

// assume signs the request in as the identity its link token names, if the token is good for the site
func assume(controlDB *db.ControlDB, config *config.ServerConfig, w http.ResponseWriter, r *http.Request, siteID string) error {
	name, by, err := config.Security.VerifyAssumeJWTToken(r.URL.Query().Get("token"), siteID)
	if err != nil {
		return err
	}
	identity, err := getTheIdentityeModel(controlDB, r, name)
	if err != nil {
		return err
	}
	token, expiresAt, err := config.Security.GenerateUserJWTToken(siteID, identity.Name, string(identity.Role))
	if err != nil {
		return fmt.Errorf("failed to generate JWT token: %w", err)
	}
	config.Security.SetUserLoginCookie(w, token, expiresAt)
	slog.InfoContext(r.Context(), "Control admin signed in as an identity", "identity", identity.Name, "by", by)
	controlDB.Logs.SignedInByControlEvent(r.Context(), siteID, identity.Name, by)
	return nil
}
