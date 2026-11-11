package geoip

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocate(t *testing.T) {
	// A path that cannot be opened falls back to the built in country database
	for _, locator := range []*Locator{New(""), New("does-not-exist.mmdb")} {
		tests := []struct {
			ip   string
			want Location
		}{
			{ip: "8.8.8.8", want: Location{Country: "US"}},
			{ip: "202.36.163.1", want: Location{Country: "NZ"}},
			{ip: "127.0.0.1", want: Location{}},
			{ip: "10.0.0.1", want: Location{}},
			{ip: "not an ip", want: Location{}},
			{ip: "", want: Location{}},
		}
		for _, tt := range tests {
			require.Equal(t, tt.want, locator.Locate(tt.ip), tt.ip)
		}
	}
	var none *Locator
	require.Equal(t, Location{}, none.Locate("8.8.8.8"))
}
