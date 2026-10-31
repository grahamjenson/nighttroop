package nighttroop_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
	require.Len(t, *doses.JSON200.Result, 1)

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
	require.Len(t, *readings.JSON200.Result, 1)
	foods, err := site.ListFoodIntakesWithResponse(ctx, nil, auth)
	require.NoError(t, err)
	require.Equal(t, "Toast", *(*foods.JSON200.Result)[0].Name)

	// The summary has the reading and the meal, and keeps what the schema does not describe
	summary, err := site.GetSummaryWithResponse(ctx, &nighttroop.GetSummaryParams{Hours: ptr(int64(6))}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, summary.StatusCode(), string(summary.Body))
	day := (*summary.JSON200.Result)[0]
	require.Equal(t, int64(6), *day.Hours)
	require.Len(t, *day.GlucoseReadings, 1)
	require.Len(t, *day.FoodIntakes, 1)
	require.Contains(t, day.AdditionalProperties, "today")
}

func ptr[T any](v T) *T {
	return &v
}
