package user

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

func GETIdentitiesHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		userContext := httputils.GetUserContext(r.Context())

		identities, err := userContext.UserDB.Queries.AllIdentities(r.Context())

		adminIdentities := []*templates.UserIdentity{}
		userIdentities := []*templates.UserIdentity{}
		deviceIdentities := []*templates.UserIdentity{}
		syncIdentities := []*templates.UserIdentity{}

		for _, identity := range identities {
			switch identity.Role.GetType() {
			case consts.IdentityTypeAdmin:
				adminIdentities = append(adminIdentities, templates.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeUser:
				userIdentities = append(userIdentities, templates.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeDevice:
				deviceIdentities = append(deviceIdentities, templates.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeSync:
				syncIdentities = append(syncIdentities, templates.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			}
		}

		if err != nil {
			errors.HandleInternalError(w, r, http.StatusInternalServerError, "Failed to get identities", err)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {

			err = templates.UserSettingsIdentitiesHTML(w, userContext, adminIdentities, userIdentities, deviceIdentities, syncIdentities)
			if err != nil {
				fmt.Println("Error rendering identities template:", err)
				errors.TemplateFailureError(w, r, err, "Failed to render identities HTML template", "subdomain", userContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = templates.UserSettingsIdentitiesHTMX(w, userContext, adminIdentities, userIdentities, deviceIdentities, syncIdentities)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render identities settings HTMX template", "subdomain", userContext.Subdomain)
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

// Create Admin
func POSTIdentityAdminHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handler logic here
	}
}

// Create User
func POSTIdentityUserHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handler logic here
	}
}

// Create Device
func POSTIdentityDeviceHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handler logic here
	}
}
