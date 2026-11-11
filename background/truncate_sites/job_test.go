package truncate_sites_test

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/truncate_sites"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, with their specs, to run them straight away
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name string, spec string, cmd func(context.Context) error) {
	s[name+" "+spec] = cmd
}

// Every 10 minutes, at fixed minutes, every site is truncated
func TestTruncateSitesCron(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	jobs := scheduler{}
	truncate_sites.SetupCron(jobs, controlDB)
	require.NoError(t, jobs["TruncateSites "+truncate_sites.Spec](t.Context()))

	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(truncate_sites.Spec)
	require.NoError(t, err)
	at := time.Date(2026, 11, 8, 10, 20, 0, 0, time.UTC)
	require.Equal(t, time.Date(2026, 11, 8, 10, 30, 0, 0, time.UTC), schedule.Next(at))
}
