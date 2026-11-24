package check_resources

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// A check finds the DBs' disk nearly full when less than DiskRefuseFreeMB is free, so sites are not created and their
// writes are refused, and finds it has room again once more is free
func TestCheckRefusesWhileDiskNearlyFull(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	refuse := cfg.DiskRefuseFreeMB
	t.Cleanup(func() { cfg.DiskRefuseFreeMB = refuse })

	// No disk has this many MB free
	cfg.DiskRefuseFreeMB = 1 << 40
	require.NoError(t, Check(t.Context(), cfg, controlDB, time.Now()))
	require.True(t, controlDB.DiskFull())

	cfg.DiskRefuseFreeMB = 0
	require.NoError(t, Check(t.Context(), cfg, controlDB, time.Now().Add(time.Second)))
	require.False(t, controlDB.DiskFull())
}
