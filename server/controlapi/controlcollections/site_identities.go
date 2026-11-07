package controlcollections

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// IdentitiesHandler is a site's identities, newest first
func (s *SitesCollection) IdentitiesHandler(w http.ResponseWriter, r *http.Request) {
	acceptType := httputils.GetAcceptType(r)
	site, identities, herr := s.siteIdentities(r)
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	WriteResponse(w, r, acceptType, http.StatusOK, identities, control_sites_templ.SiteIdentities(site.ID, identities))
}

// InspectHandler is a site's own page: its details and every one of its identities, newest first, which an admin can
// reset or sign in to the site as
func (s *SitesCollection) InspectHandler(w http.ResponseWriter, r *http.Request) {
	acceptType := httputils.GetAcceptType(r)
	site, identities, herr := s.siteIdentities(r)
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	inspect := &controlmodels.SiteInspect{Site: site, Identities: identities}
	WriteResponse(w, r, acceptType, http.StatusOK, inspect, control_sites_templ.SiteInspect(inspect))
}

// siteIdentities is the site in the URL and its identities, newest first
func (s *SitesCollection) siteIdentities(r *http.Request) (*controlmodels.Site, []*controlmodels.SiteIdentity, *apiutils.HTTPError) {
	controlContext := httputils.GetControlContext(r.Context())
	site, herr := get(r.Context(), s, chi.URLParam(r, "id"), controlContext)
	if herr == nil && site == nil {
		herr = apiutils.ErrNotFound
	}
	if herr != nil {
		return nil, nil, herr
	}
	rows, err := controlContext.ControlDB.Queries.AllSiteIdentities(r.Context(), controldb.AllSiteIdentitiesParams{SiteID: site.ID})
	identities, herr := apiutils.Searched(rows, err, "site identities", controlmodels.SiteIdentityFromDB)
	return site, identities, herr
}

// ResetPasswordHandler gives a site's identity a new random password, shown once. Only admins can.
func (s *SitesCollection) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	s.resetIdentity(w, r, func(ctx context.Context, controlDB *db.ControlDB, identity *controldb.SiteIdentity, by string) (*db.IdentityValues, error) {
		return controlDB.ResetSiteIdentityPassword(ctx, identity, by)
	})
}

// ResetTOTPHandler gives a site's identity that logs in with TOTP a new TOTP secret, shown once. Only admins can.
func (s *SitesCollection) ResetTOTPHandler(w http.ResponseWriter, r *http.Request) {
	s.resetIdentity(w, r, func(ctx context.Context, controlDB *db.ControlDB, identity *controldb.SiteIdentity, by string) (*db.IdentityValues, error) {
		if identity.TotpSecretEnc == "" {
			return nil, apiutils.ErrConstraintCheckFailedWithHint("only an identity that logs in with TOTP can have it reset")
		}
		return controlDB.ResetSiteIdentityTOTP(ctx, identity, by)
	})
}

// resetIdentity resets the credentials of the site's identity in the URL, recorded as done by the admin, and shows
// the new ones
func (s *SitesCollection) resetIdentity(w http.ResponseWriter, r *http.Request, reset func(context.Context, *db.ControlDB, *controldb.SiteIdentity, string) (*db.IdentityValues, error)) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	identity, err := adminsSiteIdentity(r)
	var values *db.IdentityValues
	if err == nil {
		values, err = reset(r.Context(), controlContext.ControlDB, identity, controlContext.IdentityModel.Name)
	}
	if herr := apiutils.ErrorFor(err, "site identities"); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	credentials := &controlmodels.SiteIdentityCredentials{Name: identity.Name, NSToken: values.NSToken, TotpSecret: values.TotpSecret}
	if values.Password != "" {
		credentials.Password = &values.Password
	}
	WriteResponse(w, r, acceptType, http.StatusOK, credentials, control_sites_templ.SiteIdentityReset(credentials))
}

// AssumeHandler sends an admin's browser, in the new tab its form opens, to a link, good once and for a minute, that
// signs them in to the site as its identity in the URL, or answers the API with the link. The site records that they
// did on its events. Only admins can.
func (s *SitesCollection) AssumeHandler(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
		identity, err := adminsSiteIdentity(r)
		var token string
		if err == nil {
			token, err = cfg.Security.GenerateAssumeJWTToken(identity.SiteID, identity.Name, controlContext.IdentityModel.Name)
		}
		if herr := apiutils.ErrorFor(err, "site identities"); herr != nil {
			herr.Write(acceptType, w)
			return
		}
		link := &controlmodels.SiteIdentityAssume{Name: identity.Name, URL: cfg.SiteURL(identity.SiteID) + "/login/assume?token=" + url.QueryEscape(token)}
		if acceptType.API() {
			apiutils.WriteModelJSON(w, http.StatusOK, link)
			return
		}
		http.Redirect(w, r, link.URL, http.StatusSeeOther)
	}
}

// adminsSiteIdentity is the site's identity in the URL, for an admin only
func adminsSiteIdentity(r *http.Request) (*controldb.SiteIdentity, error) {
	controlContext := httputils.GetControlContext(r.Context())
	if controlContext.IdentityModel.Role != consts.ControlRoleAdmin {
		return nil, apiutils.ErrForbidden
	}
	identityID, err := strconv.ParseInt(chi.URLParam(r, "identity"), 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	return controlContext.ControlDB.Queries.GetSiteIdentityByID(r.Context(), controldb.GetSiteIdentityByIDParams{SiteID: chi.URLParam(r, "id"), ID: identityID})
}
