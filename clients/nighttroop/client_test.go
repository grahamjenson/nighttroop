package nighttroop_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nighttroop"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

// The generated client creates a site, signs in as its admin, and uses each collection in the schema
func TestClient(t *testing.T) {
	config, controlDB := testutils.NewControlDB(t)
	ts := httptest.NewServer(server.GetMainHandler(controlDB, config))
	t.Cleanup(ts.Close)
	ctx := context.Background()
	httpClient := nighttroop.WithHTTPClient(testutils.HostClient(ts))

	public, err := nighttroop.NewClientWithResponses("http://localhost", httpClient)
	require.NoError(t, err)
	created, err := public.CreateSiteWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, created.StatusCode(), string(created.Body))
	newSite := created.JSON200

	// A site's API is on its own host
	site, err := nighttroop.NewClientWithResponses("http://"+newSite.SiteSubdomain+".localhost", httpClient)
	require.NoError(t, err)

	// A wrong password is refused
	refused, err := site.LoginWithResponse(ctx, nighttroop.LoginRequest{Name: "admin", Password: "wrong", Totp: ptr("123123")})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, refused.StatusCode())

	// Tests' TOTP codes are always 123123
	login, err := site.LoginWithResponse(ctx, nighttroop.LoginRequest{Name: "admin", Password: newSite.ApiKey, Totp: ptr("123123")})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, login.StatusCode(), string(login.Body))
	auth := nighttroop.RequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+login.JSON200.Token)
		return nil
	})

	// Without the JWT, the site's data is refused
	unauthed, err := site.ListInsulinDosesWithResponse(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, unauthed.StatusCode())

	// An insulin dose is created, listed, read, updated and deleted
	dose, err := site.CreateInsulinDoseWithResponse(ctx, nighttroop.InsulinDose{Dose: ptr(4.5), Notes: ptr("Lunch")}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, dose.StatusCode(), string(dose.Body))
	id := *dose.JSON201.Result.Id
	require.Equal(t, 4.5, *dose.JSON201.Result.Dose)
	require.Equal(t, "admin", *dose.JSON201.Result.CreatedByIdentityName)

	doses, err := site.ListInsulinDosesWithResponse(ctx, &nighttroop.ListInsulinDosesParams{Limit: ptr(int64(10))}, auth)
	require.NoError(t, err)
	require.Len(t, *doses.JSON200.Result.Items, 1)
	require.Equal(t, int64(1), *doses.JSON200.Result.Total, "the search counts what it matches over every page")

	updated, err := site.UpdateInsulinDoseWithResponse(ctx, id, nighttroop.InsulinDose{Notes: ptr("Dinner")}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updated.StatusCode(), string(updated.Body))
	require.Equal(t, "Dinner", *updated.JSON200.Result.Notes)
	require.Equal(t, 4.5, *updated.JSON200.Result.Dose, "what is not sent stays as it is")

	deleted, err := site.DeleteInsulinDoseWithResponse(ctx, id, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, deleted.StatusCode(), string(deleted.Body))
	gone, err := site.GetInsulinDoseWithResponse(ctx, id, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, gone.StatusCode())

	// A meter reading, in the default mmol/L, and a meal
	reading, err := site.CreateGlucoseReadingWithResponse(ctx, nighttroop.GlucoseReading{Reading: ptr(6.5)}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, reading.StatusCode(), string(reading.Body))
	require.Equal(t, nighttroop.MmolL, *reading.JSON201.Result.Units)

	food, err := site.CreateFoodIntakeWithResponse(ctx, nighttroop.FoodIntake{Name: ptr("Toast"), Carbs: ptr(30.0)}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, food.StatusCode(), string(food.Body))

	readings, err := site.ListGlucoseReadingsWithResponse(ctx, nil, auth)
	require.NoError(t, err)
	require.Len(t, *readings.JSON200.Result.Items, 1)
	foods, err := site.ListFoodIntakesWithResponse(ctx, nil, auth)
	require.NoError(t, err)
	require.Equal(t, "Toast", *(*foods.JSON200.Result.Items)[0].Name)

	// The summary has the reading and the meal, and keeps what the schema does not describe
	summary, err := site.GetSummaryWithResponse(ctx, &nighttroop.GetSummaryParams{Hours: ptr(int64(6))}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, summary.StatusCode(), string(summary.Body))
	day := (*summary.JSON200.Result.Items)[0]
	require.Equal(t, int64(6), *day.Hours)
	require.Len(t, *day.GlucoseReadings, 1)
	require.Len(t, *day.FoodIntakes, 1)
	require.Contains(t, day.AdditionalProperties, "today")
}

