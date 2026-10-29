package config

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTrustedProxies(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{name: "the default loopback", raw: "127.0.0.1,::1", want: []string{"127.0.0.1/32", "::1/128"}},
		{name: "a docker network range", raw: "172.30.0.0/24", want: []string{"172.30.0.0/24"}},
		{name: "a range is masked to its network", raw: "172.30.0.9/24", want: []string{"172.30.0.0/24"}},
		{name: "spaces and empty entries", raw: " 172.30.0.2 , ,::1", want: []string{"172.30.0.2/32", "::1/128"}},
		{name: "none", raw: "", want: []string{}},
		{name: "a hostname is refused", raw: "caddy", wantErr: true},
		{name: "a bad range is refused", raw: "172.30.0.0/99", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTrustedProxies(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			want := []netip.Prefix{}
			for _, w := range tt.want {
				want = append(want, netip.MustParsePrefix(w))
			}
			require.Equal(t, want, got)
		})
	}
}
