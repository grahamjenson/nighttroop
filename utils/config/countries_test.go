package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCountries(t *testing.T) {
	tests := []struct {
		raw     string
		want    []string
		wantErr bool
	}{
		{raw: "", want: []string{}},
		{raw: "NZ", want: []string{"NZ"}},
		{raw: " nz, AU ,", want: []string{"NZ", "AU"}},
		{raw: "NZL", wantErr: true},
		{raw: "N1", wantErr: true},
		{raw: "NZ,New Zealand", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseCountries(tt.raw)
		if tt.wantErr {
			require.Error(t, err, tt.raw)
			continue
		}
		require.NoError(t, err, tt.raw)
		require.Equal(t, tt.want, got, tt.raw)
	}
}
