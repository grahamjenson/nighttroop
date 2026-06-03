package user

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/user_templ"

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

		adminIdentities := []*user_templ.UserIdentity{}
		userIdentities := []*user_templ.UserIdentity{}
		deviceIdentities := []*user_templ.UserIdentity{}
		syncIdentities := []*user_templ.UserIdentity{}

		for _, identity := range identities {
			switch identity.Role.GetType() {
			case consts.IdentityTypeAdmin:
				adminIdentities = append(adminIdentities, user_templ.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeUser:
				userIdentities = append(userIdentities, user_templ.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeDevice:
				deviceIdentities = append(deviceIdentities, user_templ.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
			case consts.IdentityTypeSync:
				if identity.LastSeenAt == nil {
					continue // skip sync identities that have never been seen
				}
				syncIdentities = append(syncIdentities, user_templ.IdentityModelToUserIdentity(identity, userContext.IdentityName, ""))
				// remove sync users that have not been seen before

			}
		}

		if err != nil {
			errors.HandleInternalError(w, r, http.StatusInternalServerError, "Failed to get identities", err)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {

			err = user_templ.Identities(adminIdentities, userIdentities, deviceIdentities, syncIdentities).Render(r.Context(), w)
			if err != nil {
				fmt.Println("Error rendering identities template:", err)
				errors.TemplateFailureError(w, r, err, "Failed to render identities HTML template", "subdomain", userContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = user_templ.IdentitesForm(adminIdentities, userIdentities, deviceIdentities, syncIdentities).Render(r.Context(), w)
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
