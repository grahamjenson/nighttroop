package authclients

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecentKeepsEachIdentitysNewestClients(t *testing.T) {
	recent, now := New(10), time.Now()

	// The same app on the same network is one client, moved to the front when it authenticates again
	recent.Add("site/1", "AndroidAPS/3.2", "203.0.113.7", now)
	recent.Add("site/1", "xDrip+/2024", "203.0.113.9:4321", now.Add(time.Second))
	recent.Add("site/1", "AndroidAPS/3.2", "203.0.113.8", now.Add(2*time.Second))
	clients := recent.Get("site/1")
	require.Len(t, clients, 2)
	require.Equal(t, Client{UserAgent: "AndroidAPS/3.2", Network: "203.0.113.0/24", LastAt: now.Add(2 * time.Second).UnixMilli()}, clients[0])
	require.Equal(t, "xDrip+/2024", clients[1].UserAgent)

	// Another identity's clients are its own
	require.Empty(t, recent.Get("site/2"))
	require.Empty(t, recent.Get("other/1"))

	// Only the newest few are kept, however many there are
	for i := range 20 {
		recent.Add("site/1", fmt.Sprintf("app-%d", i), "203.0.113.1", now)
	}
	clients = recent.Get("site/1")
	require.Len(t, clients, maxClients)
	require.Equal(t, "app-19", clients[0].UserAgent)
}

func TestRecentDropsTheLeastRecentlyUsedIdentity(t *testing.T) {
	recent, now := New(2), time.Now()
	recent.Add("site/1", "a", "", now)
	recent.Add("site/2", "a", "", now)
	recent.Add("site/3", "a", "", now)
	require.Empty(t, recent.Get("site/1"))
	require.Len(t, recent.Get("site/3"), 1)
}

func TestNetwork(t *testing.T) {
	tests := []struct {
		ip      string
		network string
	}{
		{"203.0.113.7", "203.0.113.0/24"},
		{"203.0.113.7:1234", "203.0.113.0/24"},
		{"::ffff:203.0.113.7", "203.0.113.0/24"},
		{"2001:db8:abcd:12::1", "2001:db8:abcd::/48"},
		{"[2001:db8:abcd:12::1]:443", "2001:db8:abcd::/48"},
		{"not an ip", ""},
		{"", ""},
	}
	for _, test := range tests {
		require.Equal(t, test.network, network(test.ip), test.ip)
	}
}

func TestShared(t *testing.T) {
	now := time.Now()
	recently, long := now.Add(-time.Hour).UnixMilli(), now.Add(-2*SharedWithin).UnixMilli()
	tests := []struct {
		name    string
		clients []Client
		shared  bool
	}{
		{"none", nil, false},
		{"one app", []Client{{UserAgent: "a", LastAt: recently}}, false},
		{"one app on two networks", []Client{{UserAgent: "a", Network: "1", LastAt: recently}, {UserAgent: "a", Network: "2", LastAt: recently}}, false},
		{"two apps", []Client{{UserAgent: "a", LastAt: recently}, {UserAgent: "b", LastAt: recently}}, true},
		{"an app since replaced", []Client{{UserAgent: "a", LastAt: recently}, {UserAgent: "b", LastAt: long}}, false},
	}
	for _, test := range tests {
		require.Equal(t, test.shared, Shared(test.clients, now), test.name)
	}
}
