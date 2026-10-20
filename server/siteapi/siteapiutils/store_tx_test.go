package siteapiutils_test

import (
	"context"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils"

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

	// An error rolls back the writes
	tx := siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	require.Equal(t, apiutils.ErrBadBody, tx.Finish(apiutils.ErrBadBody))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.Error(t, err, "rolled back write should not exist")

	// Success commits the writes
	tx = siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	require.Nil(t, tx.Finish(nil))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.NoError(t, err, "committed write should exist")

	// A transaction that is never used opens nothing and finishes cleanly
	require.Nil(t, siteapiutils.NewStoreTx(ctx, siteContext).Finish(nil))
}

func ptr[T any](v T) *T { return &v }
