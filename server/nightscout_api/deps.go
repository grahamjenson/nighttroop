package nightscout_api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/auth"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/store"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type NightTroopAPIStore struct {
	controlDB *db.ControlDB
	config    config.ServerConfig
}

func (nts *NightTroopAPIStore) Create(ctx context.Context, collection string, data map[string]any, subject string) (model.Record, bool, error) {
	return model.Record{}, false, nil
}

func (nts *NightTroopAPIStore) Get(ctx context.Context, collection, identifier string) (model.Record, error) {
	return model.Record{ID: "asdasd"}, nil
}

func (nts *NightTroopAPIStore) Search(ctx context.Context, collection string, query store.Query) ([]model.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	fmt.Println("UserContext in Search:", userContext)
	return []model.Record{{ID: "asdasd"}}, nil
}

func (nts *NightTroopAPIStore) Replace(ctx context.Context, collection, identifier string, data map[string]any, subject string) (model.Record, bool, error) {
	return model.Record{}, false, nil
}

func (nts *NightTroopAPIStore) Patch(ctx context.Context, collection, identifier string, patch map[string]any, subject string) (model.Record, error) {
	return model.Record{}, nil
}

func (nts *NightTroopAPIStore) Delete(ctx context.Context, collection, identifier string, permanent bool, subject string) error {
	return nil
}

func (nts *NightTroopAPIStore) DeleteMatching(ctx context.Context, collection string, query store.Query, permanent bool, subject string) (int, error) {
	return 0, nil
}

func (nts *NightTroopAPIStore) History(ctx context.Context, collection string, since int64, limit int) ([]model.Record, error) {
	return nil, nil
}

func (nts *NightTroopAPIStore) LastModified(ctx context.Context, collections []string) (map[string]int64, error) {
	return nil, nil
}

type NightTroopAPIAuth struct {
	controlDB *db.ControlDB
	config    config.ServerConfig
}

func (m *NightTroopAPIAuth) Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *auth.Identity)) http.HandlerFunc {
	// Collections
	// 'entries'
	// 'treatments'
	// 'devicestatus'
	// 'profile'
	// 'settings'
	// 'food'

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
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := m.AuthenticateExplicit(r)
		if err != nil {
			fmt.Println("WRITING UNAUTH")
			writeUnauthorized(w, err)
			return
		}

		if !m.HasPermission(*identity, permission) {
			fmt.Println("WRITING FORBIDDEN")
			writeForbidden(w, permission)
			return
		}
		next(w, r, identity)
	}
}

func (m *NightTroopAPIAuth) HasPermission(identity auth.Identity, permission string) bool {
	for _, granted := range identity.Permissions {
		if permissionMatch(granted, permission) {
			return true
		}
	}
	return false
}

func (m *NightTroopAPIAuth) AuthenticateExplicit(r *http.Request) (*auth.Identity, error) {
	userContext := httputils.GetUserContext(r.Context())
	if userContext == nil {
		return nil, auth.ErrMissingCredentials
	}

	ntIdentity := NighttrropIdentityToNightscoutIdentity(userContext.IdentityName, userContext.IdentityRole)
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
	if err == auth.ErrBadToken || err == auth.ErrBadJWT {
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

func NighttrropIdentityToNightscoutIdentity(identityName string, ntid consts.IdentityRole) *auth.Identity {
	permissions := []string{}
	switch ntid {
	case consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead:
		permissions = append(permissions, "api:*:read")
	case consts.IdentityRoleUser, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite:
		permissions = append(permissions, "api:*:*")
	default:
		return nil
	}
	return &auth.Identity{
		Name:        identityName,
		Permissions: permissions,
	}
}
