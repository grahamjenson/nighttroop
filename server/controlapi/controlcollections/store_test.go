package controlcollections

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestControlStoreRoutes checks what the routes answer before reaching the database
func TestControlStoreRoutes(t *testing.T) {
	tests := []struct {
		name    string
		role    consts.ControlRole
		method  string
		path    string
		status  int
		message string
	}{
		{"read only cannot disable a site", consts.ControlRoleRead, http.MethodPost, "/sites/some-site", http.StatusForbidden, ""},
		{"read only cannot hydrate a site", consts.ControlRoleRead, http.MethodPost, "/archived-sites/some-site", http.StatusForbidden, ""},
		{"jobs cannot be updated", consts.ControlRoleAdmin, http.MethodPost, "/jobs/1", http.StatusMethodNotAllowed, "jobs cannot be updated here"},
		{"job stats cannot be updated", consts.ControlRoleAdmin, http.MethodPost, "/job-stats/day", http.StatusMethodNotAllowed, "job-stats cannot be updated here"},
		{"connectors cannot be updated", consts.ControlRoleAdmin, http.MethodPost, "/connectors/site::virtual", http.StatusMethodNotAllowed, "connectors cannot be updated here"},
	}

	r := chi.NewRouter()
	NewControlStore().Routes(r)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(`{"state":"disabled"}`))
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			ctx := httputils.CreateControlContext(req.Context(), nil, "", httputils.HTTPActionWrite)
			ctx = httputils.SetAdminControlContext(ctx, &controldb.ControlIdentity{Name: "tester", Role: test.role}, config.ServerStatus{})
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req.WithContext(ctx))

			require.Equal(t, test.status, w.Code)
			require.Contains(t, w.Body.String(), test.message)
		})
	}
}
