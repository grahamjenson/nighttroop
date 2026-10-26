package httputils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestRecordServerErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		panics  bool
		site    bool
		message string
	}{
		{"a panic is recorded, with its site", 0, true, true, "500 Internal Server Error"},
		{"a 503 is recorded, without a site", http.StatusServiceUnavailable, false, false, "503 Service Unavailable"},
		{"a 404 is not recorded", http.StatusNotFound, false, true, ""},
		{"a 200 is not recorded", http.StatusOK, false, true, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, controlDB, err := testutils.GetConfigControlDBForTest()
			require.NoError(t, err)
			router := chi.NewRouter()
			AddStandardMiddleware(router, cfg, controlDB.Logs)
			router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {
				if test.panics {
					panic("broken")
				}
				w.WriteHeader(test.status)
			})

			r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123", nil)
			if test.site {
				r = r.WithContext(CreateSiteContext(r.Context(), cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionRead, AcceptTypeJSON, ContentTypeJSON))
			}
			router.ServeHTTP(httptest.NewRecorder(), r)

			recorded, err := controlDB.Logs.Queries.SearchErrors(context.TODO(), logsdb.SearchErrorsParams{Limit: 10})
			require.NoError(t, err)
			if test.message == "" {
				require.Empty(t, recorded)
				return
			}
			require.Len(t, recorded, 1)
			require.Equal(t, test.message, recorded[0].Message)
			// The route's pattern is recorded, not the path with its token
			require.Contains(t, *recorded[0].Attributes, `"route":"/api/v2/authorization/request/{token}"`)
			require.NotContains(t, *recorded[0].Attributes, "pathtoken123")
			if test.site {
				require.Equal(t, "somesite", *recorded[0].SiteID)
			} else {
				require.Nil(t, recorded[0].SiteID)
			}
		})
	}
}
