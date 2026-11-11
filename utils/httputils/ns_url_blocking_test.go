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
		// A fully qualified name's trailing dot, or capitals, still names this server
		{"nighttroop.example.", true},
		{"abcdefgh.ns.nighttroop.example.", true},
		{"ABCDEFGH.NS.NightTroop.Example.", true},
		{"localhost.", true},
		{"8.8.8.8.", false},
		// "This network", which Linux connects to as the local host
		{"0.1.2.3", true},
		{"0.255.255.255", true},
		{"1.0.0.0", false},
		// IETF protocol assignments, benchmarking, reserved and broadcast
		{"192.0.0.8", true},
		{"192.0.1.1", false},
		{"198.18.0.1", true},
		{"198.19.255.255", true},
		{"198.20.0.0", false},
		{"240.0.0.1", true},
		{"255.255.255.255", true},
		// Multicast, IPv4 and IPv6, including link- and interface-local
		{"224.0.0.1", true},
		{"239.255.255.250", true},
		{"ff02::1", true},
		{"ff01::1", true},
		{"ff0e::1", true},
		// IPv4-mapped IPv6 reaches the IPv4 address
		{"::ffff:10.0.0.5", true},
		{"::ffff:127.0.0.1", true},
		{"::ffff:8.8.8.8", false},
		// NAT64 reaches the IPv4 address in its last 4 bytes, so a private one is refused and a public one is not
		{"64:ff9b::a00:5", true},
		{"64:ff9b::7f00:1", true},
		{"64:ff9b::a9fe:a9fe", true},
		{"64:ff9b::808:808", false},
		// Local-use NAT64 is only ever routed inside a network
		{"64:ff9b:1::a00:5", true},
		{"64:ff9b:1::808:808", true},
		// 6to4 reaches the IPv4 address in its bytes 2 to 5
		{"2002:a00:5::1", true},
		{"2002:c0a8:101::1", true},
		{"2002:808:808::1", false},
		// Public addresses still pass
		{"1.1.1.1", false},
		{"93.184.215.14", false},
		{"2606:4700:4700::1111", false},
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
