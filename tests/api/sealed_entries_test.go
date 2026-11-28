package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
func TestEntriesOfSealedDaysAreMadeFromColdDays(t *testing.T) {
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

	// A server that does not delete rows reads only rows, as all of them are there
	cfg.RetentionDeleteRows = false
	require.Equal(t, []string{"recent2", "recent1", "recent0", "meter", "noted"}, ids(entries(all)))
}
