package server

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// fileSite is the server with a new site whose DBs are files, as on a server and not in memory as in other tests, so a
// commit syncs to disk and a read does not wait for a write's transaction. send sends a request as the site's phone.
func fileSite(t testing.TB) (*config.ServerConfig, *db.SiteCreateResult, func(method, path, body string) *httptest.ResponseRecorder) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	require.NoError(t, controlDB.SeedDevAdmin(t.Context()))
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	handler := GetMainHandler(controlDB, &cfg)
	jwt, _, err := cfg.Security.GenerateUserJWTToken(site.SiteDB.Subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	require.NoError(t, err)
	return &cfg, site, func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://"+site.SiteDB.Subdomain+".ns.localhost"+path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+jwt)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
}

// reading is a CGM reading at a time in milliseconds, as an uploader sends it
func reading(at int64) string {
	return fmt.Sprintf(`{"type":"sgv","sgv":%d,"date":%d,"dateString":%q,"direction":"Flat","device":"xDrip-DexcomG6"}`,
		100+at/1000%50, at, time.UnixMilli(at).UTC().Format(time.RFC3339))
}

// readings are a batch of count readings a second apart from at, then the documents in more
func readings(at int64, count int, more ...string) string {
	docs := make([]string, count)
	for i := range docs {
		docs[i] = reading(at + int64(i)*1000)
	}
	return "[" + strings.Join(append(docs, more...), ",") + "]"
}

// ids are the _id of each document of a v1 response, in its order
func ids(t *testing.T, w *httptest.ResponseRecorder) []string {
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var docs []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &docs))
	out := make([]string, len(docs))
	for i, doc := range docs {
		out[i] = doc["_id"].(string)
	}
	return out
}