// The generated client uses a site's Nightscout API, v1, v2 and v3, on its Nightscout host with its PWD phone's token
func TestNightscoutClient(t *testing.T) {
	config, controlDB := testutils.NewControlDB(t)
	ts := httptest.NewServer(server.GetMainHandler(controlDB, config))
	t.Cleanup(ts.Close)
	ctx := context.Background()
	httpClient := nighttroop.WithHTTPClient(testutils.HostClient(ts))

	public, err := nighttroop.NewClientWithResponses("http://localhost", httpClient)
	require.NoError(t, err)
	created, err := public.CreateSiteWithResponse(ctx)
	require.NoError(t, err)
	newSite := created.JSON200
	ns, err := nighttroop.NewClientWithResponses("http://"+newSite.SiteSubdomain+".ns.localhost", httpClient)
	require.NoError(t, err)

	// Without a token the API is refused
	unauthed, err := ns.NsV1StatusWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, unauthed.StatusCode())

	// The token itself, as the token query parameter, or the JWT it is exchanged for
	byToken, err := ns.NsV3TestWithResponse(ctx, func(_ context.Context, req *http.Request) error {
		query := req.URL.Query()
		query.Set("token", newSite.PwdNstoken)
		req.URL.RawQuery = query.Encode()
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, byToken.StatusCode(), string(byToken.Body))
	authorization, err := ns.NsV2AuthorizeWithResponse(ctx, newSite.PwdNstoken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, authorization.StatusCode(), string(authorization.Body))
	auth := nighttroop.RequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+*authorization.JSON200.Token)
		return nil
	})

	status, err := ns.NsV1StatusWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, "ok", (*status.JSON200)["status"])
	versions, err := ns.NsV1VersionsWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Len(t, *versions.JSON200, 3)
	version, err := ns.NsV3VersionWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, version.StatusCode())
	verified, err := ns.NsV1VerifyAuthWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, "OK", (*verified.JSON200)["message"].(map[string]any)["message"])
	experiments, err := ns.NsV1ExperimentsTestWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, "ok", (*experiments.JSON200)["status"])

	// v1: an entry is created, searched, the current one read, counted and deleted
	now := time.Now().UnixMilli()
	entries, err := ns.NsV1CreateWithResponse(ctx, nighttroop.Entries, []nighttroop.NsDocument{{Type: ptr("sgv"), Sgv: ptr(120.0), Date: ptr(now)}}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, entries.StatusCode(), string(entries.Body))
	entryID := *(*entries.JSON200)[0].UnderscoreId
	found, err := ns.NsV1SearchWithResponse(ctx, nighttroop.Entries, &nighttroop.NsV1SearchParams{Count: ptr(int64(10)), DateGte: ptr(now - 1000)}, auth)
	require.NoError(t, err)
	require.Len(t, *found.JSON200, 1)
	require.Equal(t, "mg/dL", *(*found.JSON200)[0].Units, "glucose is in mg/dL, whatever the identity's default mmol/L")
	require.InDelta(t, 120, *(*found.JSON200)[0].Sgv, 0.01)
	current, err := ns.NsV1CurrentWithResponse(ctx, nighttroop.Entries, auth)
	require.NoError(t, err)
	require.Equal(t, entryID, *(*current.JSON200)[0].UnderscoreId)
	byID, err := ns.NsV1GetWithResponse(ctx, nighttroop.Entries, entryID, nil, auth)
	require.NoError(t, err)
	require.Len(t, *byID.JSON200, 1)
	count, err := ns.NsV1CountWithResponse(ctx, nighttroop.Entries, nil, auth)
	require.NoError(t, err)
	require.EqualValues(t, 1, (*count.JSON200)["count"])

	// v3: a treatment is created, read, patched, found in a search and the history, then deleted
	bolus := nighttroop.NsDocument{EventType: ptr("Correction Bolus"), Insulin: ptr(2.0), Date: ptr(now)}
	bolus.Set("app", "client test")
	treatment, err := ns.NsV3CreateWithResponse(ctx, nighttroop.Treatments, bolus, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, treatment.StatusCode(), string(treatment.Body))
	identifier := *treatment.JSON201.Identifier
	patched, err := ns.NsV3PatchWithResponse(ctx, nighttroop.Treatments, identifier, nighttroop.NsDocument{Notes: ptr("Lunch")}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, patched.StatusCode(), string(patched.Body))
	read, err := ns.NsV3ReadWithResponse(ctx, nighttroop.Treatments, identifier, nil, auth)
	require.NoError(t, err)
	require.Equal(t, 2.0, *read.JSON200.Result.Insulin)
	require.Equal(t, "Lunch", *read.JSON200.Result.Notes)
	search, err := ns.NsV3SearchWithResponse(ctx, nighttroop.Treatments, &nighttroop.NsV3SearchParams{Limit: ptr(int64(10))}, auth)
	require.NoError(t, err)
	require.Len(t, *search.JSON200.Result, 1)
	history, err := ns.NsV3HistoryWithResponse(ctx, nighttroop.Treatments, 0, nil, auth)
	require.NoError(t, err)
	require.Len(t, *history.JSON200.Result, 1)
	lastModified, err := ns.NsV3LastModifiedWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, lastModified.StatusCode(), string(lastModified.Body))

	// v2: the properties and summary
	properties, err := ns.NsV2PropertiesNamedWithResponse(ctx, "bgnow", auth)
	require.NoError(t, err)
	require.Contains(t, *properties.JSON200, "bgnow")
	nsSummary, err := ns.NsV2SummaryWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, nsSummary.StatusCode(), string(nsSummary.Body))
	checked, err := ns.NsV2CheckPermissionWithResponse(ctx, "api:settings:create", auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, checked.StatusCode(), string(checked.Body))

	removed, err := ns.NsV3DeleteWithResponse(ctx, nighttroop.Treatments, identifier, nil, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, removed.StatusCode(), string(removed.Body))
	deleted, err := ns.NsV1DeleteWithResponse(ctx, nighttroop.Entries, entryID, nil, auth)
	require.NoError(t, err)
	require.EqualValues(t, 1, (*deleted.JSON200)["deletedCount"])
}

func ptr[T any](v T) *T {
	return &v
}
