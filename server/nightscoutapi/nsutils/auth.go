package nsutils

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"nighttroop.t1d.nz/models/sitemodels"
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

func HasPermission(identity Identity, permission string) bool {
	for _, granted := range identity.Permissions {
		if permissionMatch(granted, permission) {
			return true
		}
	}
	return false
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

func NighttroopIdentityToNightscoutIdentity(identityName string, ntid consts.IdentityRole, client *string) (*Identity, error) {
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
	// "api:settings:admin"

	permissions := []string{}
	switch ntid {
	case consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead:
		permissions = append(permissions, "api:*:read")
	case consts.IdentityRoleUser, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite:
		permissions = append(permissions, "api:*:*")
	default:
		return nil, fmt.Errorf("unknown identity role: %s", ntid)
	}
	// A device AAPS pairs as a client, by its QR code, reads everything and writes only the settings that client
	// control is made of, as its role does in Nightscout; whoever sees the code can read the site
	if client != nil && *client == sitemodels.AAPSClient {
		permissions = []string{"api:*:read", "api:settings:create", "api:settings:update"}
	}
	return &Identity{
		Name:        identityName,
		Permissions: permissions,
	}, nil
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
	if HasPermission(*identity, "api:treatments:create") {
		permissions = append(slices.Clone(permissions), "api:treatments:create")
	}
	return map[string]any{"token": token, "sub": siteContext.IdentityModel.Name, "permissionGroups": [][]string{permissions},
		"iat": time.Now().Unix(), "exp": expiresAt.Unix()}, nil
}
