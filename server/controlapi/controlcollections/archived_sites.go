package controlcollections

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/s3utils"
)

// ArchivedSitesCollection lists the archived sites, and lets admins hydrate one, bringing it back live from its archive
type ArchivedSitesCollection struct{}

func (s *ArchivedSitesCollection) GetName() string {
	return "archived sites"
}

// GetRoute is where the collection is in its URL, /archived-sites
func (s *ArchivedSitesCollection) GetRoute() string {
	return "archived-sites"
}

func (s *ArchivedSitesCollection) ToModel(row *controldb.Site) *controlmodels.Site {
	return controlmodels.SiteFromDB(row)
}

// SearchRows is the archived sites, the most recently archived first
func (s *ArchivedSitesCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controldb.Site, error) {
	return controlContext.ControlDB.Queries.SearchArchivedSites(ctx, controldb.SearchArchivedSitesParams(controlmodels.QueryToSearchSites(query)))
}

func (s *ArchivedSitesCollection) SearchHTMXResponse(models []*controlmodels.Site, query *apiutils.Query) templ.Component {
	return control_sites_templ.ArchivedSitesTable(models, query.Skip, query.Limit)
}

func (s *ArchivedSitesCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controldb.Site, error) {
	return controlContext.ControlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: id})
}

// UpdateRow hydrates an archived site, changed to live: its archive is downloaded from S3, so without S3 it is refused
func (s *ArchivedSitesCollection) UpdateRow(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site, model *controlmodels.Site) (*controldb.Site, error) {
	config := controlContext.ControlDB.ServerConfig
	switch {
	case model.State != "live":
		return nil, apiutils.ErrBadBodyWithHint("an archived site's state can only be changed to live")
	case oldModel.State != "archived":
		return nil, apiutils.ErrConstraintCheckFailedWithHint("only an archived site can be hydrated")
	case !config.Status.S3Working:
		return nil, apiutils.ErrConstraintCheckFailedWithHint("a site can only be hydrated when S3 is connected")
	}
	started := time.Now()
	if err := controlContext.ControlDB.HydrateSite(ctx, oldModel.ID, s3utils.NewS3Client(config).DownloadFilesFromS3); err != nil {
		return nil, err
	}
	controlContext.ControlDB.Logs.SiteHydratedEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name, started)
	return s.GetRow(ctx, oldModel.ID, controlContext)
}

func (s *ArchivedSitesCollection) UpdateHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteHydrated(model)
}
