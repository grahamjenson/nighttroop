package server

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// However many records are asked for, no endpoint returns more than a search's worth, so no request can read a whole
// site at once, nor delete all of a type of treatment or entry at once. A Nightscout API search of readings returns up
// to its own most, a day of 1 minute readings twice over, as xDrip+'s follower asks for that many since its newest.
func TestNoEndpointReturnsMoreThanASearch(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	handler := GetMainHandler(controlDB, config)
	subdomain := site.SiteDB.Subdomain
	jwt, _, err := config.Security.GenerateUserJWTToken(subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	require.NoError(t, err)

	send := func(method, url, body string) any {
		r := httptest.NewRequest(method, url, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+jwt)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Less(t, w.Code, 300, "%s %s: %s", method, url, w.Body.String())
		var out any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out), url)
		return out
	}

	// More glucose readings and treatments than a search returns, readings more than a search of them returns
	const records, readings = apiquery.MaxQueryLimit + 50, apiquery.MaxEntriesQueryLimit + 50
	now := time.Now()
	for i := range readings {
		_, err := site.SiteDB.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: fmt.Sprintf("reading-%d", i),
			Mgdl: 120, ReadingType: "sgv", OccurredAt: now.Add(-time.Duration(i) * time.Minute).UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	notes := []string{}
	for i := range records {
		notes = append(notes, fmt.Sprintf(`{"eventType": "Note", "notes": "note %d", "created_at": %q}`, i, now.Add(-time.Duration(i)*time.Minute).Format(time.RFC3339)))
	}
	ns := "http://" + subdomain + ".ns.localhost"
	// An upload is at most 100 records
	for start := 0; start < records; start += 100 {
		send("POST", ns+"/api/v1/treatments?token="+site.PwdNSToken, "["+strings.Join(notes[start:min(start+100, records)], ",")+"]")
	}

	const absurd = "1000000"
	for url, most := range map[string]int{
		"http://" + subdomain + ".localhost/data/glucose?limit=" + absurd:      apiquery.MaxQueryLimit,
		"http://" + subdomain + ".localhost/data/event-groups?limit=" + absurd: apiquery.MaxQueryLimit,
		ns + "/api/v1/entries.json?count=" + absurd:                            apiquery.MaxEntriesQueryLimit,
		ns + "/api/v1/entries/sgv.json?count=" + absurd:                        apiquery.MaxEntriesQueryLimit,
		ns + "/api/v1/treatments.json?count=" + absurd:                         apiquery.MaxQueryLimit,
		ns + "/api/v1/slice/entries/dateString/sgv/2?count=" + absurd:          apiquery.MaxEntriesQueryLimit,
		ns + "/api/v1/times/2?count=" + absurd:                                 apiquery.MaxQueryLimit,
		ns + "/api/v3/entries?limit=" + absurd:                                 apiquery.MaxEntriesQueryLimit,
		ns + "/api/v3/treatments?limit=" + absurd:                              apiquery.MaxQueryLimit,
		ns + "/api/v3/entries/history/0?limit=" + absurd:                       apiquery.MaxEntriesQueryLimit,
	} {
		result := send("GET", url+"&token="+site.PwdNSToken, "")
		// The site API and Nightscout's v3 put the records in result
		if wrapped, ok := result.(map[string]any); ok {
			result = wrapped["result"]
		}
		// and the site API's search a page of them in items
		if page, ok := result.(map[string]any); ok {
			result = page["items"]
		}
		require.Len(t, result, most, url)
	}

	// xDrip+'s follower, offline for 1300 readings, is given all of them since its newest, the newest first, where the
	// newest 1000 left the rest unread for good
	gap := send("GET", ns+"/api/v1/entries.json?count=2880&find[date][$gt]="+fmt.Sprint(now.Add(-1300*time.Minute).UnixMilli())+"&token="+site.PwdNSToken, "").([]any)
	require.Len(t, gap, 1300)
	require.Equal(t, "reading-0", gap[0].(map[string]any)["_id"])

	// A skip is capped too, so an absurd one stops at MaxQuerySkip, past these rows, rather than walking every row
	page := send("GET", "http://"+subdomain+".localhost/data/glucose?skip="+absurd, "").(map[string]any)["result"].(map[string]any)
	require.Equal(t, float64(apiquery.MaxQuerySkip), page["skip"])
	require.Empty(t, page["items"])

	// The clean readings are worked out from the raw ones, and paged the same
	clean := send("GET", "http://"+subdomain+".localhost/data/clean_glucose?hours="+absurd+"&limit="+absurd, "")
	require.LessOrEqual(t, len(clean.(map[string]any)["result"].(map[string]any)["items"].([]any)), apiquery.MaxQueryLimit)
	count := send("GET", ns+"/api/v1/count/entries/where?count="+absurd+"&token="+site.PwdNSToken, "")
	require.Equal(t, float64(apiquery.MaxEntriesQueryLimit), count.(map[string]any)["count"])

	// Deleting a type of treatment deletes at most a search's worth each time, saying how many
	deleted := send("DELETE", ns+"/api/v1/treatments/Note?token="+site.PwdNSToken, "")
	require.Equal(t, float64(apiquery.MaxQueryLimit), deleted.(map[string]any)["deletedCount"])
	deleted = send("DELETE", ns+"/api/v1/treatments/Note?token="+site.PwdNSToken, "")
	require.Equal(t, float64(records-apiquery.MaxQueryLimit), deleted.(map[string]any)["deletedCount"])

	// So does deleting a type of entry, soft deleting or for good, which deletes those soft deleted too
	for _, permanent := range []string{"", "&permanent=true"} {
		for _, want := range []int{apiquery.MaxQueryLimit, apiquery.MaxQueryLimit, readings - 2*apiquery.MaxQueryLimit} {
			deleted = send("DELETE", ns+"/api/v1/entries/sgv?token="+site.PwdNSToken+permanent, "")
			require.Equal(t, float64(want), deleted.(map[string]any)["deletedCount"], permanent)
		}
	}
}
