package site

import (
	"nighttroop.t1d.nz/data/sql/control"
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

	SyncSourceFromDB(source *control.SiteSyncSource) *SyncSource

	SyncSourceToUpdate(source *SyncSource) control.UpdateSyncSourceParams

	SyncSourceToCreate(source *SyncSource) control.CreateSyncSourceParams
}
