package controlcollections

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/i18n"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/models/publicmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/utils"
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

// UpdateRow disables a live, inactive site, which stops it being served, or archives a disabled one
func (s *SitesCollection) UpdateRow(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site, model *controlmodels.Site) (*controldb.Site, error) {
	var err error
	switch model.State {
	case "disabled":
		err = s.disable(ctx, controlContext, oldModel)
	case "archived":
		err = s.archive(ctx, controlContext, oldModel)
	default:
		err = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.site_state"))
	}
	if err != nil {
		return nil, err
	}
	return s.GetRow(ctx, oldModel.ID, controlContext)
}

func (s *SitesCollection) disable(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site) error {
	days := controlContext.ControlDB.ServerConfig.DisableAfterInactivityDays
	if !oldModel.Inactive(days, time.Now()) {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.disable_unused_only", i18n.Data{"Days": days}))
	}
	result, err := controlContext.ControlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: oldModel.ID})
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.disable_live_only"))
	}
	controlContext.ControlDB.Logs.SiteDisabledEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name)
	return nil
}

// archive uploads a disabled site's DB to its archive in S3 and deletes it here, so without S3 it is refused
func (s *SitesCollection) archive(ctx context.Context, controlContext *httputils.ControlContext, oldModel *controlmodels.Site) error {
	config := controlContext.ControlDB.ServerConfig
	if oldModel.State != "disabled" {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.archive_disabled_only"))
	}
	if !config.Status.S3Working {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.archive_needs_s3"))
	}
	started := time.Now()
	if err := controlContext.ControlDB.ArchiveSite(ctx, oldModel.ID, s3utils.NewS3Client(config).UploadFilesToS3); err != nil {
		return err
	}
	controlContext.ControlDB.Logs.SiteArchivedEvent(ctx, oldModel.ID, controlContext.IdentityModel.Name, started)
	return nil
}

func (s *SitesCollection) UpdateHTMXResponse(model *controlmodels.Site) templ.Component {
	return control_sites_templ.SiteUpdated(model)
}

// CreateHandler creates a site at the site ID an admin picks, such as a vanity one for a special user, and shows its
// credentials once. Only admins can, and the ID must be 8 lowercase letters, like every site's, and not taken. MAX_SITES
// and turning off public site creation do not stop them.
func (s *SitesCollection) CreateHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	model, herr := apiutils.ExtractModelFromRequest[controlmodels.Site](httputils.GetContentType(r), r)
	controlDB := controlContext.ControlDB
	if herr == nil && !utils.SyntacticallyValidateSiteID(model.ID) {
		herr = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.site_id", i18n.Data{"Length": utils.SITE_ID_LENGTH}))
	}
	var site *db.SiteCreateResult
	if herr == nil {
		var err error
		if site, err = controlDB.CreateSiteWithID(r.Context(), model.ID); errors.Is(err, db.ErrSiteIDTaken) {
			herr = &apiutils.HTTPError{Status: http.StatusConflict, Message: err.Error()}
		} else if err != nil {
			herr = apiutils.ErrInternalServerError
		}
	}
	if herr != nil && acceptType == httputils.AcceptTypeHTMX {
		apiutils.WriteStatus(w, acceptType, herr.Status)
		control_sites_templ.SiteNotCreated(herr.Message).Render(r.Context(), w)
		return
	}
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	info := publicmodels.NewSiteInfo{URL: site.SiteDB.URL(), AdminPassword: site.AdminPassword, AdminTotpSecret: site.TotpSecret,
		PwdNSToken: site.PwdNSToken, PwdNSURL: site.SiteDB.NSURL(), ReadOnlyNSToken: site.ReadOnlyNSToken, BackupKey: site.BackupPassword,
		RecoverURL: site.SiteDB.ServerConfig.RecoverURL()}
	WriteResponse(w, r, acceptType, http.StatusCreated, info, control_sites_templ.SiteCreated(info))
}

// SettingsHandler changes the server's settings: whether the public can create sites. Only admins can.
func (s *SitesCollection) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	model, herr := apiutils.ExtractModelFromRequest[controlmodels.ServerSettings](httputils.GetContentType(r), r)
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	err := controlContext.ControlDB.Queries.UpdatePublicSiteCreation(r.Context(), controldb.UpdatePublicSiteCreationParams{
		PublicSiteCreation: model.PublicSiteCreation, UpdatedByIdentityName: &controlContext.IdentityModel.Name})
	if err != nil {
		slog.ErrorContext(r.Context(), "Unable to update the server settings", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	WriteResponse(w, r, acceptType, http.StatusOK, model, control_sites_templ.PublicSiteCreationChanged(model.PublicSiteCreation))
}

// DownloadDBHandler sends a control admin the site's download (backup.SiteDownload), the same zip of a snapshot of its
// DB its own admins get from Download Data, to look into what it holds. It holds the site's data, so only an admin can,
// it shares the site's few downloads a minute, and it is recorded as an event the site's own admins see. A site whose
// DB is not here, as one archived, has none to download.
func (s *SitesCollection) DownloadDBHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	siteID := chi.URLParam(r, "id")
	siteDB, _, err := controlContext.ControlDB.GetSite(r.Context(), siteID)
	if err != nil {
		apiutils.ErrNotFound.Write(acceptType, w)
		return
	}
	if allowed, wait := httputils.WithinRateLimitAll(controlContext.ControlDB.ServerConfig.RateLimiters.SiteDownloadLimiter(siteID)); !allowed {
		apiutils.ErrRateLimitExceeded(wait).Write(acceptType, w)
		return
	}
	// The same download as the site's admin gets from its Download Data
	download, err := backup.SnapshotSite(r.Context(), siteDB.SQLiteDB, siteID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to copy a site's DB for a control admin", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	defer download.Close()
	controlContext.ControlDB.Logs.SiteDBDownloadedByControlEvent(r.Context(), siteID, controlContext.IdentityModel.Name)
	if err := download.Send(w); err != nil {
		slog.ErrorContext(r.Context(), "failed to send a site's DB to a control admin", "error", err)
	}
}

// TruncateHandler truncates all of a site's device reports, as the job every 10 minutes does its newest, then shrinks
// its DB, while its DB is here. Only admins can.
func (s *SitesCollection) TruncateHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	site, herr := get(r.Context(), s, chi.URLParam(r, "id"), controlContext)
	if herr == nil && site == nil {
		herr = apiutils.ErrNotFound
	}
	if herr == nil && site.State != "live" && site.State != "disabled" {
		herr = apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.truncate"))
	}
	if herr == nil {
		started := time.Now()
		sizeKB, err := controlContext.ControlDB.TruncateSite(r.Context(), site.ID)
		if herr = apiutils.ErrorFor(err, s.GetName()); herr == nil {
			controlContext.ControlDB.Logs.SiteTruncatedEvent(r.Context(), site.ID, controlContext.IdentityModel.Name, started, site.EstimatedDbSizeKb, sizeKB)
			site, herr = get(r.Context(), s, site.ID, controlContext)
		}
	}
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	WriteResponse(w, r, acceptType, http.StatusOK, site, control_sites_templ.SiteTruncated(site))
}
