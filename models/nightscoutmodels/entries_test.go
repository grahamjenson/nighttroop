package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// An entry as AAPS uploads it is stored with its reading, type, trend, time, device and the identity that sent it
func TestEntryToCreate(t *testing.T) {
	var entry Entry
	require.NoError(t, json.Unmarshal([]byte(EntryFixture), &entry))
	assert.Equal(t, int64(0), entry.UtcOffsetMinutes.Int64())

	params := EntryToCreate(entry, ConverterContext{123, "identityName"})
	assert.NotEmpty(t, params.ID)
	assert.Equal(t, 56.0, params.Mgdl)
	assert.Equal(t, "sgv", params.ReadingType)
	assert.Equal(t, new(int64(5)), params.Trend, "FortyFiveDown")
	assert.Equal(t, int64(1733776878000), params.OccurredAt)
	assert.Equal(t, new("AAPS-DexcomG6"), params.CreatedByDeviceName)
	assert.Equal(t, "identityName", params.CreatedByIdentityName)
	assert.Equal(t, new(int64(123)), params.CreatedByIdentityID)
}

// Each trend number is a direction that reads back as the same number, and an entry gives it back as Nightscout's 0 to
// 9, with none for TripleUp and TripleDown. Another way of writing a direction is read as its number, an unknown one as
// NOT COMPUTABLE, and none or a number without a direction is none
func TestTrendDirection(t *testing.T) {
	trends := []struct {
		direction string
		entry     *types.FlexibleFloat64
	}{
		{"NONE", new(types.FlexibleFloat64(0))},
		{"DoubleUp", new(types.FlexibleFloat64(1))},
		{"SingleUp", new(types.FlexibleFloat64(2))},
		{"FortyFiveUp", new(types.FlexibleFloat64(3))},
		{"Flat", new(types.FlexibleFloat64(4))},
		{"FortyFiveDown", new(types.FlexibleFloat64(5))},
		{"SingleDown", new(types.FlexibleFloat64(6))},
		{"DoubleDown", new(types.FlexibleFloat64(7))},
		{"NOT COMPUTABLE", new(types.FlexibleFloat64(8))},
		{"RATE OUT OF RANGE", new(types.FlexibleFloat64(9))},
		{"TripleUp", nil},
		{"TripleDown", nil},
	}
	for trend, tt := range trends {
		t.Run(tt.direction, func(t *testing.T) {
			number := int64(trend)
			assert.Equal(t, tt.direction, *TrendDirection(&number))
			assert.Equal(t, number, *DirectionTrend(&tt.direction))
			assert.Equal(t, tt.entry, EntryTrend(&number))
		})
	}

	written := []struct {
		direction *string
		trend     *int64
	}{
		{nil, nil},
		{new(""), nil},
		{new("singleup"), new(int64(2))},
		{new("↗"), new(int64(3))},
		{new("Sideways"), new(int64(8))},
	}
	for _, tt := range written {
		name := "nil"
		if tt.direction != nil {
			name = fmt.Sprintf("%q", *tt.direction)
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.trend, DirectionTrend(tt.direction))
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
