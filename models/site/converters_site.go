package site

import (
	"nighttroop.t1d.nz/data/sql/controldb"
)

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:format function
// goverter:output:file ./z_site_converters.go
// goverter:useZeroValueOnPointerInconsistency
type SiteConverter interface {

	///////
	// SYNC SOURCE
	///////

	SyncSourceFromDB(source *controldb.SiteSyncSource) *SyncSource

	SyncSourceToUpdate(source *SyncSource) controldb.UpdateSyncSourceParams

	SyncSourceToCreate(source *SyncSource) controldb.CreateSyncSourceParams
}
