package sitecollections

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
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

// A toast shows only what was written for users: a server error, or the database's own words, says it is on the
// Errors page, linked for an admin, and a hint shows without the message's start
func TestWriteErrorKeepsTheServersWordsFromToasts(t *testing.T) {
	tests := []struct {
		name  string
		herr  *apiutils.HTTPError
		role  consts.IdentityRole
		shows []string
		hides []string
	}{
		{"a server error, to an admin", &apiutils.HTTPError{Status: http.StatusBadGateway, Message: "Apple refused the push: 403 InvalidProviderToken"},
			consts.IdentityRoleAdmin, []string{"Something went wrong.", `href="/logs/errors"`}, []string{"InvalidProviderToken"}},
		{"a server error, to a user", apiutils.ErrInternalServerError, consts.IdentityRoleUser,
			[]string{"recorded for the site&#39;s admins"}, []string{"/logs/errors", "Internal server error"}},
		{"the database's words", apiutils.ErrDatabaseRefused(errors.New("CHECK constraint failed: length(notes) <= 1024")),
			consts.IdentityRoleAdmin, []string{"Something went wrong.", `href="/logs/errors"`}, []string{"CHECK", "Constraint check failed"}},
		{"a hint for users", apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.title_required")), consts.IdentityRoleUser,
			[]string{"title is required"}, []string{"Constraint check failed", "Something went wrong."}},
		{"a body hint for users", apiutils.ErrBadBodyWithHint(i18n.NewError("errors.data.day")), consts.IdentityRoleUser,
			[]string{"date must be a day"}, []string{"Bad or missing request body"}},
		{"a missing field", apiutils.ErrMissingRequiredFieldWithHint("errors.missing_field.carbs"), consts.IdentityRoleUser,
			[]string{"Missing required field carbs"}, []string{"Something went wrong."}},
		{"a library's error, as it is", apiutils.ErrBadBodyWithHint(errors.New("unexpected EOF")), consts.IdentityRoleUser,
			[]string{"unexpected EOF"}, []string{"Bad or missing request body"}},
		{"a message for users", &apiutils.HTTPError{Status: http.StatusGone, Message: "The app is no longer registered on this device"},
			consts.IdentityRoleUser, []string{"The app is no longer registered on this device"}, []string{"Something went wrong."}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, nil, "test", "", httputils.HTTPActionWrite, httputils.AcceptTypeHTMX, httputils.ContentTypeForm)
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: test.role})
			w := httptest.NewRecorder()
			writeError(w, httptest.NewRequest(http.MethodPost, "/data/glucose", nil).WithContext(ctx), httputils.AcceptTypeHTMX, test.herr)
			require.Equal(t, test.herr.Status, w.Code)
			for _, shown := range test.shows {
				require.Contains(t, w.Body.String(), shown)
			}
			for _, hidden := range test.hides {
				require.NotContains(t, w.Body.String(), hidden)
			}
		})
	}
}

// A page loaded in full, or opened in the main content, is in its hub's tabs, the latter with the header's links out of
// band to mark its hub, as are the account menu's; a part of a page, such as a table, a form's answer or a modal, is not
func TestWriteResponseWrapsPagesInTheirHub(t *testing.T) {
	tests := []struct {
		name     string
		accept   httputils.AcceptType
		target   string
		fullPage bool
		hub      bool
		oob      bool
	}{
		{"a page loaded in full", httputils.AcceptTypeHTML, "", true, true, false},
		{"a page opened in the main content", httputils.AcceptTypeHTMX, "main#main-content", true, true, true},
		{"a table of the page", httputils.AcceptTypeHTMX, "div#glucoseReadingsTable", true, false, false},
		{"a request with no target", httputils.AcceptTypeHTMX, "", true, false, false},
		{"a form's answer", httputils.AcceptTypeHTMX, "main#main-content", false, false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, nil, "test", "", httputils.HTTPActionRead, test.accept, httputils.ContentTypeJSON)
			ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin})
			r := httptest.NewRequest(http.MethodGet, "/data/insulin", nil).WithContext(ctx)
			if test.target != "" {
				r.Header.Set("HX-Target", test.target)
			}
			w := httptest.NewRecorder()
			WriteResponse(w, r, httputils.GetSiteContext(ctx), http.StatusOK, nil, templ.Raw(`<p id="page">Insulin</p>`), test.fullPage)
			body := w.Body.String()
			require.Contains(t, body, `<p id="page">Insulin</p>`)
			require.Equal(t, test.hub, strings.Contains(body, `id="site--hub--data"`), body)
			require.Equal(t, test.hub, strings.Contains(body, `id="site--data--insulin-button"`), body)
			require.Equal(t, test.oob, strings.Contains(body, `hx-swap-oob="true"`), body)
			require.Equal(t, test.oob, strings.Contains(body, `id="site--nav--menu" hx-swap-oob="true"`), body)
		})
	}
}
