package siteapiutils_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestStoreTx(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	siteContext := &httputils.SiteContext{Config: cfg, ControlDB: controlDB, SiteDB: site.SiteDB, Subdomain: site.SiteDB.Subdomain}
	ctx := context.Background()
	virtual := controldb.GetSyncSourceParams{Type: consts.SyncSourceTypeVirtual, SiteID: siteContext.Subdomain}

	write := func(tx *siteapiutils.StoreTx) {
		q, err := tx.Control()
		require.NoError(t, err)
		require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{
			SiteID: siteContext.Subdomain, Type: consts.SyncSourceTypeVirtual,
			UrlOrUsernameEnc: ptr("virtual"), PasswordEnc: ptr("virtual"),
		}))
	}
	hookRan := false
	hook := func(ctx context.Context) *siteapiutils.HTTPError { hookRan = true; return nil }

	// An error rolls back the writes and skips the hooks
	tx := siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	tx.AfterCommit(hook)
	require.Equal(t, siteapiutils.ErrBadBody, tx.Finish(siteapiutils.ErrBadBody))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.Error(t, err, "rolled back write should not exist")
	require.False(t, hookRan)

	// Success commits the writes, then runs the hooks
	tx = siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	tx.AfterCommit(hook)
	require.Nil(t, tx.Finish(nil))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.NoError(t, err, "committed write should exist")
	require.True(t, hookRan)

	// A transaction that is never used opens nothing and finishes cleanly
	require.Nil(t, siteapiutils.NewStoreTx(ctx, siteContext).Finish(nil))
}

func ptr[T any](v T) *T { return &v }
