package background

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/utils/testutils"
)

// Stop waits for a running cron job, such as a backup, so it does not overlap the final backup or the DBs closing
func TestStopWaitsForRunningCronJob(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	b := &BackGroundManager{cron: cron.NewCron(controlDB.Jobs)}
	started := make(chan struct{}, 1)
	var finished atomic.Bool
	b.cron.AddFunc("slow", "@every 1s", func(context.Context) error {
		select {
		case started <- struct{}{}:
		default:
		}
		time.Sleep(500 * time.Millisecond)
		finished.Store(true)
		return nil
	})
	b.Start()
	<-started

	b.Stop()
	require.True(t, finished.Load())
}
