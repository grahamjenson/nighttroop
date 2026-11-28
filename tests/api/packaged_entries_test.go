package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/testutils"
)

// This file is after nightscout_test.go by name, as the test config is loaded once, with the sign in limits TestNS sets.
//
// Where the server deletes the rows of old days, the Nightscout API still gives their CGM readings, made from the
// days' cold glucose: an app that reads months back gets each 5 minutes, with the rows that are left, each once.
func TestEntriesOfPackagedDaysAreMadeFromColdDays(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.RetentionDeleteRows, cfg.RetentionHotDays = false, retention.DefaultHotDays })
	router, q, ctx := server.GetMainHandler(controlDB, cfg), site.SiteDB.Queries(), t.Context()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	old, recent := today.AddDate(0, 0, -40), today.AddDate(0, 0, -2)
	write := func(id string, at time.Time, mgdl float64, readingType string, notes *string) {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: mgdl, ReadingType: readingType, OccurredAt: at.UnixMilli(), Notes: notes, CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	// 40 days ago: ten readings from midnight, 100 to 109, a second late in the 5 minutes of each; one with a note at
	// 02:00; and a meter's at 03:00. 2 days ago: three readings.
	for i := range 10 {
		write(fmt.Sprint("old", i), old.Add(time.Duration(i)*5*time.Minute+time.Second), float64(100+i), "sgv", nil)
	}
	write("noted", old.Add(2*time.Hour), 150, "sgv", new("felt low"))
	write("meter", old.Add(3*time.Hour), 160, "mbg", nil)
	for i := range 3 {
		write(fmt.Sprint("recent", i), recent.Add(time.Duration(i)*5*time.Minute), float64(200+i), "sgv", nil)
	}

	// entries are the entries a search gives: each one's ID, date, glucose and type
	type entry struct {
		ID   string  `json:"_id"`
		Date int64   `json:"date"`
		Sgv  float64 `json:"sgv"`
		Mbg  float64 `json:"mbg"`
		Type string  `json:"type"`
	}
	entries := func(params url.Values) []entry {
		params.Set("token", site.PwdNSToken)
		req := httptest.NewRequest("GET", "http://"+site.SiteDB.Subdomain+".ns.localhost/api/v1/entries.json?"+params.Encode(), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		var found []entry
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &found), rr.Body.String())
		return found
	}
	ids := func(found []entry) []string {
		list := []string{}
		for _, e := range found {
			list = append(list, e.ID)
		}
		return list
	}
	all := url.Values{"count": {"100"}}
	before := entries(all)
	require.Len(t, before, 15)

	// The old day's rows are deleted, but the one with a note and the meter's
	cfg.RetentionDeleteRows = true
	_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	rows, err := q.AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 100})
	require.NoError(t, err)
	require.Len(t, rows, 5)

	// A search still has every reading, the newest first: the old day's are made from its cold day, each at the start
	// of its 5 minutes, with an ID shaped as a Nightscout one that is the same each time
	found := entries(all)
	require.Len(t, found, 15)
	require.Equal(t, []string{"recent2", "recent1", "recent0", "meter", "noted"}, ids(found)[:5])
	for i, cold := range found[5:] {
		step := old.Add(time.Duration(9-i) * 5 * time.Minute).UnixMilli()
		require.Equal(t, entry{ID: fmt.Sprintf("%08xc01dc01dc01dc01d", step/1000), Date: step, Sgv: float64(109 - i), Type: "sgv"}, cold)
		require.Len(t, cold.ID, 24)
	}
	require.Equal(t, found, entries(all))
	// The reading with a note is a row and is in the cold day: it is given once, as its row
	require.Equal(t, 1, strings.Count(strings.Join(ids(found), ","), "noted"))
	require.NotContains(t, ids(found), fmt.Sprintf("%08xc01dc01dc01dc01d", old.Add(2*time.Hour).Unix()))

	// By dates, by glucose, and a page of them
	ms := func(at time.Time) string { return fmt.Sprint(at.UnixMilli()) }
	inRange := entries(url.Values{"count": {"100"}, "find[date][$gte]": {ms(old.Add(10 * time.Minute))}, "find[date][$lt]": {ms(old.Add(25 * time.Minute))}})
	require.Len(t, inRange, 3)
	require.Equal(t, []float64{104, 103, 102}, []float64{inRange[0].Sgv, inRange[1].Sgv, inRange[2].Sgv})
	high := entries(url.Values{"count": {"100"}, "find[sgv][$gte]": {"108"}, "find[date][$lt]": {ms(old.Add(24 * time.Hour))}})
	// A search by glucose is of every type's, as it is of rows: the meter's 160, then the noted 150, then the cold day's
	require.Len(t, high, 4)
	require.Equal(t, []string{"meter", "noted"}, ids(high)[:2])
	require.Equal(t, []float64{109, 108}, []float64{high[2].Sgv, high[3].Sgv})
	require.Equal(t, found[3:9], entries(url.Values{"count": {"6"}, "skip": {"3"}}))
	require.Equal(t, found[:2], entries(url.Values{"count": {"2"}}))

	// A search a made reading cannot match is of rows only: a meter's readings, or those of a device
	require.Equal(t, []string{"meter"}, ids(entries(url.Values{"count": {"100"}, "find[type]": {"mbg"}})))
	require.Empty(t, entries(url.Values{"count": {"100"}, "find[device]": {"nighttroop"}}))

	// A search the oldest first, as API v3's sort=date, has them too: the old day's ten, then the two rows of it, then
	// the recent day's, and a page of them
	oldest := func(params string) []string {
		req := httptest.NewRequest("GET", "http://"+site.SiteDB.Subdomain+".ns.localhost/api/v3/entries?sort=date&"+params+"&token="+site.PwdNSToken, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		var page struct {
			Result []struct {
				Identifier string `json:"identifier"`
			} `json:"result"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &page), rr.Body.String())
		list := []string{}
		for _, e := range page.Result {
			list = append(list, e.Identifier)
		}
		return list
	}
	ascending := slices.Clone(ids(found))
	slices.Reverse(ascending)
	require.Equal(t, ascending, oldest("limit=100"))
	require.Equal(t, ascending[:4], oldest("limit=4"))
	require.Equal(t, ascending[8:12], oldest("limit=4&skip=8"))

	// A made reading is read by its ID, and deleted by it: its step of the cold day is cleared, so it is in no search
	// and is not found again. An ID of a step with no reading is not found.
	ns := site.SiteDB.Subdomain + ".ns.localhost"
	one := "/api/v1/entries/" + found[5].ID + "?token=" + site.PwdNSToken
	status, got, _ := testutils.JSONRequest[[]entry](t, router, http.MethodGet, "http://"+ns+one, nil)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, []entry{found[5]}, got)
	status, _, _ = testutils.JSONRequest[any](t, router, http.MethodGet, "http://"+ns+"/api/v1/entries/"+fmt.Sprintf("%08xc01dc01dc01dc01d", old.Add(12*time.Hour).Unix())+"?token="+site.PwdNSToken, nil)
	require.Equal(t, http.StatusNotFound, status)
	status, _, _ = testutils.JSONRequest[any](t, router, http.MethodDelete, "http://"+ns+one, nil)
	require.Equal(t, http.StatusOK, status)
	status, _, _ = testutils.JSONRequest[any](t, router, http.MethodGet, "http://"+ns+one, nil)
	require.Equal(t, http.StatusNotFound, status)
	require.Equal(t, append(ids(found[:5]), ids(found[6:])...), ids(entries(all)))
	// Packaged again, the day keeps the step clear
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.Equal(t, append(ids(found[:5]), ids(found[6:])...), ids(entries(all)))

	// A server that does not delete rows reads only rows, as all of them are there
	cfg.RetentionDeleteRows = false
	require.Equal(t, []string{"recent2", "recent1", "recent0", "meter", "noted"}, ids(entries(all)))
	status, _, _ = testutils.JSONRequest[any](t, router, http.MethodGet, "http://"+ns+"/api/v1/entries/"+found[6].ID+"?token="+site.PwdNSToken, nil)
	require.Equal(t, http.StatusNotFound, status)
}

// A site's admin looks at what the site keeps of its old days: each cold day with what its bytes say they are, and one
// day with each of its values. It is only read, and only by an admin.
func TestColdDaysAreLookedAtByAnAdmin(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	router, q, ctx := server.GetMainHandler(controlDB, cfg), site.SiteDB.Queries(), t.Context()
	host, day := site.SiteDB.Subdomain+".localhost", time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -3)
	call := func(method, host, path, token, body string) (int, map[string]any) {
		req := httptest.NewRequest(method, "http://"+host+path, strings.NewReader(body))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		var answer map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &answer)
		return rr.Code, answer
	}
	status, signedIn := call("POST", host, "/login", "", `{"name": "admin", "password": "`+site.AdminPassword+`", "totp": "123123"}`)
	require.Equal(t, http.StatusOK, status)
	admin := signedIn["token"].(string)
	status, phone := call("GET", site.SiteDB.Subdomain+".ns.localhost", "/api/v2/authorization/request/token="+site.PwdNSToken, "", "")
	require.Equal(t, http.StatusOK, status)

	// No day is packaged yet
	status, none := call("GET", host, "/usage/cold-days", admin, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, map[string]any{"days": []any{}, "total": 0.0, "total_bytes": 0.0}, none["result"])

	// Two readings of a day, at 00:05 and 12:00, and one of the day after, are packaged
	for i, at := range []time.Duration{5 * time.Minute, 12 * time.Hour, 24 * time.Hour} {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint("r", i), Mgdl: float64(100 + 10*i), ReadingType: "sgv", OccurredAt: day.Add(at).UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)

	// The list is the newest day first, each with what it is, and without its values
	status, listed := call("GET", host, "/usage/cold-days", admin, "")
	require.Equal(t, http.StatusOK, status)
	result := listed["result"].(map[string]any)
	require.Equal(t, 2.0, result["total"])
	days := result["days"].([]any)
	require.Len(t, days, 2)
	first, second := days[0].(map[string]any), days[1].(map[string]any)
	require.Equal(t, day.AddDate(0, 0, 1).Format(time.DateOnly), first["date"])
	require.Equal(t, result["total_bytes"], first["bytes"].(float64)+second["bytes"].(float64))
	packagedAt := second["packaged_at"]
	require.Positive(t, packagedAt)
	delete(second, "packaged_at")
	delete(second, "bytes")
	delete(second, "second_order")
	delete(second, "deflated")
	require.Equal(t, map[string]any{"kind": "glucose", "date": day.Format(time.DateOnly), "version": 1.0, "shape": "series", "step_seconds": 300.0,
		"scale": 1.0, "steps": 288.0, "values": 2.0, "min": 100.0, "max": 110.0, "error": nil}, second)
	// A page of them
	status, paged := call("GET", host, "/usage/cold-days?limit=1&skip=1", admin, "")
	require.Equal(t, http.StatusOK, status)
	require.Len(t, paged["result"].(map[string]any)["days"], 1)
	require.Equal(t, day.Format(time.DateOnly), paged["result"].(map[string]any)["days"].([]any)[0].(map[string]any)["date"])

	// One day has each value at the time its step starts
	status, opened := call("GET", host, "/usage/cold-days/glucose/"+day.Format(time.DateOnly), admin, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, []any{
		map[string]any{"at": float64(day.Add(5 * time.Minute).UnixMilli()), "value": 100.0},
		map[string]any{"at": float64(day.Add(12 * time.Hour).UnixMilli()), "value": 110.0},
	}, opened["result"].(map[string]any)["points"])
	status, _ = call("GET", host, "/usage/cold-days/glucose/1999-01-01", admin, "")
	require.Equal(t, http.StatusNotFound, status)

	// Only an admin looks, and no one writes: there is no route to make, change or delete a cold day
	status, _ = call("GET", host, "/usage/cold-days", phone["token"].(string), "")
	require.Equal(t, http.StatusForbidden, status)
	status, _ = call("GET", host, "/usage/cold-days/glucose/"+day.Format(time.DateOnly), phone["token"].(string), "")
	require.Equal(t, http.StatusForbidden, status)
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		for _, path := range []string{"/usage/cold-days", "/usage/cold-days/glucose/" + day.Format(time.DateOnly)} {
			status, _ = call(method, host, path, admin, `{}`)
			require.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed}, status, method+" "+path)
		}
	}
}

// Where the server deletes the rows of old days, the Nightscout API still gives their treatments, made from the days'
// cold events: each with its own ID, type, time and values, the newest first with the treatments that are rows.
func TestTreatmentsOfPackagedDaysAreMadeFromColdDays(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.RetentionDeleteRows = false })
	router, q, ctx := server.GetMainHandler(controlDB, cfg), site.SiteDB.Queries(), t.Context()
	nsHost := "http://" + site.SiteDB.Subdomain + ".ns.localhost"
	today := time.Now().UTC().Truncate(24 * time.Hour)
	old, recent := today.AddDate(0, 0, -40), today.AddDate(0, 0, -2)
	iso := func(at time.Time) string { return at.Format(time.RFC3339) }
	do := func(method, path, body string) (int, []byte) {
		req := httptest.NewRequest(method, nsHost+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr.Code, rr.Body.Bytes()
	}
	// 40 days ago: a meal with its bolus, one treatment of two events; a temp target; a site change, which has no
	// event; and a note. 2 days ago: a correction.
	status, body := do("POST", "/api/v1/treatments?token="+site.PwdNSToken, `[
		{"eventType": "Meal Bolus", "carbs": 45, "insulin": 4.5, "notes": "Pasta", "enteredBy": "xDrip", "created_at": "`+iso(old.Add(1*time.Hour))+`"},
		{"eventType": "Temporary Target", "targetTop": 160, "targetBottom": 140, "units": "mg/dl", "duration": 60, "reason": "Exercise", "created_at": "`+iso(old.Add(2*time.Hour))+`"},
		{"eventType": "Site Change", "enteredBy": "xDrip", "created_at": "`+iso(old.Add(3*time.Hour))+`"},
		{"eventType": "Note", "notes": "Felt fine", "created_at": "`+iso(old.Add(4*time.Hour))+`"},
		{"eventType": "Correction Bolus", "insulin": 1.5, "created_at": "`+iso(recent.Add(time.Hour))+`"}]`)
	require.Equal(t, http.StatusOK, status, string(body))

	treatments := func(params url.Values) []map[string]any {
		params.Set("token", site.PwdNSToken)
		status, body := do("GET", "/api/v1/treatments.json?"+params.Encode(), "")
		require.Equal(t, http.StatusOK, status, string(body))
		var found []map[string]any
		require.NoError(t, json.Unmarshal(body, &found), string(body))
		return found
	}
	// of is what of each treatment a cold event keeps
	of := func(found []map[string]any) [][]any {
		kept := [][]any{}
		for _, treatment := range found {
			kept = append(kept, []any{treatment["_id"], treatment["eventType"], treatment["created_at"], treatment["carbs"], treatment["insulin"], treatment["notes"],
				treatment["enteredBy"], treatment["duration"], treatment["targetTop"], treatment["targetBottom"], treatment["reason"]})
		}
		return kept
	}
	all := url.Values{"count": {"100"}}
	before := treatments(all)
	require.Len(t, before, 5)

	// The days are packaged, then the rows of the old day are deleted, as the server will once it deletes treatments
	cfg.RetentionDeleteRows = true
	_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	groups, err := q.EventGroupsOfDay(ctx, sitedb.EventGroupsOfDayParams{From: old.UnixMilli(), To: old.Add(24 * time.Hour).UnixMilli()})
	require.NoError(t, err)
	require.Len(t, groups, 4)
	for _, group := range groups {
		_, err := q.HardDeleteEventGroup(ctx, sitedb.HardDeleteEventGroupParams{ID: group.ID})
		require.NoError(t, err)
	}
	require.Len(t, treatments(url.Values{"count": {"100"}, "find[eventType]": {"Meal Bolus"}, "find[carbs][$gte]": {"1"}}), 0, "a search by carbs is of rows only")

	// A search still has every treatment, the newest first, each as it was: its own ID, and what its events kept
	found := treatments(all)
	require.Equal(t, of(before), of(found))
	require.Equal(t, "Meal Bolus", found[4]["eventType"])
	require.Equal(t, []any{45.0, 4.5, "Pasta", "xDrip"}, []any{found[4]["carbs"], found[4]["insulin"], found[4]["notes"], found[4]["enteredBy"]})

	// By dates, by type, and a page of them
	ms := func(at time.Time) string { return fmt.Sprint(at.UnixMilli()) }
	inRange := treatments(url.Values{"count": {"100"}, "find[created_at][$gte]": {iso(old.Add(90 * time.Minute))}, "find[created_at][$lt]": {iso(old.Add(210 * time.Minute))}})
	require.Equal(t, [][]any{of(before)[2], of(before)[3]}, of(inRange))
	require.Equal(t, [][]any{of(before)[3]}, of(treatments(url.Values{"count": {"100"}, "find[eventType]": {"Temporary Target"}})))
	require.Equal(t, of(before)[1:3], of(treatments(url.Values{"count": {"2"}, "skip": {"1"}})))
	require.Empty(t, treatments(url.Values{"count": {"100"}, "find[date][$lt]": {ms(old)}}))

	// A server that does not delete rows reads only rows
	cfg.RetentionDeleteRows = false
	require.Len(t, treatments(all), 1)
}
