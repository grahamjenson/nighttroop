package sitesettings

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
)

// The handlers check a store can do what its routes need when the router is made; this checks it as it compiles
var (
	_ interface {
		siteapiutils.Store[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.Getter[sitemodels.UserIdentity]
		siteapiutils.Searcher[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.Creator[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.Updater[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
		siteapiutils.Deleter[sitemodels.UserIdentity, sitemodels.UserIdentity, *controldb.Queries]
	} = &IdentitiesStore{}
	_ interface {
		siteapiutils.Store[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.Searcher[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.Creator[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.Updater[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
		siteapiutils.Deleter[sitemodels.SyncSource, sitemodels.SyncSource, *controldb.Queries]
	} = &SyncSettingsStore{}
	_ interface {
		siteapiutils.Store[sitemodels.Settings, sitemodels.Settings, *controldb.Queries]
		siteapiutils.Getter[sitemodels.Settings]
		siteapiutils.Updater[sitemodels.Settings, sitemodels.Settings, *controldb.Queries]
	} = &SettingsStore{}
)
