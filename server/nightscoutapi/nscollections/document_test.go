package nscollections

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// mmolSite is a site whose settings write glucose in mmol/L
func mmolSite() *httputils.SiteContext {
	return &httputils.SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh", Units: new(consts.UnitsMMOLL)}}
}

// dayOfEntries are a day of CGM readings, five minutes apart, as documents
func dayOfEntries() []nsutils.Document {
	siteContext := mmolSite()
	entries := NewEntriesCollection().(Collection[nightscoutmodels.Entry, sitedb.GlucoseReading])
	documents := make([]nsutils.Document, 0, 288)
	for i := range int64(288) {
		delta, noise, direction, device := float64(i%7)-3, 1.0, "Flat", "xDrip-DexcomG6"
		occurredAt := 1_700_000_000_000 + i*300_000
		documents = append(documents, entries.document(&sitedb.GlucoseReading{ID: "0123456789abcdef01234567", Mgdl: 100 + float64(i%50),
			ReadingType: "sgv", Delta: &delta, Direction: &direction, Noise: &noise, CreatedByDeviceName: &device,
			OccurredAt: occurredAt, CreatedAt: occurredAt, UpdatedAt: occurredAt, CreatedByIdentityName: "xdrip"}, siteContext))
	}
	return documents
}

// A 288 entry /api/v1/entries search response, as its documents are written
func BenchmarkWriteRecordsV1Entries(b *testing.B) {
	documents := dayOfEntries()
	b.ReportAllocs()
	for b.Loop() {
		nsutils.WriteRecordsV1(httptest.NewRecorder(), documents, nil)
	}
}

// mapFields are a document's fields as they were made before they were written from its model's JSON: by reading the
// JSON back as a map, and setting the server's fields over it
func mapFields[NS any](s stored[NS], v3 bool, only []string) map[string]any {
	out, _ := utils.ToMap(&s.model)
	if units, ok := out["units"].(string); ok && s.unitsLabels[units] != "" {
		out["units"] = s.unitsLabels[units]
	}
	if out == nil {
		out = map[string]any{}
	}
	out["created_at"] = time.UnixMilli(s.occurredAt).UTC().Format(time.RFC3339Nano)
	out["utcOffset"] = s.utcOffset
	if v3 {
		out["identifier"], out["date"], out["srvCreated"], out["srvModified"] = s.id, s.occurredAt, s.srvCreated, s.srvModified
		out["subject"], out["isValid"], out["isReadOnly"] = s.subject, s.deletedAt == nil, false
		delete(out, "_id")
	} else {
		out["_id"] = s.id
		delete(out, "identifier")
	}
	if len(only) == 0 {
		return out
	}
	projected := map[string]any{}
	for _, field := range only {
		if value, ok := out[field]; ok && field != "_id" {
			projected[field] = value
		}
	}
	return projected
}

// written is a value as the API writes it
func written(t *testing.T, value any) string {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	require.NoError(t, encoder.Encode(value))
	return out.String()
}

// model is a Nightscout model read from JSON as a client sends it
func model[NS any](t *testing.T, sent string) NS {
	var out NS
	require.NoError(t, json.Unmarshal([]byte(sent), &out))
	return out
}

// A document is written from its model's JSON with the fields, and values, it had when it was read back as a map
// first, in either API version and with or without fields asked for: the server's over the model's, units labelled
// as the settings write them, and HTML, nested objects and numbers as they were
func TestDocumentJSONIsItsFields(t *testing.T) {
	m := meta{id: "0123456789abcdef01234567", srvCreated: 1_700_000_001_000, srvModified: 1_700_000_002_000,
		occurredAt: 1_700_000_000_000, utcOffset: 780, subject: "Tom & <Jerry>"}
	deleted := m
	deleted.deletedAt = new(int64(1_700_000_003_000))
	documents := map[string]nsutils.Document{
		"entry": dayOfEntries()[0],
		"labelled units": stored[nightscoutmodels.Treatment]{meta: m, unitsLabels: map[string]string{"mg/dL": "mg/dl"},
			model: model[nightscoutmodels.Treatment](t, `{"eventType":"BG Check","glucose":5.5,"units":"mg/dl","glucoseType":"Finger"}`)},
		"treatment": stored[nightscoutmodels.Treatment]{meta: deleted, model: model[nightscoutmodels.Treatment](t,
			`{"_id":"x","identifier":"y","eventType":"Meal Bolus","insulin":1.25,"carbs":30,"notes":"<b>pie</b> & \u2028\t\b chips","created_at":"2020-01-01T00:00:00Z","utcOffset":60}`)},
		"device status": stored[nightscoutmodels.DeviceStatus]{meta: m, model: model[nightscoutmodels.DeviceStatus](t,
			`{"device":"openaps://phone","openaps":{"iob":{"iob":1.50,"time":1700000000000,"big":12345678901234567},"z":{"b":[1,2.0,3e2],"a":"<"}},"pump":{"battery":{"percent":80}}}`)},
		"setting": stored[setting]{meta: m, model: setting{Identifier: "units", Value: map[string]any{"z": 1, "a": []any{1.5, "x"}}, Type: "json"}},
		"empty":   stored[setting]{},
	}
	for name, document := range documents {
		for _, v3 := range []bool{false, true} {
			for _, only := range [][]string{nil, {"_id", "identifier", "units", "notes", "date", "openaps", "value", "missing"}} {
				var want map[string]any
				switch s := document.(type) {
				case stored[nightscoutmodels.Entry]:
					want = mapFields(s, v3, only)
				case stored[nightscoutmodels.Treatment]:
					want = mapFields(s, v3, only)
				case stored[nightscoutmodels.DeviceStatus]:
					want = mapFields(s, v3, only)
				case stored[setting]:
					want = mapFields(s, v3, only)
				}
				require.Equal(t, written(t, want), written(t, document.JSON(v3, only)), "%s v3=%v fields=%v", name, v3, only)
				if only == nil {
					require.Equal(t, written(t, want), written(t, document.Fields(v3)), "%s v3=%v", name, v3)
				}
			}
		}
	}
}
