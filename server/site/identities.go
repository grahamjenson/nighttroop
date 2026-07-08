package site

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/server/components/site_templ"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

func GETIdentitiesHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

		userIdentities, deviceIdentities, err := extractAndGetIdentities(siteContext, r.Context())
		if err != nil {
			errors.HandleInternalError(w, r, http.StatusInternalServerError, "Failed to get identities", err)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {

			err = site_templ.Identities(userIdentities, deviceIdentities).Render(r.Context(), w)
			if err != nil {
				fmt.Println("Error rendering identities template:", err)
				errors.TemplateFailureError(w, r, err, "Failed to render identities HTML template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = site_templ.IdentitesTable(userIdentities, deviceIdentities).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render identities settings HTMX template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))

	}
}

func extractAndGetIdentities(siteContext *httputils.SiteContext, context context.Context) ([]*site_templ.UserIdentity, []*site_templ.UserIdentity, error) {
	identities, err := siteContext.ControlDB.Queries.AllIdentities(context, control.AllIdentitiesParams{Subdomain: siteContext.Subdomain})
	if err != nil {
		return nil, nil, err
	}
	userIdentities := []*site_templ.UserIdentity{}
	deviceIdentities := []*site_templ.UserIdentity{}

	for _, identity := range identities {
		switch identity.Role.GetType() {
		case consts.IdentityTypeAdmin:
			userIdentities = append([]*site_templ.UserIdentity{site_templ.IdentityModelToUserIdentity(identity, siteContext.IdentityModel.Name)}, userIdentities...)
		case consts.IdentityTypeUser:
			userIdentities = append(userIdentities, site_templ.IdentityModelToUserIdentity(identity, siteContext.IdentityModel.Name))
		case consts.IdentityTypeDevice:
			deviceIdentities = append(deviceIdentities, site_templ.IdentityModelToUserIdentity(identity, siteContext.IdentityModel.Name))
		default:
			continue
		}
	}
	return userIdentities, deviceIdentities, nil
}

// Create Admin
func POSTIdentityHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

		// Input is a form that has a name and role
		err := r.ParseForm()
		if err != nil {
			errors.HandleInternalError(w, r, http.StatusBadRequest, "Failed to parse form", err)
			return
		}

		name := r.FormValue("name")
		role := r.FormValue("role")

		if name == "" || role == "" {
			errors.HandleInternalError(w, r, http.StatusBadRequest, "Name and role are required", nil)
			return
		}

		identityValues, err := siteContext.ControlDB.CreateIdentity(siteContext.Subdomain, name, consts.IdentityRole(role))
		if err != nil {
			errors.HandleInternalError(w, r, http.StatusInternalServerError, "Failed to create identity", err)
			return
		}

		identity := &site_templ.UserIdentity{
			ID:      "",
			Name:    name,
			Role:    role,
			Created: time.Now().Format(time.RFC3339),
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = site_templ.CreateIdentityResponse(siteContext.SiteDB.URL(), identity, identityValues.Password, identityValues.TotpSecret, identityValues.NSToken).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render identities settings HTMX template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}
	}
}
