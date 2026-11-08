package sitesettings

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/models/sitemodels"

	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

type SyncSettingsCollection struct{}

// ////
// Shared
// ////
func (s *SyncSettingsCollection) GetName() string {
	return "sync_settings"
}

// GetRoute is where the collection is in its URL
func (s *SyncSettingsCollection) GetRoute() string {
	return "sync"
}

// Upserts makes an update of a type of sync source the site has not got yet create it
func (s *SyncSettingsCollection) Upserts() {}

// ////
// CREATE
// ////
func (s *SyncSettingsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SyncSource) (*sitemodels.SyncSource, error) {
	var err error
	params := sitemodels.SyncSourceToCreate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.SiteID
	if id == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("id/type")
	}

	sstype, err := consts.NewSyncSourceTypeFromString(*id)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", *id))
	}
	params.Type = sstype

	// The virtual source has no upstream; the sync job ignores these values.
	if params.Type == consts.SyncSourceTypeVirtual {
		params.UrlOrUsernameEnc, params.PasswordEnc = new("virtual"), new("virtual")
	}

	if params.UrlOrUsernameEnc == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	if params.UrlOrUsernameEnc, err = cleanURLOrUsername(ctx, siteContext, params.Type, params.UrlOrUsernameEnc); err != nil {
		return nil, err
	}

	if params.Type == consts.SyncSourceTypeLibre {
		if params.Region == nil || *params.Region == "" {
			params.Region = new(consts.LibreRegionGlobal)
		}
		if err := libreRegionValidate(params.Region); err != nil {
			return nil, err
		}
	}

	if params.PasswordEnc == nil || *params.PasswordEnc == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("password")
	}

	if err := encryptCredentials(siteContext.Config.Security, params.UrlOrUsernameEnc, params.PasswordEnc); err != nil {
		return nil, err
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	if err := q.CreateSyncSource(ctx, params); err != nil {
		return nil, syncSourceWriteError(err)
	}

	return model, nil
}

func (s *SyncSettingsCollection) CreateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ///
// Update
// ///

// UpdateRow changes the fields sent of the site's sync source, keeping its type
func (s *SyncSettingsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SyncSource, model *sitemodels.SyncSource) (*sitemodels.SyncSource, error) {
	if allowed, seconds := httputils.UserUpdateSyncWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	params := sitemodels.SyncSourceToUpdate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.SiteID
	params.Type = oldModel.Type

	var err error

	if params.UrlOrUsernameEnc != nil {
		if params.UrlOrUsernameEnc, err = cleanURLOrUsername(ctx, siteContext, params.Type, params.UrlOrUsernameEnc); err != nil {
			return nil, err
		}
	}

	if params.Type == consts.SyncSourceTypeLibre && params.Region != nil {
		if err := libreRegionValidate(params.Region); err != nil {
			return nil, err
		}
	}

	// A blank password, or the placeholder the form shows for the saved one, keeps the saved one
	if params.PasswordEnc != nil && (strings.TrimSpace(*params.PasswordEnc) == "" || *params.PasswordEnc == consts.PASSWORD_GARBAGE_VALUE) {
		params.PasswordEnc = nil
	}

	if err := encryptCredentials(siteContext.Config.Security, params.UrlOrUsernameEnc, params.PasswordEnc); err != nil {
		return nil, err
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	if err := q.UpdateSyncSource(ctx, params); err != nil {
		return nil, syncSourceWriteError(err)
	}

	return model, nil
}

func (s *SyncSettingsCollection) UpdateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ////
// Get Model
// /////
func (s *SyncSettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SyncSource, error) {
	sstype, err := consts.NewSyncSourceTypeFromString(id)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", id))
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	model, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{
		Type:   sstype,
		SiteID: siteContext.SiteID,
	})

	// None yet is not an error: saving a sync source the site has not got creates it
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching sync source", "error", err)
		return nil, apiutils.ErrBadBody
	}
	if model == nil {
		return nil, nil
	}

	respMode, err := sitemodels.SyncSourceFromDB(model, siteContext.Config.Security)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to convert sync source from DB: %s", err))
	}
	return respMode, nil
}

// ToModel is the sync source: it is decrypted as it is read, so its row is its model
func (s *SyncSettingsCollection) ToModel(row *sitemodels.SyncSource, _ *httputils.SiteContext) *sitemodels.SyncSource {
	return row
}

// ///
// Search Models
// ///

