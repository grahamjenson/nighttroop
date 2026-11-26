package site_banners_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/site_banners"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/testutils"
)

type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name string, _ string, cmd func(context.Context) error) { s[name] = cmd }
func (s scheduler) AddFuncSoon(name string, _, _ time.Duration, cmd func(context.Context) error) {
	s[name] = cmd
}

// The job deletes the banners past their time, and keeps the ones still to show
func TestExpiredSiteBannersAreDeleted(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, q := t.Context(), controlDB.Queries
	for message, expires := range map[string]time.Duration{"over an hour ago": -time.Hour, "over a second ago": -time.Second, "in an hour": time.Hour} {
		_, err := q.CreateSiteBanner(ctx, controldb.CreateSiteBannerParams{Message: message, Level: "info", SiteIdsJson: "[]",
			ExpiresAt: time.Now().Add(expires).UnixMilli(), CreatedByIdentityName: "admin"})
		require.NoError(t, err)
	}

	jobs := scheduler{}
	site_banners.SetupCron(jobs, controlDB)
	require.NoError(t, jobs["DeleteExpiredSiteBanners"](ctx))
	kept, err := q.AllSiteBanners(ctx)
	require.NoError(t, err)
	require.Len(t, kept, 1)
	require.Equal(t, "in an hour", kept[0].Message)
}
