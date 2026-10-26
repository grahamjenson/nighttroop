package controlcollections

import (
	"context"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sync_sources_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// SyncSourcesCollection lets control identities see every site's sync sources and how their syncs are going
type SyncSourcesCollection struct{}

func (s *SyncSourcesCollection) GetName() string {
	return "sync sources"
}

// GetRoute is where the collection is in its URL, /sync-sources
func (s *SyncSourcesCollection) GetRoute() string {
	return "sync-sources"
}

func (s *SyncSourcesCollection) ToModel(row *controldb.SiteSyncSource) *controlmodels.SyncSource {
	return controlmodels.SyncSourceFromDB(row)
}

func (s *SyncSourcesCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controldb.SiteSyncSource, error) {
	return controlContext.ControlDB.Queries.SearchSyncSources(ctx, controlmodels.QueryToSearchSyncSources(query))
}

func (s *SyncSourcesCollection) SearchHTMXResponse(models []*controlmodels.SyncSource, query *apiutils.Query) templ.Component {
	return control_sync_sources_templ.SyncSourcesTable(models, query.Skip, query.Limit)
}

// GetRow finds a sync source by its site::type
func (s *SyncSourcesCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controldb.SiteSyncSource, error) {
	siteID, sourceType, ok := strings.Cut(id, "::")
	if !ok {
		return nil, nil
	}
	return controlContext.ControlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: consts.SyncSourceType(sourceType)})
}

// GetHTMXResponse is the sync source's reports and recent errors, shown in a modal
func (s *SyncSourcesCollection) GetHTMXResponse(model *controlmodels.SyncSource) templ.Component {
	return control_sync_sources_templ.SyncSourceModal(model)
}
