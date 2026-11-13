package sitesettings

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// A Nightscout URL is saved without a fully qualified name's trailing dot, so it cannot name this server, or be
// another URL for a source already synced, by adding one
func TestNightscoutURLSavedWithoutTrailingDot(t *testing.T) {
	cfg := &config.ServerConfig{HostName: "nighttroop.example", Mode: consts.Production}

	_, err := nsURLNormalizeValidate(new("https://abcdefgh.ns.nighttroop.example./"), cfg)
	require.ErrorContains(t, err, "Nightscout cannot be on this host.")

	normalized, err := nsURLNormalizeValidate(new("https://8.8.8.8./"), cfg)
	require.NoError(t, err)
	require.Equal(t, "https://8.8.8.8", *normalized)
}

// A Nightscout URL's host MAC is the same for each way of writing one host and port, and differs for another host or
// port, so the control site counts the sites syncing from one host. Other connectors have none.
func TestNightscoutHostMAC(t *testing.T) {
	cfg := &config.ServerConfig{HostName: "nighttroop.example", Mode: consts.Development}
	sec := security.NewSecurityConfig("", consts.Test, "localhost", false)
	macOf := func(rawURL string) string {
		normalized, err := nsURLNormalizeValidate(&rawURL, cfg)
		require.NoError(t, err, rawURL)
		mac, err := hostMAC(sec, consts.ConnectorTypeNightscout, normalized)
		require.NoError(t, err, rawURL)
		require.NotContains(t, *mac, "ns.example")
		return *mac
	}

	same := macOf("https://ns.example")
	for _, rawURL := range []string{"https://NS.Example/", "https://ns.example.", "https://ns.example:443", "http://ns.example", "http://ns.example:80/", "https://user:pass@ns.example"} {
		require.Equal(t, same, macOf(rawURL), rawURL)
	}
	for _, rawURL := range []string{"https://other.example", "https://ns.example:8443", "https://sub.ns.example", "http://ns.example:443"} {
		require.NotEqual(t, same, macOf(rawURL), rawURL)
	}

	mac, err := hostMAC(sec, consts.ConnectorTypeDexcom, new("dexcom-user"))
	require.NoError(t, err)
	require.Nil(t, mac)
}
