package httputils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// A sync's Nightscout URL is refused when it is an address the internet cannot reach, so a site cannot make the
// server a proxy into the networks it is on. The same addresses are refused as a sync connects.
func TestNightscoutURLBlockedForAddressesNotOnTheInternet(t *testing.T) {
	cfg := &config.ServerConfig{HostName: "nighttroop.example"}
	tests := []struct {
		host    string
		blocked bool
	}{
		{"8.8.8.8", false},
		{"2001:4860:4860::8888", false},
		{"0.0.0.0", true},
		{"127.0.0.1", true},
		{"10.1.2.3", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"::1", true},
		{"fd00:ec2::254", true},
		{"fe80::1", true},
		// Shared address space, which carrier-grade NAT and Tailscale's tailnets use, and Alibaba Cloud's metadata
		{"100.64.0.1", true},
		{"100.100.100.200", true},
		{"100.127.255.255", true},
		{"100.63.255.255", false},
		{"100.128.0.0", false},
		{"nighttroop.example", true},
		{"abcdefgh.nighttroop.example", true},
	}
	for _, test := range tests {
		t.Run(test.host, func(t *testing.T) {
			err := IsNightscoutURLBlocked(test.host, cfg)
			if test.blocked {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
