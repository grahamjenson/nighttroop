package publicapi

import (
	"log/slog"
	"net/http"
	"nighttroop.t1d.nz/utils/consts"

	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// POSTCheckTOTPHandler checks a code from the authenticator app a new site's admin TOTP secret was added to, and once
// it matches, enables the site's login button. It is only a check for the person setting up: it changes nothing, and
// does not look the site up, so it cannot tell anyone whether a site exists.
func POSTCheckTOTPHandler(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		site := r.FormValue("site")
		var err error
		if utils.SyntacticallyValidateSiteID(site) && config.Security.TOTPValidate(r.FormValue("totp"), r.FormValue("secret")) {
			err = public_templ.TOTPChecked(config.SiteURL(site)+"/login?name="+consts.FIRST_ADMIN_NAME).Render(r.Context(), w)
		} else {
			err = public_templ.TOTPCheckFailed().Render(r.Context(), w)
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "Unable to show the TOTP check", "error", err)
		}
	}
}
