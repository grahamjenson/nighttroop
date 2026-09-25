package nighttroop

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
	// ACTIVITIES
	///////

	SyncSourceFromDB(source *control.SiteSyncSource) *SiteSyncSource

	SyncSourceToUpdate(source *SiteSyncSource) control.UpdateSyncSourceParams

	SyncSourceToCreate(source *SiteSyncSource) control.CreateSyncSourceParams
}
