package apiutils_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/apiutils"
)

func TestFormToMap(t *testing.T) {
	type model struct {
		Enabled    bool     `json:"enabled"`
		Off        bool     `json:"off"`
		Count      int      `json:"count"`
		Name       *string  `json:"name"`
		Rate       *float64 `json:"rate"`
		Minutes    *int64   `json:"minutes"`
		Notes      *string  `json:"notes"`
		OccurredAt int64    `json:"occurred_at"`
		Apps       []string `json:"apps"`
		BundleIDs  []string `json:"bundle_ids"`
		Tags       []string `json:"tags"`
	}
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	tests := []struct {
		name     string
		form     url.Values
		location *time.Location
		// want is the whole map, so a field not in it was not sent; a bool is always sent, false when not ticked
		want map[string]any
	}{
		{
			name:     "a checkbox is a bool, a number is a float, a string is kept, and a dotted name nests",
			form:     url.Values{"enabled": {"on"}, "count": {"3"}, "name": {"true"}, "rate": {"0"}, "a.b": {"c"}},
			location: time.UTC,
			want: map[string]any{
				"enabled": true, "off": false, "count": 3.0, "name": "true", "rate": 0.0, "a": map[string]any{"b": "c"},
			},
		},
		{
			// A number input left blank is not sent, so a JSON decode of it does not fail and an update leaves it as
			// it is
			name:     "a blank number is not sent, a blank string is",
			form:     url.Values{"minutes": {""}, "rate": {"1.5"}, "notes": {""}},
			location: time.UTC,
			want:     map[string]any{"enabled": false, "off": false, "rate": 1.5, "notes": ""},
		},
		{
			name:     "a date and time is the milliseconds of that time in the timezone times are shown in",
			form:     url.Values{"occurred_at": {"2026-10-31T09:30"}},
			location: auckland,
			want: map[string]any{
				"enabled": false, "off": false, "occurred_at": time.Date(2026, 10, 31, 9, 30, 0, 0, auckland).UnixMilli(),
			},
		},
		{
			name:     "a list is every value sent, each split at commas and spaces, and one not sent is left out",
			form:     url.Values{"apps": {"loop", "trio"}, "bundle_ids": {"a.b, c.d  e.f,"}},
			location: time.UTC,
			want: map[string]any{
				"enabled": false, "off": false, "apps": []string{"loop", "trio"}, "bundle_ids": []string{"a.b", "c.d", "e.f"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, apiutils.FormToMap[model](tt.form, tt.location))
		})
	}
}
