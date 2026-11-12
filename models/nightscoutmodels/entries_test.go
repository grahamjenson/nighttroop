package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

// A trend rate is a change, so it is in mg/dL a minute unless it is sent in mmol/L, and not guessed from its size
func Test_entryTrendRate(t *testing.T) {
	tests := []struct {
		name string
		json string
		want float64
	}{
		{"mg/dL", `{"type": "sgv", "sgv": 144, "trendRate": 1}`, 1},
		{"mg/dL sent", `{"type": "sgv", "sgv": 144, "units": "mg/dL", "trendRate": -3.5}`, -3.5},
		{"mmol/L sent", `{"type": "sgv", "sgv": 8, "units": "mmol/L", "trendRate": 0.1}`, 1.80182},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry Entry
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &entry))
			assert.InDelta(t, tt.want, entry.TrendRate.Float64(), 0.01)
		})
	}
}
