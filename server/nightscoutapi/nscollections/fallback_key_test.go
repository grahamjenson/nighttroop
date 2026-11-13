package nscollections

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
)

// keyCase is a model sent, as its JSON, and a model stored, and whether the sent one repeats the stored one by the
// collection's fallback key
type keyCase[DM any] struct {
	name      string
	sent      string
	stored    *DM
	duplicate bool
}

// checkKey checks the key's fields are those named, and whether each model sent repeats the one stored
func checkKey[NS, DM any](t *testing.T, key FallbackKey[NS, DM], fields []string, cases []keyCase[DM]) {
	t.Helper()
	names := make([]string, len(key.fields))
	for i, field := range key.fields {
		names[i] = field.name
	}
	require.Equal(t, fields, names)
	for _, test := range cases {
		var sent NS
		require.NoError(t, json.Unmarshal([]byte(test.sent), &sent), test.name)
		require.Equal(t, test.duplicate, key.matches(sent, test.stored), test.name)
	}
}

const keyDate = int64(1762000000123)

func TestEntriesFallbackKey(t *testing.T) {
	key := EntriesCollection{}.FallbackKey()
	require.True(t, key.whateverItsIdentifier, "a reading is found by its date and type whatever its _id")
	sgv := &sitedb.GlucoseReading{OccurredAt: keyDate, ReadingType: "sgv", CreatedByDeviceName: new("xDrip-DexcomG6")}
	mbg := &sitedb.GlucoseReading{OccurredAt: keyDate, ReadingType: "mbg", CreatedByDeviceName: new("careportal")}
	checkKey(t, key, []string{"date", "type"}, []keyCase[sitedb.GlucoseReading]{
		{"same date and type", `{"date":1762000000123,"sgv":120,"device":"xDrip-DexcomG6"}`, sgv, true},
		{"another device, as Trio uploads what it downloads", `{"date":1762000000123,"sgv":120,"device":"Trio"}`, sgv, true},
		{"another value, which a refresh takes", `{"date":1762000000123,"sgv":130}`, sgv, true},
		{"a finger prick sent back by xDrip+", `{"date":1762000000123,"mbg":110,"device":"careportal via Nightscout"}`, mbg, true},
		{"the type named", `{"date":1762000000123,"type":"sgv","sgv":120}`, sgv, true},
		{"another type", `{"date":1762000000123,"mbg":120}`, sgv, false},
		{"a millisecond later", `{"date":1762000000124,"sgv":120}`, sgv, false},
		{"no date", `{"sgv":120}`, sgv, false},
	})
}

func TestTreatmentsFallbackKey(t *testing.T) {
	key := TreatmentCollection{}.FallbackKey()
	require.False(t, key.whateverItsIdentifier, "a treatment sent with an identifier is found only by it")
	bolus := func(change func(*sitedb.EventGroupsBundle)) *sitedb.EventGroupsBundle {
		stored := &sitedb.EventGroupsBundle{
			EventGroup:  sitedb.EventGroup{OccurredAt: keyDate, GroupType: "Meal Bolus", CreatedByDeviceName: new("AAPS")},
			InsulinDose: &sitedb.InsulinDose{Dose: 2},
		}
		if change != nil {
			change(stored)
		}
		return stored
	}
	meal := bolus(func(b *sitedb.EventGroupsBundle) { b.FoodIntake = &sitedb.FoodIntake{Carbs: new(30.0)} })
	checkKey(t, key, []string{"date", "device", "eventType", "insulin", "carbs"}, []keyCase[sitedb.EventGroupsBundle]{
		{"same date, device, eventType and insulin", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2,"device":"AAPS"}`, bolus(nil), true},
		{"no device, which finds any", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2}`, bolus(nil), true},
		{"0 carbs is none", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2,"carbs":0}`, bolus(nil), true},
		{"same insulin and carbs", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2,"carbs":30}`, meal, true},
		{"other notes and enteredBy", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2,"notes":"again","enteredBy":"me"}`, bolus(nil), true},
		{"other insulin", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2.5}`, bolus(nil), false},
		{"the Bolus Wizard's carbs beside its bolus", `{"date":1762000000123,"eventType":"Meal Bolus","carbs":30,"device":"AAPS"}`, bolus(nil), false},
		{"the Bolus Wizard's bolus beside its carbs", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2}`,
			bolus(func(b *sitedb.EventGroupsBundle) {
				b.InsulinDose, b.FoodIntake = nil, &sitedb.FoodIntake{Carbs: new(30.0)}
			}), false},
		{"another device", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2,"device":"xDrip"}`, bolus(nil), false},
		{"another eventType", `{"date":1762000000123,"eventType":"Correction Bolus","insulin":2}`, bolus(nil), false},
		{"a millisecond later", `{"date":1762000000124,"eventType":"Meal Bolus","insulin":2}`, bolus(nil), false},
		{"a careportal note beside xDrip+'s bolus, stored with its uuid", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2}`,
			bolus(func(b *sitedb.EventGroupsBundle) { b.EventGroup.Uuid = new("6f1c2c4e8a3b4f0e9d7a5b3c") }), false},
		{"stored with Loop's syncIdentifier", `{"date":1762000000123,"eventType":"Meal Bolus","insulin":2}`,
			bolus(func(b *sitedb.EventGroupsBundle) { b.EventGroup.SyncIdentifier = new("loop-dose-1") }), false},
	})
}

func TestDeviceStatusFallbackKey(t *testing.T) {
	key := DeviceStatusCollection{}.FallbackKey()
	require.False(t, key.whateverItsIdentifier)
	stored := &sitedb.DeviceReport{OccurredAt: keyDate, CreatedByDeviceName: new("openaps://samsung")}
	checkKey(t, key, []string{"date", "device"}, []keyCase[sitedb.DeviceReport]{
		{"same date and device", `{"date":1762000000123,"device":"openaps://samsung"}`, stored, true},
		{"no device, which finds any", `{"date":1762000000123}`, stored, true},
		{"another device", `{"date":1762000000123,"device":"loop://iPhone"}`, stored, false},
		{"a millisecond later", `{"date":1762000000124,"device":"openaps://samsung"}`, stored, false},
	})
}

func TestProfileFallbackKey(t *testing.T) {
	key := ProfileCollection{}.FallbackKey()
	require.False(t, key.whateverItsIdentifier)
	stored := &sitedb.ProfilePresetBundle{{OccurredAt: keyDate, CreatedByDeviceName: new("Loop")}}
	checkKey(t, key, []string{"date", "device"}, []keyCase[sitedb.ProfilePresetBundle]{
		{"same date and device", `{"date":1762000000123,"device":"Loop"}`, stored, true},
		{"the date as its startDate", `{"startDate":"2025-11-01T12:26:40.123Z","device":"Loop"}`, stored, true},
		{"no device, which finds any", `{"date":1762000000123}`, stored, true},
		{"another device", `{"date":1762000000123,"device":"Trio"}`, stored, false},
		{"a millisecond later", `{"date":1762000000124,"device":"Loop"}`, stored, false},
	})
}

// Food and activities have no fallback key: only an identifier finds one sent again
func TestNoFallbackKey(t *testing.T) {
	_, food := any(FoodCollection{}).(CollectionKeyed[nightscoutmodels.Food, sitedb.FoodPreset])
	require.False(t, food)
	_, activity := any(ActivityCollection{}).(CollectionKeyed[nightscoutmodels.Activity, sitedb.Activity])
	require.False(t, activity)
}
