package siteapiutils_test

import (
	"context"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events/sitemodels"
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

	var published []sitemodels.Event
	unsubscribe := cfg.SiteModelEvents.Subscribe(siteContext.Subdomain, func(event sitemodels.Event) { published = append(published, event) })
	defer unsubscribe()

	// An error rolls back the writes, and publishes none of its changes
	tx := siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	tx.Created("rolled back")
	require.Equal(t, apiutils.ErrBadBody, tx.Finish(apiutils.ErrBadBody))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.Error(t, err, "rolled back write should not exist")
	require.Empty(t, published)

	// Success commits the writes, then publishes the changes in order
	tx = siteapiutils.NewStoreTx(ctx, siteContext)
	write(tx)
	tx.Created("created")
	tx.Updated("updated")
	tx.Deleted("deleted")
	require.Empty(t, published, "nothing is published before the commit")
	require.Nil(t, tx.Finish(nil))
	_, err = controlDB.Queries.GetSyncSource(ctx, virtual)
	require.NoError(t, err, "committed write should exist")
	require.Equal(t, []sitemodels.Event{
		{SiteID: siteContext.Subdomain, Action: sitemodels.Created, Model: "created"},
		{SiteID: siteContext.Subdomain, Action: sitemodels.Updated, Model: "updated"},
		{SiteID: siteContext.Subdomain, Action: sitemodels.Deleted, Model: "deleted"},
	}, published)

	// A transaction that is never used opens nothing and finishes cleanly
	require.Nil(t, siteapiutils.NewStoreTx(ctx, siteContext).Finish(nil))
}

func ptr[T any](v T) *T { return &v }
