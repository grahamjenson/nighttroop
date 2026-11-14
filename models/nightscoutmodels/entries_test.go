package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/types"
)

const EntryFixture = `  {
    "_id": "67575609fecbdb404acf9c3d",
    "device": "AAPS-DexcomG6",
    "date": 1733776878000,
    "dateString": "2024-12-09T20:41:18.000Z",
    "isValid": true,
    "sgv": 56,
    "direction": "FortyFiveDown",
    "type": "sgv",
    "created_at": "2024-12-09T20:41:45.284Z"
  }`

func Test_JSON_entry(t *testing.T) {
	var entry Entry
	err := json.Unmarshal([]byte(EntryFixture), &entry)
	assert.NoError(t, err)

	assert.Equal(t, int64(0), entry.UtcOffsetMinutes.Int64())

	entryJson, err := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("entry JSON: %s %v\n", entryJson, err)
}

func Test_entryToCreateParams(t *testing.T) {
	var entry Entry
	err := json.Unmarshal([]byte(EntryFixture), &entry)
	assert.NoError(t, err)

	createParams := EntryToCreate(entry, ConverterContext{123, "identityName"})

	assert.NotNil(t, createParams)

	paramsJson, err := json.MarshalIndent(createParams, "", "  ")
	fmt.Printf("CreateentryParams JSON: %s %v\n", paramsJson, err)
	fmt.Println(" ")
	entryJson, err := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("entry JSON: %s %v\n", entryJson, err)
}

func TestDirectionTrend(t *testing.T) {
	tests := []struct {
		direction *string
		want      *int64
	}{
		{nil, nil},
		{new(""), nil},
		{new("NONE"), new(int64(0))},
		{new("DoubleUp"), new(int64(1))},
		{new("singleup"), new(int64(2))},
		{new("↗"), new(int64(3))},
		{new("Flat"), new(int64(4))},
		{new("FortyFiveDown"), new(int64(5))},
		{new("SingleDown"), new(int64(6))},
		{new("DoubleDown"), new(int64(7))},
		{new("NOT COMPUTABLE"), new(int64(8))},
		{new("RATE OUT OF RANGE"), new(int64(9))},
		{new("TripleUp"), new(int64(10))},
		{new("TripleDown"), new(int64(11))},
		{new("Sideways"), new(int64(8))},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.direction), func(t *testing.T) {
			assert.Equal(t, tt.want, DirectionTrend(tt.direction))
		})
	}
}

func TestTrendDirection(t *testing.T) {
	tests := []struct {
		trend int64
		want  string
		entry *types.FlexibleFloat64
	}{
		{0, "NONE", new(types.FlexibleFloat64(0))},
		{1, "DoubleUp", new(types.FlexibleFloat64(1))},
		{3, "FortyFiveUp", new(types.FlexibleFloat64(3))},
		{4, "Flat", new(types.FlexibleFloat64(4))},
		{7, "DoubleDown", new(types.FlexibleFloat64(7))},
		{8, "NOT COMPUTABLE", new(types.FlexibleFloat64(8))},
		{9, "RATE OUT OF RANGE", new(types.FlexibleFloat64(9))},
		{10, "TripleUp", nil},
		{11, "TripleDown", nil},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, *TrendDirection(&tt.trend))
			assert.Equal(t, tt.trend, *DirectionTrend(&tt.want))
			assert.Equal(t, tt.entry, EntryTrend(&tt.trend))
		})
	}
	for _, trend := range []*int64{nil, new(int64(-1)), new(int64(12))} {
		assert.Nil(t, TrendDirection(trend))
	}
}

func TestEntryTrendSentAlone(t *testing.T) {
	tests := []struct {
		json string
		want *string
	}{
		{`{"trend": 3}`, new("FortyFiveUp")},
		{`{"trend": "3"}`, new("FortyFiveUp")},
		{`{"trend": 0}`, new("NONE")},
		{`{"trend": 9}`, new("RATE OUT OF RANGE")},
		{`{"trend": 10}`, nil},
		{`{"trend": -1}`, nil},
		{`{"trend": 2.5}`, nil},
		{`{"trend": 3, "direction": "Flat"}`, new("Flat")},
		{`{"trend": 3, "direction": ""}`, new("FortyFiveUp")},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			var entry Entry
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &entry))
			assert.Equal(t, tt.want, entry.Direction)
		})
	}
}
