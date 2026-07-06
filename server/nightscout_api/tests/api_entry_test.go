package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
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

func TestAPIEntry(t *testing.T) {
	cfg, controlDB, userDB := testutils.GetConfigControlDBAndSiteDBForTest()
	path := fmt.Sprintf("%s://ns.%s.%s/api/v3/entries", cfg.HostProtocol, userDB.Subdomain, cfg.HostNameWithPort)
	mainHandler := server.GetMainHandler(controlDB, cfg)
	for _, entry := range testAPIEntries {
		t.Run(entry.name, func(t *testing.T) {
			for i, input := range entry.inputs {
				expected := entry.expected[i]
				statusCode, output, _ := testutils.JSONRequest[map[string]any](t, mainHandler, "POST", path, input)
				require.Equal(t, expected.statusCode, statusCode, "Status code should match expected")
				require.Equal(t, expected.output, output, "Response body should match expected")
			}
		})
	}
}
