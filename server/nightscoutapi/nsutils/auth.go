package nsutils

import (
	"net/http"
	"slices"
	"time"

	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

func Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *Identity)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := AuthenticateExplicit(r)
		if err != nil {
			WriteError(w, err)
			return
		}

		if !HasPermission(*identity, permission) {
			WriteError(w, ErrMissingPermission(permission))
			return
		}
		next(w, r, identity)
	}
}

// HasPermission is whether the identity's Nightscout permissions grant the permission, see can.NightscoutPermitted
func HasPermission(identity Identity, permission string) bool {
	return can.NightscoutPermitted(identity.Permissions, permission)
}

func AuthenticateExplicit(r *http.Request) (*Identity, *HTTPError) {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext == nil {
		return nil, ErrMissingCredentials
	}

	ntIdentity, err := NighttroopIdentityToNightscoutIdentity(siteContext.IdentityModel.Name, siteContext.IdentityModel.Role, siteContext.IdentityModel.Client)
	if err != nil {
		return nil, ErrMissingCredentials
	}
	return ntIdentity, nil
}

// NighttroopIdentityToNightscoutIdentity is the identity with its Nightscout permissions, see can.NightscoutPermissions
func NighttroopIdentityToNightscoutIdentity(identityName string, role consts.IdentityRole, client *string) (*Identity, error) {
	permissions, err := can.NightscoutPermissions(role, client)
	if err != nil {
		return nil, err
	}
	return &Identity{Name: identityName, Permissions: permissions}, nil
}

// Authorization is the identity's JWT and permissions, as Nightscout gives them from /api/v2/authorization/request and
// as status.json's authorized. LoopFollow takes an identity as one that can write only from an exact
// api:treatments:create, not a wildcard, so that is listed when it has it
func Authorization(siteContext *httputils.SiteContext) (map[string]any, error) {
	identity, err := NighttroopIdentityToNightscoutIdentity(siteContext.IdentityModel.Name, siteContext.IdentityModel.Role, siteContext.IdentityModel.Client)
	if err != nil {
		return nil, err
	}
	token, expiresAt, err := siteContext.Config.Security.GenerateUserJWTToken(siteContext.SiteID, siteContext.IdentityModel.Name, string(siteContext.IdentityModel.Role))
	if err != nil {
		return nil, err
	}
	permissions := identity.Permissions
	if HasPermission(*identity, "api:treatments:create") && !slices.Contains(permissions, "api:treatments:create") {
		permissions = append(slices.Clone(permissions), "api:treatments:create")
	}
	return map[string]any{"token": token, "sub": siteContext.IdentityModel.Name, "permissionGroups": [][]string{permissions},
		"iat": time.Now().Unix(), "exp": expiresAt.Unix()}, nil
}