// A v1 batch is stored a hundred documents at a time, each hundred in one transaction of the site's DB, so the DB syncs
// to disk once for them and not once for each. Each document is published once it is committed, in the order sent, and
// the response is as it was. A duplicate undoes its hundred's transaction: the documents before it are stored in one
// again, and it and those after it as before, each in a transaction of its own, with nothing published or left out
// twice. A document that replaces another of its batch is stored in the same transaction.
func TestV1BatchIsStoredInOneTransaction(t *testing.T) {
	cfg, site, send := fileSite(t)
	q := site.SiteDB.Queries()
	sqltiming.SetSamplePercent(100)
	t.Cleanup(func() { sqltiming.SetSamplePercent(0) })
	// begun is how many transactions site DBs have begun
	begun := func() int64 {
		for _, query := range sqltiming.Slowest(math.MaxInt) {
			if query.DB == "sitedb" && query.Query == "BEGIN" {
				return query.Count
			}
		}
		return 0
	}
	// published is each change published since it was last read, as its event's name and model's ID. A model published
	// is stored, as another connection reads it.
	var changes []string
	defer cfg.Events.Subscribe(site.SiteDB.Subdomain, func(event events.Event) {
		switch model := event.Model.(type) {
		case *sitedb.GlucoseReading:
			_, err := q.GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: model.ID})
			assert.NoError(t, err, "a reading is published only once it is committed")
			changes = append(changes, string(event.Name)+" "+model.ID)
		case *sitedb.EventGroupsBundle:
			changes = append(changes, string(event.Name)+" "+model.EventGroup.ID)
		}
	})()
	published := func() []string {
		taken := changes
		changes = nil
		return taken
	}
	named := func(name string, ids ...string) []string {
		out := make([]string, len(ids))
		for i, id := range ids {
			out[i] = name + " " + id
		}
		return out
	}
	at := time.Now().Add(-48 * time.Hour).UnixMilli()

	// 250 new readings are three transactions: two hundreds and a fifty
	before := begun()
	stored := ids(t, send("POST", "/api/v1/entries", readings(at, 250)))
	require.Len(t, stored, 250)
	require.EqualValues(t, 3, begun()-before)
	require.Equal(t, named("glucose_reading.created", stored...), published())
	newest, err := q.SearchGlucoseReadings(t.Context(), sitedb.SearchGlucoseReadingsParams{Limit: 1000})
	require.NoError(t, err)
	require.Len(t, newest, 250)
	for i, reading := range newest {
		require.Equal(t, stored[249-i], reading.ID, "the readings are stored in the order sent")
	}

	// The same readings sent again are each a duplicate, left as it is: each hundred tries one transaction, which its
	// first duplicate undoes, and is then checked as before, each reading in a transaction of its own
	before = begun()
	require.Equal(t, stored, ids(t, send("POST", "/api/v1/entries", readings(at, 250))))
	require.EqualValues(t, 3+250, begun()-before)
	require.Empty(t, published())

	// A duplicate undoes its batch's transaction: the documents before it are stored in one again, then it and those
	// after it each in one of its own. What the undone try left out is forgotten: six readings with a problem are left
	// out once, where twelve would refuse the batch. A calibration is not kept, and the duplicate is given back as it
	// is stored.
	bad, calibration := `{"type":"unknown","date":1}`, `{"type":"cal","date":2,"slope":1}`
	before = begun()
	got := ids(t, send("POST", "/api/v1/entries", readings(0, 0, bad, bad, bad, bad, bad, bad, calibration,
		reading(at+300_000), reading(at), reading(at+301_000))))
	require.Len(t, got, 3)
	require.EqualValues(t, 4, begun()-before)
	require.Equal(t, stored[0], got[1])
	require.Equal(t, named("glucose_reading.created", got[0], got[2]), published())

	// A treatment sent twice in a batch with its _id is created, then replaced by the second, which reads the first in
	// the batch's transaction
	note := func(id, notes string) string {
		return fmt.Sprintf(`{"_id":%q,"eventType":"Note","notes":%q,"created_at":%q}`, id, notes, time.UnixMilli(at).UTC().Format(time.RFC3339))
	}
	first, second := strings.Repeat("a", 24), strings.Repeat("b", 24)
	before = begun()
	require.Equal(t, []string{first, first, second},
		ids(t, send("POST", "/api/v1/treatments", "["+note(first, "one")+","+note(first, "two")+","+note(second, "three")+"]")))
	require.EqualValues(t, 1, begun()-before)
	require.Equal(t, []string{"event_groups_bundle.created " + first, "event_groups_bundle.updated " + first, "event_groups_bundle.created " + second}, published())
	var treatments []map[string]any
	require.NoError(t, json.Unmarshal(send("GET", "/api/v1/treatments.json", "").Body.Bytes(), &treatments))
	require.Len(t, treatments, 2)
	require.ElementsMatch(t, []any{"two", "three"}, []any{treatments[0]["notes"], treatments[1]["notes"]})
}

// A v1 upload of 100 readings, as xDrip+ or Juggluco send a backlog, to a site whose DB is a file, so each commit syncs
// to disk: of new readings, stored in one transaction; of the same readings again, each a duplicate; and of new
// readings with a duplicate last, which undoes their transaction, so each is stored in one of its own.
// Run with go test -run XXX -bench V1EntriesBatch ./server
func BenchmarkV1EntriesBatch(b *testing.B) {
	// The limits of a site's writes, which the server's one config reads when it is first made
	for _, name := range []string{"GLOBAL_IP", "GLOBAL", "SITE_POST"} {
		os.Setenv(name+"_PER_MINUTE_LIMIT", "100000000")
		os.Setenv(name+"_PER_MINUTE_BURST", "100000000")
	}
	_, _, send := fileSite(b)
	at := time.Now().Add(-48 * time.Hour).UnixMilli()
	post := func(body string) {
		if w := send("POST", "/api/v1/entries", body); w.Code != http.StatusOK {
			b.Fatal(w.Code, w.Body.String())
		}
	}
	b.Run("new", func(b *testing.B) {
		for b.Loop() {
			at += 100_000
			post(readings(at, 100))
		}
	})
	b.Run("sent again", func(b *testing.B) {
		post(readings(at, 100))
		for b.Loop() {
			post(readings(at, 100))
		}
	})
	b.Run("new with a duplicate", func(b *testing.B) {
		post(readings(at, 1))
		for b.Loop() {
			at += 100_000
			post(readings(at, 99, reading(at-100_000)))
		}
	})
}
