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
