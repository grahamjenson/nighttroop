package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/auth"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/internal/testsupport"
)

func testDeps() deps.Dependencies {
	return deps.Dependencies{
		Config: config.Config{
			APISecret:    "secret",
			Enable:       []string{"careportal", "rawbg"},
			DefaultRoles: []string{"readable"},
			AppVersion:   "1.0.0",
			API3MaxLimit: 100,
		}.WithDefaults(),
		Store: testsupport.NewMemoryStore(),
		Auth:  testsupport.NewAuthManager("secret", []string{"readable"}),
	}
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.RouteContext(req.Context())
	if routeCtx == nil {
		routeCtx = chi.NewRouteContext()
	}
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) any {
	t.Helper()
	var out any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v body=%s", err, rec.Body.String())
	}
	return out
}

func TestSearchHandlers(t *testing.T) {
	dep := testDeps()
	mem := dep.Store.(*testsupport.MemoryStore)
	_, _, _ = mem.Create(context.Background(), "entries", map[string]any{
		"date":       time.Now().UnixMilli(),
		"dateString": "2014-07-19T09:00:00",
		"device":     "sensor",
	}, "seed")

	rec := httptest.NewRecorder()
	Versions()(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Versions failed")
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/?find[sgv][$gte]=100", nil), "collection", "entries")
	EchoRoute(dep)(rec, req, &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("EchoRoute failed")
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodGet, "/?count=10", nil), "collection", "entries")
	req = withRouteParam(req, "field", "dateString")
	req = withRouteParam(req, "type", "sgv")
	req = withRouteParam(req, "prefix", "2014-07")
	SliceRoute(dep)(rec, req, &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("SliceRoute failed")
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodGet, "/", nil), "*", "2014-07/T09:.json")
	TimesEchoRoute(dep)(rec, req, &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TimesEchoRoute failed")
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodGet, "/", nil), "*", "2014-07/T09:.json")
	TimesRoute(dep)(rec, req, &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TimesRoute failed")
	}
}

func TestEntryTreatmentGenericProfileAndVerifyHandlers(t *testing.T) {
	dep := testDeps()
	mem := dep.Store.(*testsupport.MemoryStore)
	ctx := context.Background()
	entry, _, _ := mem.Create(ctx, "entries", map[string]any{
		"date":       time.Now().UnixMilli(),
		"dateString": "2024-01-01T00:00:00",
		"device":     "sensor",
	}, "seed")
	_, _, _ = mem.Create(ctx, "profile", map[string]any{
		"created_at": "2024-01-01T00:00:00.000Z",
	}, "seed")

	rec := httptest.NewRecorder()
	EntriesCurrent(dep)(rec, httptest.NewRequest(http.MethodGet, "/", nil), &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("EntriesCurrent failed")
	}
	rec = httptest.NewRecorder()
	EntriesList(dep)(rec, httptest.NewRequest(http.MethodGet, "/", nil), &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("EntriesList failed")
	}
	rec = httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/", nil), "spec", entry.Identifier())
	EntriesSpec(dep)(rec, req, &auth.Identity{Name: "test"})
	if rec.Code != http.StatusOK {
		t.Fatalf("EntriesSpec failed")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`[{"date":1700000000000,"dateString":"2024-01-01T00:00:00Z","device":"sensor"}]`))
	EntriesCreate(dep, false)(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("EntriesCreate preview failed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	req = withRouteParam(req, "spec", "sgv")
	EntriesDelete(dep)(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("EntriesDelete failed")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"eventType":"Meal Bolus","created_at":"2024-01-02T00:00:00.000Z","carbs":"10","app":"test","device":"pump"}`))
	TreatmentsCreate(dep)(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TreatmentsCreate failed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`[{"eventType":"Meal Bolus","created_at":"2024-01-02T00:00:00.000Z","carbs":"10","app":"test","device":"pump"}]`))
	TreatmentsCreate(dep)(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TreatmentsCreate array failed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	TreatmentsList(dep)(rec, httptest.NewRequest(http.MethodGet, "/", nil), &auth.Identity{Name: "reader"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TreatmentsList failed")
	}
	rec = httptest.NewRecorder()
	TreatmentsDelete(dep)(rec, httptest.NewRequest(http.MethodDelete, "/", nil), &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("TreatmentsDelete failed")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"created_at":"2024-01-02T00:00:00.000Z","device":"rig","app":"test"}`))
	GenericCollectionCreate(dep, "devicestatus")(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("GenericCollectionCreate failed")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`[{"created_at":"2024-01-02T00:00:00.000Z","device":"rig-2","app":"test"}]`))
	GenericCollectionCreate(dep, "devicestatus")(rec, req, &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("GenericCollectionCreate array failed")
	}
	rec = httptest.NewRecorder()
	GenericCollectionList(dep, "devicestatus", "created_at")(rec, httptest.NewRequest(http.MethodGet, "/", nil), &auth.Identity{Name: "reader"})
	if rec.Code != http.StatusOK {
		t.Fatalf("GenericCollectionList failed")
	}
	rec = httptest.NewRecorder()
	GenericCollectionDelete(dep, "devicestatus", "created_at")(rec, httptest.NewRequest(http.MethodDelete, "/", nil), &auth.Identity{Name: "writer"})
	if rec.Code != http.StatusOK {
		t.Fatalf("GenericCollectionDelete failed")
	}

	rec = httptest.NewRecorder()
	ProfileList(dep)(rec, httptest.NewRequest(http.MethodGet, "/", nil), &auth.Identity{Name: "reader"})
	if rec.Code != http.StatusOK {
		t.Fatalf("ProfileList failed")
	}

	rec = httptest.NewRecorder()
	VerifyAuth(dep)(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("VerifyAuth unauthorized failed")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("api-secret", "secret")
	VerifyAuth(dep)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("VerifyAuth authorized failed")
	}
}
