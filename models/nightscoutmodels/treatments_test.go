package nightscoutmodels

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

func TestSyncIdentifier(t *testing.T) {
	omnipod := "74656d70426173616c20302e3520323032362d31312d30325430383a35323a30305a"
	tooLong := strings.Repeat("a", MaxSyncIdentifierLength+1)
	tests := []struct {
		name       string
		sent       string
		kept       *string
		identifier *types.IdentifierString
	}{
		{"a carb entry's UUID", `{"syncIdentifier": "6B9F1A1C-2D3E-4F50-8A1B-2C3D4E5F6071"}`, new("6B9F1A1C-2D3E-4F50-8A1B-2C3D4E5F6071"), types.ToIdentifierString("6B9F1A1C-2D3E-4F50-8A1B-2C3D4E5F6071")},
		{"an Omnipod dose's hex", `{"syncIdentifier": "` + omnipod + `"}`, new(omnipod), types.ToIdentifierString(omnipod)},
		{"too long to give back", `{"syncIdentifier": "` + tooLong + `"}`, nil, types.ToIdentifierString(tooLong)},
		{"a number", `{"syncIdentifier": 12345}`, nil, types.ToIdentifierString("12345")},
		{"empty", `{"syncIdentifier": ""}`, nil, nil},
		{"null", `{"syncIdentifier": null}`, nil, nil},
		{"an _id first", `{"_id": "abc", "syncIdentifier": "6B9F1A1C"}`, new("6B9F1A1C"), types.ToIdentifierString("abc")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var treatment Treatment
			require.NoError(t, json.Unmarshal([]byte(test.sent), &treatment))
			require.Equal(t, test.kept, treatment.SyncIdentifier)
			require.Equal(t, test.identifier, treatment.Identifier)
		})
	}
}
