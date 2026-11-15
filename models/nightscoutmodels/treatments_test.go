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

// A timestamp in milliseconds in the second of created_at, as xDrip+ sends, is the treatment's date to the
// millisecond; one in another second, or not a number, leaves the date created_at's
func TestTimestampMilliseconds(t *testing.T) {
	tests := []struct {
		name string
		sent string
		date int64
	}{
		{"xDrip+'s timestamp", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": 1794512149419}`, 1794512149419},
		{"no timestamp", `{"created_at": "2026-11-12T19:35:49Z"}`, 1794512149000},
		{"a timestamp in another second", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": 1794512150419}`, 1794512149000},
		{"a timestamp string", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": "2026-11-12T19:35:49.419Z"}`, 1794512149000},
		{"only a timestamp", `{"timestamp": 1794512149419}`, 1794512149419},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treatment Treatment
			require.NoError(t, json.Unmarshal([]byte(tt.sent), &treatment))
			require.Equal(t, tt.date, treatment.Date.Int64())
		})
	}
}

// A known glucose type is kept as its type, and any other as sent, at most 32 characters, rather than refuse it
func TestGlucoseType(t *testing.T) {
	tests := []struct {
		name string
		sent string
		kept GlucoseType
	}{
		{"a finger stick", `"Finger"`, Finger},
		{"a sensor reading", `"Sensor"`, Sensor},
		{"a meter reading", `"Meter"`, Meter},
		{"AAPS's and iAPS's typed in finger stick", `"Manual"`, Manual},
		{"lower case", `"manual"`, Manual},
		{"xDrip4iOS's reading after a colon", `"Finger: 6.7 mmol/L"`, Finger},
		{"another type", `"Lab"`, "Lab"},
		{"too long", `"` + strings.Repeat("a", 40) + `"`, GlucoseType(strings.Repeat("a", 32))},
		{"too long in the middle of a character", `"` + strings.Repeat("a", 31) + `é"`, GlucoseType(strings.Repeat("a", 31))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treatment Treatment
			require.NoError(t, json.Unmarshal([]byte(`{"eventType": "BG Check", "glucose": 120, "glucoseType": `+tt.sent+`}`), &treatment))
			require.Equal(t, tt.kept, *treatment.GlucoseType)
		})
	}
}
