package sitecollections

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// readFailingCollection is a collection whose rows cannot be read, with err
type readFailingCollection struct{ err error }

func (readFailingCollection) GetName() string  { return "glucose_readings" }
func (readFailingCollection) GetRoute() string { return "glucose" }
func (readFailingCollection) ToModel(row *struct{}, _ *httputils.SiteContext) *struct{} {
	return row
}
func (c readFailingCollection) GetRow(context.Context, *siteapiutils.CollectionTx, string, *httputils.SiteContext) (*struct{}, error) {
	return nil, c.err
}
func (readFailingCollection) DeleteRow(context.Context, *siteapiutils.CollectionTx, string, *httputils.SiteContext) error {
	return nil
}
func (readFailingCollection) DeleteHTMXResponse(*struct{}) templ.Component { return nil }

// A delete answers why the row could not be read: not found when it is missing, and the database's error otherwise
func TestDeleteAnswersTheReadError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"missing", sql.ErrNoRows, http.StatusNotFound},
		{"database error", errors.New("disk I/O error"), http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Delete("/data/glucose/{id}", DeleteHandler[struct{}, struct{}](readFailingCollection{err: test.err}))
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, nil, "test", "", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/data/glucose/1", nil).WithContext(ctx))
			require.Equal(t, test.want, w.Code)
		})
	}
}
