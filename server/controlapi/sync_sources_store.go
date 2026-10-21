package controlapi

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

// SyncSourcesStore lets control identities see every site's sync sources and how their syncs are going
type SyncSourcesStore struct{}

func (s *SyncSourcesStore) SearchModels(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controlmodels.SyncSource, *apiutils.HTTPError) {
	sources, err := controlContext.ControlDB.Queries.SearchSyncSources(ctx, controlmodels.QueryToSearchSyncSources(query))
	return apiutils.Searched(sources, err, "sync sources", controlmodels.SyncSourceFromDB)
}

func (s *SyncSourcesStore) SearchHTMXResponse(models []*controlmodels.SyncSource, query *apiutils.Query) templ.Component {
	return control_sync_sources_templ.SyncSourcesTable(models, query.Skip, query.Limit)
}

// GetModel finds a sync source by its site::type
func (s *SyncSourcesStore) GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controlmodels.SyncSource, *apiutils.HTTPError) {
	siteID, sourceType, ok := strings.Cut(id, "::")
	if !ok {
		return nil, nil
	}
	source, err := controlContext.ControlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: consts.SyncSourceType(sourceType)})
	return apiutils.Found(source, err, "sync source", controlmodels.SyncSourceFromDB)
}

// GetHTMXResponse is the sync source's reports and recent errors, shown in a modal
func (s *SyncSourcesStore) GetHTMXResponse(model *controlmodels.SyncSource) templ.Component {
	return control_sync_sources_templ.SyncSourceModal(model)
}
