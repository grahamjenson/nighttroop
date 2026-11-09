package clear_raw_json_test

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/clear_raw_json"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, with their specs, to run them straight away
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name string, spec string, cmd func(context.Context) error) {
	s[name+" "+spec] = cmd
}

// Each hour, at a fixed minute, every site is truncated, which does nothing for now
func TestTruncateSitesCron(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	jobs := scheduler{}
	clear_raw_json.SetupCron(jobs, controlDB)
	require.NoError(t, jobs["TruncateSites "+clear_raw_json.Spec](t.Context()))

	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(clear_raw_json.Spec)
	require.NoError(t, err)
	at := time.Date(2026, 11, 8, 10, 20, 0, 0, time.UTC)
	require.Equal(t, time.Date(2026, 11, 8, 11, 17, 0, 0, time.UTC), schedule.Next(at))
}
