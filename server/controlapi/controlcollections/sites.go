package controlcollections

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/s3utils"
)

// SitesCollection lets control identities see every site, whatever its state, and admins disable and archive them
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

// UpdateRow disables a live, inactive site, which stops it being served, or archives a disabled one
func (s *SitesCollection) UpdateRow(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site, model *controlmodels.Site) (*controldb.Site, error) {
	var err error
	switch model.State {
	case "disabled":
		err = s.disable(ctx, controlContext, oldModel)
	case "archived":
		err = s.archive(ctx, controlContext, oldModel)
	default:
		err = apiutils.ErrBadBodyWithHint("a site's state can only be changed to disabled or archived")
	}
	if err != nil {
		return nil, err
	}
	return s.GetRow(ctx, oldModel.ID, controlContext)
}

func (s *SitesCollection) disable(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site) error {
	days := controlContext.ControlDB.ServerConfig.DisableAfterInactivityDays
	if !oldModel.Inactive(days, time.Now()) {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("only a site not written to or seen in %d days can be disabled", days))
	}
	result, err := controlContext.ControlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: oldModel.ID})
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return apiutils.ErrConstraintCheckFailedWithHint("only a live site can be disabled")
	}
	controlContext.ControlDB.Logs.SiteDisabledEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name)
	return nil
}

// archive uploads a disabled site's DB to its archive in S3 and deletes it here, so without S3 it is refused
func (s *SitesCollection) archive(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site) error {
	config := controlContext.ControlDB.ServerConfig
	if oldModel.State != "disabled" {
		return apiutils.ErrConstraintCheckFailedWithHint("only a disabled site can be archived")
	}
	if !config.Status.S3Working {
		return apiutils.ErrConstraintCheckFailedWithHint("a site can only be archived when S3 is connected")
	}
	if err := controlContext.ControlDB.ArchiveSite(ctx, oldModel.ID, s3utils.NewS3Client(config).UploadFilesToS3); err != nil {
		return err
	}
	controlContext.ControlDB.Logs.SiteArchivedEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name)
	return nil
}

func (s *SitesCollection) UpdateHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteUpdated(model)
}
