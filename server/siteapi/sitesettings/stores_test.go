package sitesettings

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
)

// The handlers check a store can do what its routes need when the router is made; this checks it as it compiles
var (
	_ interface {
		siteapiutils.RowStore[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.RowGetter[sitemodels.UserIdentity]
		siteapiutils.RowSearcher[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.RowCreator[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.RowUpdater[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.RowDeleter[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
	} = &IdentitiesStore{}
	_ interface {
		siteapiutils.RowStore[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.RowSearcher[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.RowCreator[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.RowUpdater[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.RowDeleter[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
	} = &SyncSettingsStore{}
	_ interface {
		siteapiutils.RowStore[sitemodels.Settings, sitemodels.Settings, *controldb.Queries]
		siteapiutils.RowGetter[sitemodels.Settings]
		siteapiutils.RowUpdater[sitemodels.Settings, sitemodels.Settings, *controldb.Queries]
	} = &SettingsStore{}
)
