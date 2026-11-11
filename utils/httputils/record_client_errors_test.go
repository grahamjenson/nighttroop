package httputils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestRecordClientErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		message string
	}{
		{"a 400 is recorded with its JSON's message", http.StatusBadRequest, `{"status":400,"message":"invalid collection"}`, "400 Bad Request: invalid collection"},
		{"a 422 is recorded with its text", http.StatusUnprocessableEntity, "glucose above 1000 mg/dL", "422 Unprocessable Entity: glucose above 1000 mg/dL"},
		{"a 405 is recorded without a body", http.StatusMethodNotAllowed, "", "405 Method Not Allowed"},
		{"a 401 is not recorded", http.StatusUnauthorized, `{"status":401,"message":"Missing or bad access token or JWT"}`, ""},
		{"a 404 is not recorded", http.StatusNotFound, "Not Found", ""},
		{"a 200 is not recorded", http.StatusOK, `{"message":"ok"}`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, controlDB := testutils.NewControlDB(t)
			router := chi.NewRouter()
			AddStandardMiddleware(router, cfg, controlDB)
			router.Post("/api/v1/{collection}", func(w http.ResponseWriter, r *http.Request) {
				SetIdentityModelInSiteContext(r.Context(), &controldb.SiteIdentity{Name: "pwd_phone", Role: consts.IdentityRoleDevicePWDPhone})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			})
			send := func() {
				r := httptest.NewRequest(http.MethodPost, "/api/v1/entries?token=querytoken123", nil)
				// The response is compressed, as it is for every client that accepts gzip
				r.Header.Set("Accept-Encoding", "gzip")
				r.Header.Set("User-Agent", "AAPS/3.3.1")
				r = r.WithContext(CreateSiteContext(r.Context(), cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionWrite, AcceptTypeJSON, ContentTypeJSON))
				router.ServeHTTP(httptest.NewRecorder(), r)
			}
			send()
			// The same error again within a minute is not recorded again
			send()

			recorded, err := controlDB.Logs.Queries.SearchErrors(t.Context(), logsdb.SearchErrorsParams{Limit: 10})
			require.NoError(t, err)
			if test.message == "" {
				require.Empty(t, recorded)
				return
			}
			require.Len(t, recorded, 1)
			require.Equal(t, test.message, recorded[0].Message)
			require.Equal(t, "somesite", *recorded[0].SiteID)
			attributes := *recorded[0].Attributes
			require.Contains(t, attributes, `"kind":"client"`)
			require.Contains(t, attributes, `"route":"/api/v1/{collection}"`)
			require.Contains(t, attributes, `"identity":"pwd_phone"`)
			require.Contains(t, attributes, `"role":"device_pwd_phone"`)
			require.Contains(t, attributes, `"client":"AAPS/3.3.1"`)
			require.NotContains(t, attributes, "querytoken123")
		})
	}
}

// sentEvents is a Sentry transport that keeps what it is sent
type sentEvents struct{ events []*sentry.Event }

func (s *sentEvents) Flush(time.Duration) bool              { return true }
func (s *sentEvents) FlushWithContext(context.Context) bool { return true }
func (s *sentEvents) Configure(sentry.ClientOptions)        {}
func (s *sentEvents) SendEvent(event *sentry.Event)         { s.events = append(s.events, event) }
func (s *sentEvents) Close()                                {}

// A client error goes to Sentry as a warning, grouped by its status, route and reason, with the identity's role and
// the client but not the identity's name or the request's token
func TestRecordClientErrorsReportsToSentry(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	sent := &sentEvents{}
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://key@sentry.example/1", Transport: sent, BeforeSend: RedactSentryEvent})
	require.NoError(t, err)
	router := chi.NewRouter()
	AddStandardMiddleware(router, cfg, controlDB)
	router.Get("/api/v3/{collection}", func(w http.ResponseWriter, r *http.Request) {
		SetIdentityModelInSiteContext(r.Context(), &controldb.SiteIdentity{Name: "pwd_phone", Role: consts.IdentityRoleDevicePWDPhone})
		http.Error(w, "unknown filter field", http.StatusBadRequest)
	})
	r := httptest.NewRequest(http.MethodGet, "/api/v3/entries?token=querytoken123&find[nope]=1", nil)
	r.Header.Set("User-Agent", "Trio/0.5")
	hub := sentry.NewHub(client, sentry.NewScope())
	// As the Sentry middleware sets them when Sentry is set up
	hub.Scope().SetRequest(r)
	ctx := sentry.SetHubOnContext(r.Context(), hub)
	r = r.WithContext(CreateSiteContext(ctx, cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionRead, AcceptTypeJSON, ContentTypeJSON))
	router.ServeHTTP(httptest.NewRecorder(), r)

	require.Len(t, sent.events, 1)
	event := sent.events[0]
	require.Equal(t, sentry.LevelWarning, event.Level)
	require.Equal(t, "400 GET /api/v3/{collection}: unknown filter field", event.Message)
	require.Equal(t, []string{"client_error", "400", "GET", "/api/v3/{collection}", "unknown filter field"}, event.Fingerprint)
	require.Equal(t, "device_pwd_phone", event.Tags["role"])
	require.Equal(t, "Trio/0.5", event.Tags["client"])
	for _, tag := range event.Tags {
		require.NotEqual(t, "pwd_phone", tag, "the identity's name is not sent")
	}
	require.NotContains(t, event.Request.QueryString, "querytoken123")
}

// A client whose bad requests differ, as by naming fields it made up, has at most clientErrorsPerMinute of its site's
// recorded a minute
func TestRecordClientErrorsAreCappedPerSite(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	router := chi.NewRouter()
	AddStandardMiddleware(router, cfg, controlDB)
	router.Get("/api/v1/entries", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad query parameters: "+r.URL.Query().Get("field"), http.StatusBadRequest)
	})
	for i := range clientErrorsPerMinute + 5 {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/entries?field=made-up-"+strconv.Itoa(i), nil)
		r = r.WithContext(CreateSiteContext(r.Context(), cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionRead, AcceptTypeJSON, ContentTypeJSON))
		router.ServeHTTP(httptest.NewRecorder(), r)
	}
	recorded, err := controlDB.Logs.Queries.SearchErrors(t.Context(), logsdb.SearchErrorsParams{Limit: 100})
	require.NoError(t, err)
	require.Len(t, recorded, clientErrorsPerMinute)
}
