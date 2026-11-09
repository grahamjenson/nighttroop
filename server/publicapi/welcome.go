package publicapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"nighttroop.t1d.nz/models/publicmodels"
	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// welcomeFor is how long a new site's credentials wait to be shown on its own address
const welcomeFor = 10 * time.Minute

// welcome is a new site's credentials, waiting to be shown once on its own address
type welcome struct {
	siteID string
	info   publicmodels.NewSiteInfo
}

// welcomes are the new sites' credentials not shown yet, by an ID only the browser that created the site has. They are
// only ever kept in memory, for welcomeFor, and at most 1000, many more than sites can be created in that time.
var welcomes = expirable.NewLRU[string, welcome](1000, nil, welcomeFor)

// welcomeURL keeps a new site's credentials, and is the page on the site's own address that shows them, once, with
// their ID after its #, so the ID is never in a request's address or a log
func welcomeURL(siteID string, info publicmodels.NewSiteInfo) string {
	id := security.SecureRandomString(32, security.API_KEY_CHAR_SET)
	welcomes.Add(id, welcome{siteID: siteID, info: info})
	return info.URL + "/welcome#" + id
}

// GETWelcomeHandler is a new site's welcome page, which asks for its credentials with the ID after its #
func GETWelcomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := public_templ.SiteWelcome().Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "Unable to show the welcome page", "error", err)
		}
	}
}

// POSTWelcomeHandler shows a new site's credentials, only on its own site and only once, as removing them is what
// shows them; ones already shown or not asked for in time say the backup key recovers the site
func POSTWelcomeHandler(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		found, ok := welcomes.Peek(id)
		var err error
		if ok && found.siteID == httputils.GetSiteContext(r.Context()).SiteID && welcomes.Remove(id) {
			err = public_templ.CreateSiteInfo(found.info).Render(r.Context(), w)
		} else {
			err = public_templ.WelcomeGone(config.RecoverURL()).Render(r.Context(), w)
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "Unable to show the new site's credentials", "error", err)
		}
	}
}
