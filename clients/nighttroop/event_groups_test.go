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

// The generated client groups a meal's reading, dose and food, picked from the events that can be grouped, then edits,
// reads, lists and deletes the group
func TestEventGroupsClient(t *testing.T) {
	site, auth := signedInSite(t)
	ctx := context.Background()

	reading, err := site.CreateGlucoseReadingWithResponse(ctx, nighttroop.GlucoseReading{Reading: ptr(6.5)}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, reading.StatusCode(), string(reading.Body))
	dose, err := site.CreateInsulinDoseWithResponse(ctx, nighttroop.InsulinDose{Dose: ptr(4.5)}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, dose.StatusCode(), string(dose.Body))
	food, err := site.CreateFoodIntakeWithResponse(ctx, nighttroop.FoodIntake{Name: ptr("Toast"), Carbs: ptr(30.0)}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, food.StatusCode(), string(food.Body))
	readingID, doseID, foodID := reading.JSON201.Result.Id, dose.JSON201.Result.Id, food.JSON201.Result.Id

	// The events to group are each kind a meal bolus can have, the reading in the identity's mmol/L
	choices, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: nighttroop.GroupMealBolus}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, choices.StatusCode(), string(choices.Body))
	byKind := map[nighttroop.EventKind]nighttroop.EventChoice{}
	for _, choice := range *choices.JSON200.Result {
		byKind[choice.Kind] = choice
	}
	require.Len(t, byKind, 3)
	require.Equal(t, *readingID, byKind[nighttroop.GlucoseReadingId].Id)
	require.InDelta(t, 6.5, *byKind[nighttroop.GlucoseReadingId].Amount, 0.01)
	require.Equal(t, 30.0, *byKind[nighttroop.FoodIntakeId].Amount)
	require.False(t, byKind[nighttroop.InsulinDoseId].Picked)

	// A site change can have only a glucose, a prime has no events to pick, and an unknown type is refused
	change, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: nighttroop.GroupSiteChange}, auth)
	require.NoError(t, err)
	require.Len(t, *change.JSON200.Result, 1)
	require.Equal(t, nighttroop.GlucoseReadingId, (*change.JSON200.Result)[0].Kind)
	prime, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: nighttroop.GroupPrime}, auth)
	require.NoError(t, err)
	require.Empty(t, *prime.JSON200.Result)
	unknown, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: "Unknown"}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, unknown.StatusCode())

	// The group is made of the events picked, and reads them back
	meal := nighttroop.EventGroup{GroupType: ptr(string(nighttroop.GroupMealBolus)), GlucoseReadingId: readingID, InsulinDoseId: doseID, FoodIntakeId: foodID, Notes: ptr("Lunch")}
	created, err := site.CreateEventGroupWithResponse(ctx, meal, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, created.StatusCode(), string(created.Body))
	group := created.JSON201.Result
	require.Equal(t, "Meal Bolus", *group.GroupType)
	require.Equal(t, nighttroop.MmolL, *group.GlucoseReading.Units)
	require.InDelta(t, 6.5, *group.GlucoseReading.Reading, 0.01)
	require.Equal(t, 4.5, *group.InsulinDose.Dose)
	require.Equal(t, "Toast", *group.FoodIntake.Name)

	// An event in a group is in no other, a kind the type cannot have is refused, and so is a group of no events
	again, err := site.CreateEventGroupWithResponse(ctx, nighttroop.EventGroup{GroupType: ptr("Bolus"), InsulinDoseId: doseID}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, again.StatusCode(), string(again.Body))
	require.NotEmpty(t, *again.JSON409.Message)
	wrongKind, err := site.CreateEventGroupWithResponse(ctx, nighttroop.EventGroup{GroupType: ptr("BG Check"), InsulinDoseId: doseID}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, wrongKind.StatusCode(), string(wrongKind.Body))
	none, err := site.CreateEventGroupWithResponse(ctx, nighttroop.EventGroup{GroupType: ptr("Meal Bolus")}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, none.StatusCode(), string(none.Body))

	// Grouped events are not offered to another group, only to their own, picked
	taken, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: nighttroop.GroupMealBolus}, auth)
	require.NoError(t, err)
	require.Empty(t, *taken.JSON200.Result)
	own, err := site.ListEventGroupChoicesWithResponse(ctx, &nighttroop.ListEventGroupChoicesParams{GroupType: nighttroop.GroupMealBolus, GroupId: group.Id}, auth)
	require.NoError(t, err)
	require.Len(t, *own.JSON200.Result, 3)
	for _, choice := range *own.JSON200.Result {
		require.True(t, choice.Picked, choice.Kind)
	}

	groups, err := site.ListEventGroupsWithResponse(ctx, &nighttroop.ListEventGroupsParams{Limit: ptr(int64(10))}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, groups.StatusCode(), string(groups.Body))
	require.Len(t, *groups.JSON200.Result.Items, 1)
	require.Equal(t, int64(1), *groups.JSON200.Result.Total)

	// An edit keeps only the events sent, taking the others out of the group
	edited, err := site.UpdateEventGroupWithResponse(ctx, *group.Id, nighttroop.EventGroup{InsulinDoseId: doseID, Notes: ptr("Dinner")}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, edited.StatusCode(), string(edited.Body))
	require.Equal(t, "Dinner", *edited.JSON200.Result.Notes)
	require.Equal(t, "Meal Bolus", *edited.JSON200.Result.GroupType, "the type stays when not sent")
	require.Nil(t, edited.JSON200.Result.FoodIntakeId)
	require.Nil(t, edited.JSON200.Result.GlucoseReading)

	read, err := site.GetEventGroupWithResponse(ctx, *group.Id, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, read.StatusCode(), string(read.Body))
	require.Equal(t, *doseID, *read.JSON200.Result.InsulinDoseId)

	// Deleting the group deletes its events, not those taken out of it
	deleted, err := site.DeleteEventGroupWithResponse(ctx, *group.Id, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, deleted.StatusCode(), string(deleted.Body))
	gone, err := site.GetEventGroupWithResponse(ctx, *group.Id, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, gone.StatusCode())
	goneDose, err := site.GetInsulinDoseWithResponse(ctx, *doseID, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, goneDose.StatusCode())
	kept, err := site.GetFoodIntakeWithResponse(ctx, *foodID, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, kept.StatusCode())

	// A change is a group of no events
	siteChange, err := site.CreateEventGroupWithResponse(ctx, nighttroop.EventGroup{GroupType: ptr(string(nighttroop.GroupSiteChange))}, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, siteChange.StatusCode(), string(siteChange.Body))
}

// signedInSite is a new site's client, and its admin's JWT for each request
func signedInSite(t *testing.T) (*nighttroop.ClientWithResponses, nighttroop.RequestEditorFn) {
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
	site, err := nighttroop.NewClientWithResponses("http://"+created.JSON200.SiteSubdomain+".localhost", httpClient)
	require.NoError(t, err)
	login, err := site.LoginWithResponse(ctx, nighttroop.LoginRequest{Name: "admin", Password: created.JSON200.ApiKey, Totp: ptr("123123")})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, login.StatusCode(), string(login.Body))
	return site, func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+login.JSON200.Token)
		return nil
	}
}
