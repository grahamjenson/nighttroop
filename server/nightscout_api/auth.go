package nightscout_api

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type NightTroopAPIAuth struct {
	controlDB *db.ControlDB
	config    config.ServerConfig
}

func (m *NightTroopAPIAuth) Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *di.Identity)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := m.AuthenticateExplicit(r)
		if err != nil {

			writeUnauthorized(w, err)
			return
		}

		if !m.HasPermission(*identity, permission) {

			writeForbidden(w, permission)
			return
		}
		next(w, r, identity)
	}
}

func (m *NightTroopAPIAuth) HasPermission(identity di.Identity, permission string) bool {
	for _, granted := range identity.Permissions {
		if permissionMatch(granted, permission) {
			return true
		}
	}
	return false
}

func (m *NightTroopAPIAuth) AuthenticateExplicit(r *http.Request) (*di.Identity, error) {
	userContext := httputils.GetUserContext(r.Context())
	if userContext == nil {
		return nil, di.ErrMissingCredentials
	}

	ntIdentity := NighttroopIdentityToNightscoutIdentity(userContext.IdentityModel.Name, userContext.IdentityModel.Role)
	return ntIdentity, nil
}

func permissionMatch(granted, wanted string) bool {
	if granted == "*" || granted == "api:*:*" {
		return true
	}
	gotParts := strings.Split(granted, ":")
	wantParts := strings.Split(wanted, ":")
	if len(gotParts) != len(wantParts) {
		return granted == wanted
	}
	for i := range gotParts {
		if gotParts[i] == "*" || gotParts[i] == "all" {
			continue
		}
		if gotParts[i] != wantParts[i] {
			return false
		}
	}
	return true
}

func writeUnauthorized(w http.ResponseWriter, err error) {
	message := "Unauthorized"
	if err == di.ErrBadToken || err == di.ErrBadJWT {
		message = "Bad access token or JWT"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = fmt.Fprintf(w, `{"description":%q,"message":%q,"status":401}`, err.Error(), message)
}

func writeForbidden(w http.ResponseWriter, permission string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = fmt.Fprintf(w, `{"message":%q,"status":403}`, "Missing permission "+permission)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, `{"description":%q,"message":%q,"status":500}`, err.Error(), "Internal Server Error")
}

func NighttroopIdentityToNightscoutIdentity(identityName string, ntid consts.IdentityRole) *di.Identity {
	// Collections
	// 'entries'
	// 'treatments'
	// 'devicestatus'
	// 'profile'
	// 'settings'
	// 'food'
	// 'properties'

	// Possible Permissions
	// 'api:*:read',
	// 'api:*:update',
	// 'api:*:create',
	// 'api:*:delete',
	// "api:entries:read",
	// "api:treatments:read",
	// "api:treatments:create",
	// "api:treatments:delete",
	// "api:devicestatus:read",
	// "api:devicestatus:create",
	// "api:devicestatus:delete",
	// "api:profile:read",
	// "api:profile:create",
	// "api:profile:delete",
	// "api:settings:read",
	// "api:settings:create",
	// "api:settings:delete",
	// "api:food:read",
	// "api:food:create",
	// "api:food:delete",

	permissions := []string{}
	switch ntid {
	case consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead:
		permissions = append(permissions, "api:*:read")
	case consts.IdentityRoleUser, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite:
		permissions = append(permissions, "api:*:*")
	default:
		return nil
	}
	return &di.Identity{
		Name:        identityName,
		Permissions: permissions,
	}
}
