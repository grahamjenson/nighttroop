package user

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

func GETIdentitiesHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		userDB, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		identityName := httputils.GetIdentityNameFromHTTPContext(r.Context())
		identityRole := httputils.GetRoleFromHTTPContext(r.Context())

		identities, err := userDB.Queries.AllIdentities(r.Context())
		userIdentities := templates.IdentitiesModelToUserIdentities(identities, identityName, "")

		if err != nil {
			errors.HandleInternalError(w, r, http.StatusInternalServerError, "Failed to get identities", err)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {

			err = templates.UserSettingsIdentitiesHTML(w, subdomain, userIdentities, identityRole == "admin")
			if err != nil {
				fmt.Println("Error rendering identities template:", err)
				errors.TemplateFailureError(w, r, err, "Failed to render identities HTML template", "subdomain", subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = templates.UserSettingsIdentitiesHTMX(w, userIdentities, identityRole == "admin")
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render identities settings HTMX template", "subdomain", subdomain)
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

func POSTIdentityHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handler logic here
	}
}

func PUTIdentityHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handler logic here
	}
}
