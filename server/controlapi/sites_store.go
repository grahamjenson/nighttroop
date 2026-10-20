package controlapi

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// SitesStore lets control identities see every site, whatever its state, and admins disable live ones
type SitesStore struct{}

func (s *SitesStore) SearchModels(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controlmodels.Site, *apiutils.HTTPError) {
	sites, err := controlContext.ControlDB.Queries.SearchSites(ctx, controlmodels.QueryToSearchSites(query))
	if err != nil {
		slog.Error("Search sites", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*controlmodels.Site{}
	for _, site := range sites {
		models = append(models, controlmodels.SiteFromDB(site))
	}
	return models, nil
}

func (s *SitesStore) SearchHTMXResponse(models []*controlmodels.Site, query *apiutils.Query) templ.Component {
	return control_sites_templ.SitesTable(models, query.Skip, query.Limit)
}

func (s *SitesStore) GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controlmodels.Site, *apiutils.HTTPError) {
	site, err := controlContext.ControlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: id})
	if err == sql.ErrNoRows || (err == nil && site == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching site", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	return controlmodels.SiteFromDB(site), nil
}

// GetHTMXResponse is the site's details, shown in a modal
func (s *SitesStore) GetHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteModal(model)
}

// UpdateModel can only disable a live site, which stops it being served until it is archived. Only admins can.
func (s *SitesStore) UpdateModel(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site, model *controlmodels.Site) *apiutils.HTTPError {
	if controlContext.IdentityModel.Role != consts.ControlRoleAdmin {
		return apiutils.ErrForbidden
	}
	if model.State != "disabled" {
		return apiutils.ErrBadBodyWithHint("a site's state can only be changed to disabled")
	}
	result, err := controlContext.ControlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: oldModel.ID})
	if err != nil {
		slog.Error("Error disabling site", "error", err)
		return apiutils.ErrInternalServerError
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return apiutils.ErrConstraintCheckFailedWithHint("only a live site can be disabled")
	}
	updated, herr := s.GetModel(ctx, oldModel.ID, controlContext)
	if herr != nil || updated == nil {
		return apiutils.ErrInternalServerError
	}
	*model = *updated
	return nil
}

func (s *SitesStore) UpdateHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteDisabled(model)
}
