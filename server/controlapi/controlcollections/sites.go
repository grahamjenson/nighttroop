package controlcollections

import (
	"context"
	"net/http"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils/httputils"
)

// SitesCollection lets control identities see every site, whatever its state, and admins disable live ones
type SitesCollection struct{}

func (s *SitesCollection) GetName() string {
	return "sites"
}

// GetRoute is where the collection is in its URL, /sites
func (s *SitesCollection) GetRoute() string {
	return "sites"
}

func (s *SitesCollection) ToModel(row *controldb.Site) *controlmodels.Site {
	return controlmodels.SiteFromDB(row)
}

// SearchRows is the sites newest first, or by order last_write the longest since written to first, or by order
// db_size the largest site DB first
func (s *SitesCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controldb.Site, error) {
	params := controlmodels.QueryToSearchSites(query)
	switch orderOf(query) {
	case "":
		return controlContext.ControlDB.Queries.SearchSites(ctx, params)
	case "last_write":
		return controlContext.ControlDB.Queries.SearchSitesByLastWrite(ctx, controldb.SearchSitesByLastWriteParams(params))
	case "db_size":
		return controlContext.ControlDB.Queries.SearchSitesByDBSize(ctx, controldb.SearchSitesByDBSizeParams(params))
	}
	return nil, &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "sites can be ordered by last_write or db_size"}
}

func (s *SitesCollection) SearchHTMXResponse(models []*controlmodels.Site, query *apiutils.Query) templ.Component {
	return control_sites_templ.SitesTable(models, query.Skip, query.Limit, orderOf(query))
}

// orderOf is what a query orders by, or empty for the default
func orderOf(query *apiutils.Query) string {
	if query.Order == nil {
		return ""
	}
	return *query.Order
}

func (s *SitesCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controldb.Site, error) {
	return controlContext.ControlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: id})
}

// GetHTMXResponse is the site's details, shown in a modal
func (s *SitesCollection) GetHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteModal(model)
}

// UpdateRow can only disable a live site, which stops it being served until it is archived
func (s *SitesCollection) UpdateRow(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site, model *controlmodels.Site) (*controldb.Site, error) {
	if model.State != "disabled" {
		return nil, apiutils.ErrBadBodyWithHint("a site's state can only be changed to disabled")
	}
	result, err := controlContext.ControlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: oldModel.ID})
	if err != nil {
		return nil, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, apiutils.ErrConstraintCheckFailedWithHint("only a live site can be disabled")
	}
	controlContext.ControlDB.Logs.SiteDisabledEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name)
	return s.GetRow(ctx, oldModel.ID, controlContext)
}

func (s *SitesCollection) UpdateHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteDisabled(model)
}
