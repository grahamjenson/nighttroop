package sitesettings

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
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
