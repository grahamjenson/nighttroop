package siteapiutils

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
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

func TestAnswer(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *apiutils.HTTPError
	}{
		{"no error", nil, nil},
		{"the store's own answer", apiutils.ErrMissingRequiredFieldWithHint("carbs"), apiutils.ErrMissingRequiredFieldWithHint("carbs")},
		{"no row", sql.ErrNoRows, apiutils.ErrNotFound},
		{"no row, wrapped", errors.Join(errors.New("updating"), sql.ErrNoRows), apiutils.ErrNotFound},
		{"anything else", errors.New("disk I/O error"), apiutils.ErrInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, answer(test.err, "test rows"))
		})
	}
}

// readFailingStore is a store whose rows cannot be read, with err
type readFailingStore struct{ err error }

func (readFailingStore) GetName() string                                           { return "glucose_readings" }
func (readFailingStore) ToModel(row *struct{}, _ *httputils.SiteContext) *struct{} { return row }
func (s readFailingStore) GetRow(context.Context, *sitedb.Queries, string, *httputils.SiteContext) (*struct{}, error) {
	return nil, s.err
}
func (readFailingStore) DeleteRow(context.Context, *sitedb.Queries, string, *httputils.SiteContext) error {
	return nil
}
func (readFailingStore) DeleteHTMXResponse(*struct{}) templ.Component { return nil }

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
			store := &Rows[struct{}, struct{}, *sitedb.Queries]{
				store:   readFailingStore{err: test.err},
				queries: func(*StoreTx) (*sitedb.Queries, error) { return nil, nil },
			}
			router := chi.NewRouter()
			router.Delete("/data/glucose/{id}", DeleteHandler(store))
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, nil, "test", "", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/data/glucose/abc", nil).WithContext(ctx))
			require.Equal(t, test.want, w.Code)
		})
	}
}
