package controlcollections

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestPushCredentialsForEverySite checks a credential for every site, which cannot be added here, is listed without
// its secret and can be deleted, and that only an admin adds or deletes one
func TestPushCredentialsForEverySite(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	ctx, collection := t.Context(), &PushCredentialsCollection{}
	router := chi.NewRouter()
	router.Get("/push-credentials", SearchHandler(collection))
	router.With(AdminOnly).Post("/push-credentials", collection.CreateHandler)
	router.With(AdminOnly).Delete("/push-credentials/{id}", collection.DeleteHandler)
	request := func(role consts.ControlRole, method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	secretEnc := "the encrypted secret"
	server, err := controlDB.Queries.CreatePushCredential(ctx, controldb.CreatePushCredentialParams{Name: "T1D Carer",
		Provider: consts.PushProviderFCM, TeamOrProjectID: "t1dcarer", SecretEnc: &secretEnc, AppsJson: `["t1dcarer"]`, CreatedByIdentityName: "admin"})
	require.NoError(t, err)

	list := request(consts.ControlRoleRead, http.MethodGet, "/push-credentials", "")
	require.Equal(t, http.StatusOK, list.Code)
	require.Contains(t, list.Body.String(), `"name":"T1D Carer"`)
	require.Contains(t, list.Body.String(), `"site_id":null`)
	require.NotContains(t, list.Body.String(), "secret")

	for _, body := range []string{`{"name": "T1D Carer"}`, `{"site_id": "", "name": "T1D Carer"}`} {
		w := request(consts.ControlRoleAdmin, http.MethodPost, "/push-credentials", body)
		require.Equal(t, http.StatusBadRequest, w.Code, body)
		require.Contains(t, w.Body.String(), "a credential for every site cannot be added", body)
	}
	require.Equal(t, http.StatusForbidden, request(consts.ControlRoleRead, http.MethodPost, "/push-credentials", `{"name": "x"}`).Code)

	path := "/push-credentials/" + strconv.FormatInt(server.ID, 10)
	require.Equal(t, http.StatusForbidden, request(consts.ControlRoleRead, http.MethodDelete, path, "").Code)
	deleted := request(consts.ControlRoleAdmin, http.MethodDelete, path, "")
	require.Equal(t, http.StatusOK, deleted.Code)
	require.NotContains(t, deleted.Body.String(), "secret")
	_, err = controlDB.Queries.GetAnyPushCredential(ctx, controldb.GetAnyPushCredentialParams{ID: server.ID})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
