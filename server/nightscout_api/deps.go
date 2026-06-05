package nightscout_api

import (
	"context"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/auth"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/store"
	"nighttroop.t1d.nz/utils/config"
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

func (nta *NightTroopAPIAuth) CreateRole(name string, permissions ...string) error {
	return nil
}

func (nta *NightTroopAPIAuth) CreateSubject(name string, roles []string) auth.Subject {
	return auth.Subject{}
}

func (nta *NightTroopAPIAuth) UpdateAPISecret(apiSecret string) {

}

func (nta *NightTroopAPIAuth) IssueJWT(accessToken string) (string, error) {
	return "", nil
}

func (nta *NightTroopAPIAuth) AuthenticateRequest(r *http.Request) (*auth.Identity, error) {
	return nil, nil
}

func (nta *NightTroopAPIAuth) Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *auth.Identity)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r, nil)
	}
}

func (nta *NightTroopAPIAuth) RequireV1Write(permission string, next func(http.ResponseWriter, *http.Request, *auth.Identity)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r, nil)
	}
}

func (nta *NightTroopAPIAuth) HasPermission(identity auth.Identity, permission string) bool {
	return false
}

func (nta *NightTroopAPIAuth) AuthenticateExplicit(r *http.Request) (*auth.Identity, error) {
	return nil, nil
}
