package nightscout_api_tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/nightscout_api"
	"nighttroop.t1d.nz/utils/testutils"
)

type testAPIEntryResult struct {
	output     map[string]any
	statusCode int
}

type testAPIEntry struct {
	name     string
	inputs   []map[string]any
	expected []testAPIEntryResult
}

var testAPIEntries = []testAPIEntry{
	{
		name: "Test creating entries with date",
		inputs: []map[string]any{
			{"date": 1},
			{"date": 2},
			{"date": 3},
		},
		expected: []testAPIEntryResult{
			{
				output: map[string]any{
					"_id":  "1",
					"date": 1,
				},
				statusCode: 200,
			},
			{
				output: map[string]any{
					"_id":  "2",
					"date": 2,
				},
				statusCode: 200,
			},
			{
				output: map[string]any{
					"_id":  "3",
					"date": 3,
				},
				statusCode: 200,
			},
		},
	},
}

func test_api_entry(t *testing.T) {
	cfg, controlDB, userDB := testutils.GetConfigControlDBAndUserDBForTest()
	router := nightscout_api.NewNightscoutRouter()
	for _, entry := range testAPIEntries {
		t.Run(entry.name, func(t *testing.T) {
			for i, input := range entry.inputs {
				expected := entry.expected[i]
				statusCode, output, _ := testutils.JSONRequest[map[string]any](t, router, "POST", "/api/v3/entries", input)
				require.Equal(t, expected.statusCode, statusCode, "Status code should match expected")
				require.Equal(t, expected.output, output, "Response body should match expected")
			}
		})
	}
}