func (s *SyncSettingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SyncSource, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	dbSyncSources, err := q.GetAllSyncSourcesForSite(ctx, controldb.GetAllSyncSourcesForSiteParams{
		SiteID: siteContext.SiteID,
	})

	if err != nil {
		slog.ErrorContext(ctx, "Search Sync sources", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*sitemodels.SyncSource{}
	for _, dbSyncSource := range dbSyncSources {
		model, err := sitemodels.SyncSourceFromDB(dbSyncSource, siteContext.Config.Security)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to convert sync source from DB", "error", err)
			continue
		}
		if err := addRecentLogs(ctx, siteContext, model); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// recentSyncLogs is how many of a sync source's newest events and errors its form shows
const recentSyncLogs = 5

// addRecentLogs adds the source's newest events and errors from the logs DB, the site's own only
func addRecentLogs(ctx context.Context, siteContext *httputils.SiteContext, model *sitemodels.SyncSource) error {
	logs, syncType := siteContext.ControlDB.Logs.Queries, string(model.Type)
	events, err := logs.RecentSyncEvents(ctx, logsdb.RecentSyncEventsParams{SiteID: &siteContext.SiteID, SyncType: &syncType, Limit: recentSyncLogs})
	if err != nil {
		return err
	}
	syncErrors, err := logs.RecentSyncErrors(ctx, logsdb.RecentSyncErrorsParams{SiteID: &siteContext.SiteID, SyncType: &syncType, Limit: recentSyncLogs})
	if err != nil {
		return err
	}
	for _, event := range events {
		model.RecentEvents = append(model.RecentEvents, sitemodels.SiteEventFromDB(event))
	}
	for _, syncError := range syncErrors {
		model.RecentErrors = append(model.RecentErrors, sitemodels.SiteErrorFromDB(syncError))
	}
	return nil
}

func (s *SyncSettingsCollection) SearchHTMXResponse(models []*sitemodels.SyncSource, query *apiutils.Query) templ.Component {
	return site_settings_templ.SyncSettingsPartial(models)
}

// ///
// Delete
// ///

// DeleteRow removes the site's sync source of the type in the id
func (s *SyncSettingsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	sstype, err := consts.NewSyncSourceTypeFromString(id)
	if err != nil {
		return apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", id))
	}
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}

	return q.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{Type: sstype, SiteID: siteContext.SiteID})
}

func (s *SyncSettingsCollection) DeleteHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SyncDeleted(model.Type)
}

///////
// UTIL METHODS
////////

// cleanURLOrUsername is the URL or username trimmed, as pasted spaces would make a URL invalid, with a Nightscout URL
// normalised, or why it cannot be used
func cleanURLOrUsername(ctx context.Context, siteContext *httputils.SiteContext, sstype consts.SyncSourceType, urlOrUsername *string) (*string, error) {
	trimmed := strings.TrimSpace(*urlOrUsername)
	if trimmed == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	if sstype != consts.SyncSourceTypeNightscout {
		return &trimmed, nil
	}
	normalized, err := nsURLNormalizeValidate(&trimmed, siteContext.Config)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to normalize and validate Nightscout URL", "error", err, "url", trimmed)
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", trimmed, err))
	}
	return normalized, nil
}

// encryptCredentials encrypts in place each of the sync source's URL or username and password that is set
func encryptCredentials(security *security.SecurityConfig, credentials ...*string) error {
	for _, credential := range credentials {
		if credential == nil {
			continue
		}
		encrypted, err := security.Encrypt(*credential)
		if err != nil {
			return apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt sync source credentials: %s", err))
		}
		*credential = encrypted
	}
	return nil
}

// syncSourceWriteError is the error for the control DB refusing to create or update a sync source
func syncSourceWriteError(err error) error {
	switch {
	case errors.IsSQLUniqueViolationError(err):
		return apiutils.ErrUniqueConstraintViolation
	case errors.IsSQLConstraintNotNull(err):
		return apiutils.ErrMissingRequiredField
	case errors.IsSQLConstraintCheck(err):
		return apiutils.ErrDatabaseRefused(err)
	}
	return apiutils.ErrBadBody
}

func nsURLNormalizeValidate(rawUrl *string, config *config.ServerConfig) (*string, error) {
	if rawUrl == nil {
		return nil, fmt.Errorf("URL is nil")
	}
	// Each problem is a sentence, joined with spaces
	problems := []string{}
	nsURL, err := url.Parse(*rawUrl)
	if err != nil {
		return rawUrl, fmt.Errorf("Invalid URL: %w", err)
	}

	scheme := strings.ToLower(nsURL.Scheme)
	host := strings.ToLower(nsURL.Hostname())
	path := nsURL.Path
	params := nsURL.RawQuery
	port := nsURL.Port()

	if (port == "80" && scheme == "http") || (port == "443" && scheme == "https") {
		port = "" // Remove default ports
	}

	if scheme != "https" && !(config.IsDevOrTest() && scheme == "http") {
		problems = append(problems, "Nightscout URL must use HTTPS.")
	}
	if path != "" && path != "/" {
		problems = append(problems, "Nightscout URL cannot have a path.")
	}
	if params != "" {
		problems = append(problems, "Nightscout URL cannot have query parameters.")
	}

	err = httputils.IsNightscoutURLBlocked(host, config)
	if err != nil {
		problems = append(problems, "Nightscout URL is blocked: "+err.Error())
	}

	if len(problems) > 0 {
		return rawUrl, fmt.Errorf("%s", strings.Join(problems, " "))
	}

	// Normalize the URL by removing default Ports

	normalizedUrl := fmt.Sprintf("%s://%s", scheme, host)
	// if the port is specified and is not the default port
	if port != "" {
		normalizedUrl = fmt.Sprintf("%s://%s:%s", scheme, host, port)
	}

	return new(normalizedUrl), nil
}

// libreRegionValidate refuses a region not in consts.LibreRegions, so a Libre sync only calls their hosts
func libreRegionValidate(region *string) error {
	if !consts.IsLibreRegion(*region) {
		return apiutils.ErrBadBodyWithHint(fmt.Sprintf("LibreLinkUp region '%s' is not a known region.", *region))
	}
	return nil
}
