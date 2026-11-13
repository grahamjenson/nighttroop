package nscollections

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A merge lays the fields sent over the stored document as a JSON merge patch does
func TestMergeFields(t *testing.T) {
	treatment := map[string][]string{"duration": {"durationInMilliseconds", "durationType"},
		"targetTop": {"correctionRange"}, "targetBottom": {"correctionRange"}, "correctionRange": {"targetTop", "targetBottom"}}
	glucose := []string{"glucose", "targetTop", "targetBottom", "correctionRange"}
	tests := []struct {
		name         string
		stored, sent string
		aliases      map[string][]string
		expected     string
	}{
		{"a field not sent is kept", `{"notes":"lunch","carbs":20}`, `{"carbs":30}`, nil, `{"notes":"lunch","carbs":30}`},
		{"a field sent as null is taken out", `{"notes":"lunch","carbs":20}`, `{"notes":null}`, nil, `{"carbs":20}`},
		{"an object sent in part is merged into the stored one", `{"pump":{"reservoir":50,"model":"Dash","battery":{"percent":80}},"uploader":{"battery":90,"name":"phone"}}`,
			`{"pump":{"reservoir":40,"battery":{"voltage":1.4},"model":null}}`, nil,
			`{"pump":{"reservoir":40,"battery":{"percent":80,"voltage":1.4}},"uploader":{"battery":90,"name":"phone"}}`},
		{"a list sent takes the place of the stored one", `{"correctionRange":[90,100],"units":"mg/dL"}`, `{"correctionRange":[110]}`, nil,
			`{"correctionRange":[110],"units":"mg/dL"}`},
		{"a field sent takes the place of its aliases", `{"duration":60,"durationInMilliseconds":3600000,"durationType":"indefinite"}`,
			`{"duration":30}`, treatment, `{"duration":30}`},
		{"targets sent take the place of the correction range", `{"targetTop":100,"targetBottom":90,"correctionRange":[90,100],"units":"mg/dL"}`,
			`{"targetTop":120}`, treatment, `{"targetTop":120,"targetBottom":90,"units":"mg/dL"}`},
		{"glucose not sent is in the units sent", `{"targetTop":162,"targetBottom":162,"units":"mg/dL"}`,
			`{"targetTop":9,"units":"mmol"}`, treatment, `{"targetTop":9,"targetBottom":8.991,"units":"mmol"}`},
		{"glucose not sent is in the units guessed from the glucose sent", `{"glucose":5.5,"correctionRange":[5,6],"units":"mmol/L"}`,
			`{"glucose":120}`, treatment, `{"glucose":120,"correctionRange":[90.09,108.108],"units":"mg/dL"}`},
		{"glucose stored is as it was when none is sent", `{"glucose":120,"units":"mg/dL"}`, `{"notes":"check"}`, treatment,
			`{"glucose":120,"units":"mg/dL","notes":"check"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stored, sent map[string]any
			require.NoError(t, json.Unmarshal([]byte(test.stored), &stored))
			require.NoError(t, json.Unmarshal([]byte(test.sent), &sent))
			merged, err := json.Marshal(mergeFields(stored, sent, test.aliases, glucose))
			require.NoError(t, err)
			require.JSONEq(t, test.expected, string(merged))
		})
	}
}

// mergeSite is a new site, with glucose in units, and the context of its admin's request
func mergeSite(t *testing.T, units consts.Units) (*httputils.SiteContext, *sitedb.Queries) {
	_, _, site := testutils.NewControlDBWithSite(t)
	return &httputils.SiteContext{SiteID: site.SiteDB.Subdomain, SiteDB: site.SiteDB,
		SiteModel:     &controldb.Site{ID: site.SiteDB.Subdomain, Units: &units},
		IdentityModel: &controldb.SiteIdentity{ID: 1, Name: "admin", Role: consts.IdentityRole("admin")}}, site.SiteDB.Queries()
}

// decode is a JSON object as a map, as the API reads a body
func decode(t *testing.T, body string) map[string]any {
	var fields map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &fields))
	return fields
}

// A patch sending nothing leaves every document as it was: what a document gives back is read back the same, in the
// site's units whichever they are
func TestPatchSendingNothingChangesNothing(t *testing.T) {
	date := time.Now().Add(-time.Hour).UnixMilli()
	documents := map[string][]string{
		"treatments": {
			`{"eventType":"Correction Bolus","insulin":2.5,"type":"SMB","pumpId":"123","pumpType":"OMNIPOD_DASH","pumpSerial":"abc","app":"AAPS","utcOffset":780,"isValid":true,"icfg":{"insulinLabel":"Lyumjev","insulinEndTime":25200000,"insulinPeakTime":2700000}}`,
			`{"eventType":"Meal Bolus","insulin":3,"carbs":40,"fat":10,"protein":5,"glucose":7.2,"glucoseType":"Finger","units":"mmol","notes":"lunch","enteredBy":"careportal","preBolus":15}`,
			`{"eventType":"Temp Basal","rate":0.5,"absolute":0.5,"duration":30,"durationInMilliseconds":1800000,"type":"NORMAL","app":"AAPS"}`,
			`{"eventType":"Temp Basal","percent":-50,"duration":30}`,
			`{"eventType":"Temp Basal","absolute":1.2,"rate":1.2,"duration":29.833333333333332,"temp":"absolute","automatic":true,"insulinType":"Fiasp","syncIdentifier":"loop-temp-1"}`,
			`{"eventType":"Temporary Target","targetTop":6,"targetBottom":5.5,"units":"mmol","duration":60,"reason":"Activity","app":"T1DCarer"}`,
			`{"eventType":"Temporary Override","correctionRange":[100,120],"insulinNeedsScaleFactor":1.2,"duration":0,"durationType":"indefinite","reason":"Running","enteredBy":"Loop"}`,
			`{"eventType":"Profile Switch","profile":"Day","originalProfileName":"Day","timeshift":3600000,"percentage":110,"duration":0,"profileJson":"{\"dia\":5,\"carbratio\":[{\"time\":\"00:00\",\"value\":10,\"timeAsSeconds\":0}],\"sens\":[{\"time\":\"00:00\",\"value\":2.5,\"timeAsSeconds\":0}],\"basal\":[{\"time\":\"00:00\",\"value\":0.8,\"timeAsSeconds\":0}],\"target_low\":[{\"time\":\"00:00\",\"value\":5,\"timeAsSeconds\":0}],\"target_high\":[{\"time\":\"00:00\",\"value\":6,\"timeAsSeconds\":0}],\"units\":\"mmol\",\"timezone\":\"Pacific/Auckland\"}"}`,
			`{"eventType":"Profile Switch","profile":"Night","duration":60}`,
			`{"eventType":"Note","notes":"site change soon"}`,
			`{"eventType":"Note","notes":"ate","carbs":12}`,
			`{"eventType":"Announcement","notes":"meter low"}`,
			`{"eventType":"Question","notes":"bolus?"}`,
			`{"eventType":"Exercise","name":"Run","activityType":"running","duration":45,"steps":5000,"calories":300,"heartRate":150}`,
			`{"eventType":"Sensor Change","sensorCode":"1234","transmitterId":"ABC"}`,
			`{"eventType":"BG Check","glucose":120,"glucoseType":"Finger","units":"mg/dl"}`,
			`{"eventType":"Carb Correction","carbs":30,"absorptionTime":180,"foodType":"🍕","userEnteredAt":"2024-01-01T00:00:00Z","syncIdentifier":"loop-carb-1"}`,
			`{"eventType":"<none>","insulin":1.5,"uuid":"xdrip-uuid-1","enteredBy":"xdrip"}`,
			`{"eventType":"OpenAPS Offline","mode":"DISCONNECTED_PUMP","autoForced":false,"reasons":"user","originalDuration":3600000,"duration":60}`,
			`{"eventType":"Bolus Wizard","insulin":4,"carbs":45,"glucose":110,"units":"mg/dL","bolusCalculatorResult":"{\"carbs\":45}"}`,
			`{"eventType":"Combo Bolus","insulin":4,"splitNow":60,"splitExt":40,"duration":120,"carbs":50}`,
		},
		"entries": {
			`{"type":"sgv","sgv":120,"direction":"Flat","delta":1.5,"noise":1,"filtered":120000,"unfiltered":121000,"rssi":100,"device":"xDrip-DexcomG6"}`,
			`{"type":"mbg","mbg":7.5,"units":"mmol","device":"meter"}`,
			`{"type":"mbg","mbg":110,"isCalibration":true,"sensorMgdlAtPairing":105,"device":"AAPS"}`,
		},
		"devicestatus": {
			`{"device":"loop://iPhone","uploader":{"battery":80,"name":"iPhone","timestamp":"2024-01-01T00:00:00Z"},"pump":{"manufacturer":"Insulet","model":"Dash","pumpID":"1f","suspended":false,"bolusing":false,"reservoir":50,"clock":"2024-01-01T00:00:00Z"},"loop":{"iob":{"iob":2.5},"cob":{"cob":20},"predicted":{"startDate":"2024-01-01T00:00:00Z","values":[100,110]}}}`,
			`{"device":"openaps://AAPS","uploaderBattery":60,"isCharging":true,"pump":{"battery":{"percent":70},"reservoir":120,"status":{"status":"normal"},"extended":{"ActiveProfile":"Day"}},"openaps":{"iob":{"iob":1,"basaliob":0.5},"suggested":{"COB":10,"predBGs":{"IOB":[100,101]}}},"configuration":{"aps":"SMB","insulin":2}}`,
		},
		"food":     {`{"type":"food","name":"Toast","category":"Breakfast","subcategory":"Bread","carbs":15,"fat":2,"protein":3,"energy":300,"portion":1,"unit":"g","gi":2}`},
		"activity": {`{"activityType":"running","name":"Run","duration":45,"steps":5000,"calories":300,"heartRate":150,"notes":"easy"}`},
	}
	collections := map[string]CollectionStore{"treatments": NewTreatmentCollection(), "entries": NewEntriesCollection(),
		"devicestatus": NewDeviceStatusCollection(), "food": NewFoodCollection(), "activity": NewActivityCollection()}
	for _, units := range []consts.Units{consts.UnitsMGDL, consts.UnitsMMOLL} {
		siteContext, _ := mergeSite(t, units)
		for name, docs := range documents {
			for i, doc := range docs {
				t.Run(fmt.Sprintf("%s %s %d", units, name, i), func(t *testing.T) {
					data := decode(t, doc)
					data["date"] = float64(date + int64(i)*1000)
					created, _, httpErr := collections[name].Create(t.Context(), siteContext, data)
					require.Nil(t, httpErr)
					patched, httpErr := collections[name].Patch(t.Context(), siteContext, created.Identifier(), map[string]any{})
					require.Nil(t, httpErr)
					for _, v3 := range []bool{false, true} {
						before, after := created.Fields(v3), patched.Fields(v3)
						delete(before, "srvModified")
						delete(after, "srvModified")
						require.Equal(t, before, after, "v3 %v", v3)
					}
				})
			}
		}
	}
}

// A patch of a treatment keeps the fields it does not send, takes out those it sends as null, and adds or deletes the
// treatment's events as it gains or loses their fields
func TestPatchTreatment(t *testing.T) {
	siteContext, q := mergeSite(t, consts.UnitsMGDL)
	treatments := NewTreatmentCollection()
	created, _, httpErr := treatments.Create(t.Context(), siteContext, decode(t, `{"eventType":"Note","notes":"lunch","reason":"food"}`))
	require.Nil(t, httpErr)
	id := created.Identifier()
	stored := func() *sitedb.EventGroupsBundle {
		bundle, err := TreatmentCollection{}.GetModel(q, t.Context(), id, false)
		require.NoError(t, err)
		return bundle
	}

	// Carbs sent add a food intake, and the notes not sent are kept
	patched, httpErr := treatments.Patch(t.Context(), siteContext, id, decode(t, `{"carbs":20}`))
	require.Nil(t, httpErr)
	require.Equal(t, "lunch", patched.Fields(true)["notes"])
	require.Equal(t, 20.0, patched.Fields(true)["carbs"])
	foodIntakeID := stored().EventGroup.FoodIntakeID
	require.NotNil(t, foodIntakeID)

	// New carbs keep the food intake, written over
	_, httpErr = treatments.Patch(t.Context(), siteContext, id, decode(t, `{"carbs":25}`))
	require.Nil(t, httpErr)
	require.Equal(t, foodIntakeID, stored().EventGroup.FoodIntakeID)
	require.Equal(t, 25.0, *stored().FoodIntake.Carbs)

	// Carbs and notes sent as null take them out: the food intake is deleted
	patched, httpErr = treatments.Patch(t.Context(), siteContext, id, decode(t, `{"carbs":null,"notes":null}`))
	require.Nil(t, httpErr)
	require.NotContains(t, patched.Fields(true), "carbs")
	require.NotContains(t, patched.Fields(true), "notes")
	require.Equal(t, "food", patched.Fields(true)["reason"])
	require.Nil(t, stored().EventGroup.FoodIntakeID)
	_, err := q.GetFoodIntake(t.Context(), sitedb.GetFoodIntakeParams{ID: *foodIntakeID, IncludeDeleted: true})
	require.Error(t, err)

	// When and by whom it was created are kept
	require.Equal(t, created.Fields(true)["srvCreated"], patched.Fields(true)["srvCreated"])
	require.Equal(t, "admin", stored().EventGroup.CreatedByIdentityName)
}

// A patch of a device status with a pump in part keeps the pump's other fields; the uploader and loop status it does
// not send are kept, the loop status as it is kept once a newer status cleared it
func TestPatchDeviceStatus(t *testing.T) {
	siteContext, q := mergeSite(t, consts.UnitsMGDL)
	statuses := NewDeviceStatusCollection()
	created, _, httpErr := statuses.Create(t.Context(), siteContext, decode(t, `{"device":"loop://iPhone","uploader":{"battery":80,"name":"iPhone"},`+
		`"pump":{"manufacturer":"Insulet","reservoir":50,"suspended":true},"loop":{"iob":{"iob":2.5},"cob":{"cob":20},"predicted":{"values":[100,110]}}}`))
	require.Nil(t, httpErr)
	id := created.Identifier()
	require.NoError(t, q.ClearDeviceReportStatuses(t.Context(), sitedb.ClearDeviceReportStatusesParams{ID: id}))

	patched, httpErr := statuses.Patch(t.Context(), siteContext, id, decode(t, `{"pump":{"reservoir":40,"suspended":null}}`))
	require.Nil(t, httpErr)
	require.Equal(t, map[string]any{"manufacturer": "Insulet", "reservoir": 40.0}, patched.Fields(true)["pump"])
	require.Equal(t, map[string]any{"battery": 80.0, "name": "iPhone"}, patched.Fields(true)["uploader"])
	require.Equal(t, map[string]any{"iob": map[string]any{"iob": 2.5}, "cob": map[string]any{"cob": 20.0}}, patched.Fields(true)["loop"])
	report, err := q.GetDeviceReport(t.Context(), sitedb.GetDeviceReportParams{ID: id})
	require.NoError(t, err)
	require.Nil(t, report.LoopStatusJson)
	require.Nil(t, report.PumpSuspended)
	require.Equal(t, "Insulet", *report.PumpManufacturer)
}

// A reading sent again with the values it has is not changed, and one patched keeps its notes from the site, which an
// entry does not have
func TestEntryMergeKeepsWhatOnlyTheDatabaseHas(t *testing.T) {
	siteContext, q := mergeSite(t, consts.UnitsMGDL)
	entries := NewEntriesCollection()
	date := float64(time.Now().Add(-time.Hour).UnixMilli())
	sent := map[string]any{"type": "sgv", "sgv": 120.0, "direction": "Flat", "date": date, "device": "xDrip"}
	created, _, httpErr := entries.Create(t.Context(), siteContext, sent)
	require.Nil(t, httpErr)
	again, _, httpErr := entries.Create(t.Context(), siteContext, map[string]any{"type": "sgv", "sgv": 120.0, "direction": "Flat", "date": date})
	require.Nil(t, httpErr)
	require.Equal(t, created.Modified(), again.Modified())

	reading, err := q.GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: created.Identifier()})
	require.NoError(t, err)
	params := sitedb.RewriteGlucoseReadingParams{ID: reading.ID, Mgdl: reading.Mgdl, ReadingType: reading.ReadingType,
		Direction: reading.Direction, Notes: new("site note"), OccurredAt: reading.OccurredAt, UtcOffsetMinutes: reading.UtcOffsetMinutes,
		CreatedByDeviceName: reading.CreatedByDeviceName, CreatedByIdentityName: "admin"}
	_, err = q.RewriteGlucoseReading(t.Context(), params)
	require.NoError(t, err)

	patched, httpErr := entries.Patch(t.Context(), siteContext, created.Identifier(), map[string]any{"sgv": 130.0})
	require.Nil(t, httpErr)
	require.Equal(t, 130.0, patched.Fields(true)["sgv"])
	reading, err = q.GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: created.Identifier()})
	require.NoError(t, err)
	require.Equal(t, "site note", *reading.Notes)
	require.Equal(t, "Flat", *reading.Direction)
}
