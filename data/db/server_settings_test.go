package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/testutils"
)

// The rate limits a control admin sets are kept, so the server uses them again from its next start
func TestRateLimitsSetAreLoadedAgain(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	ctx := t.Context()
	require.NoError(t, controlDB.SetRateLimits(ctx, map[string]config.Rate{"site_push": {PerMinute: 3, Burst: 4}}, "admin"))

	// As a new start would be: the env's, until the kept ones are loaded
	cfg.RateLimiters.SetRates(nil)
	require.False(t, cfg.RateLimiters.Settings()[11].Set)
	require.NoError(t, controlDB.LoadRateLimits(ctx))
	setting := cfg.RateLimiters.Settings()[11]
	require.Equal(t, "site_push", setting.Name)
	require.True(t, setting.Set)
	require.Equal(t, config.Rate{PerMinute: 3, Burst: 4}, setting.Rate)

	settings, err := controlDB.Queries.GetServerSettings(ctx)
	require.NoError(t, err)
	require.JSONEq(t, `{"site_push": {"per_minute": 3, "burst": 4}}`, settings.RateLimitsJson)
	require.Equal(t, "admin", *settings.UpdatedByIdentityName)
}
