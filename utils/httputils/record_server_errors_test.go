package httputils

import (
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
		reason  string
		message string
	}{
		{"a panic is recorded, with its site", 0, true, true, "", "500 Internal Server Error"},
		{"a 503 is recorded, without a site", http.StatusServiceUnavailable, false, false, "", "503 Service Unavailable"},
		{"a 502 is recorded with why, which its page need not say", http.StatusBadGateway, false, true, "Apple refused the push", "502 Bad Gateway: Apple refused the push"},
		{"a 400 is recorded with why, which its page need not say", http.StatusBadRequest, false, true, "CHECK constraint failed", "400 Bad Request: CHECK constraint failed"},
		{"a 404 is not recorded", http.StatusNotFound, false, true, "", ""},
		{"a 200 is not recorded", http.StatusOK, false, true, "", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, controlDB := testutils.NewControlDB(t)
			router := chi.NewRouter()
			AddStandardMiddleware(router, cfg, controlDB)
			router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {
				if test.panics {
					panic("broken")
				}
				SetErrorReason(r.Context(), test.reason)
				w.WriteHeader(test.status)
				w.Write([]byte("Something went wrong."))
			})

			r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123", nil)
			if test.site {
				r = r.WithContext(CreateSiteContext(r.Context(), cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionRead, AcceptTypeJSON, ContentTypeJSON))
			}
			router.ServeHTTP(httptest.NewRecorder(), r)

			recorded, err := controlDB.Logs.Queries.SearchErrors(t.Context(), logsdb.SearchErrorsParams{Limit: 10})
			require.NoError(t, err)
			if test.message == "" {
				require.Empty(t, recorded)
				return
			}
			require.Len(t, recorded, 1)
			require.Equal(t, test.message, recorded[0].Message)
			// The route's pattern is recorded, not the path with its token
			require.Contains(t, *recorded[0].Attributes, map[bool]string{true: `"kind":"server"`, false: `"kind":"client"`}[test.status != http.StatusBadRequest])
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
