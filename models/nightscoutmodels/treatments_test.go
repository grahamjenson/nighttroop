package nightscoutmodels

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A duration sent with a fraction of a minute, as Loop sends a dose ended early, keeps its whole minutes for its events
// and the rest in milliseconds, and is given back with the fraction; a whole one, or AAPS's with its milliseconds, is
// given back in whole minutes, and one out of range has no fraction
func TestDurationFraction(t *testing.T) {
	tests := []struct {
		name      string
		sent      string
		ms        *int64
		fraction  bool
		givenBack string // the duration's JSON, empty when it is left out
	}{
		{"Loop's bolus", `{"duration": 1.6666666666666667}`, new(int64(100000)), true, `1.6666666666666667`},
		{"Loop's suspend", `{"duration": 1.1747}`, new(int64(70482)), true, `1.1747`},
		{"under a minute", `{"duration": 0.667}`, new(int64(40020)), true, `0.667`},
		{"as a string", `{"duration": "0.5"}`, new(int64(30000)), true, `0.5`},
		{"whole minutes", `{"duration": 30}`, nil, false, `30`},
		{"AAPS's with milliseconds", `{"duration": 5, "durationInMilliseconds": 359587}`, new(int64(359587)), false, `5`},
		{"a fraction with milliseconds sent", `{"duration": 5.99, "durationInMilliseconds": 359587}`, new(int64(359587)), true, `5.993116666666666`},
		{"negative", `{"duration": -1.5}`, nil, false, `-1`},
		{"over a year", `{"duration": 600000.5}`, nil, false, `600000`},
		{"not a number", `{"duration": "NaN"}`, nil, false, ``},
		{"none", `{}`, nil, false, ``},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var treatment Treatment
			require.NoError(t, json.Unmarshal([]byte(test.sent), &treatment))
			require.Equal(t, test.ms, treatment.DurationInMilliseconds.Int64p())
			require.Equal(t, test.fraction, treatment.DurationFraction != nil && *treatment.DurationFraction)

			raw, err := json.Marshal(treatment)
			require.NoError(t, err)
			var out map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(raw, &out))
			if test.givenBack != "" || test.name == "none" {
				require.Equal(t, test.givenBack, string(out["duration"]))
			}
		})
	}
}

// Loop's syncIdentifier is given back as sent, while the identifier made from it is hashed when over 36 characters; one
// too long to give back is left out, and one that is not a string is only the identifier
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
