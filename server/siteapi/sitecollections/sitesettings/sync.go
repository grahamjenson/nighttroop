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
	"nighttroop.t1d.nz/models/sitemodels"

	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type SyncSettingsCollection struct{}

// ////
// Shared
// ////
func (s *SyncSettingsCollection) GetName() string {
	return "sync_settings"
}

// ////
// CREATE
// ////
func (s *SyncSettingsCollection) CreateRow(ctx context.Context, q *controldb.Queries, id *string, siteContext *httputils.SiteContext, model *sitemodels.SyncSource) (*sitemodels.SyncSource, error) {
	var err error
	params := sitemodels.SyncSourceToCreate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.Subdomain
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
		params.UrlOrUsernameEnc, params.PasswordEnc = ptr("virtual"), ptr("virtual")
	}

	if params.UrlOrUsernameEnc == nil || strings.TrimSpace(*params.UrlOrUsernameEnc) == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	params.UrlOrUsernameEnc = ptr(strings.TrimSpace(*params.UrlOrUsernameEnc))

	security := siteContext.Config.Security

	// Nightscout specific validation and normalization
	if params.Type == consts.SyncSourceTypeNightscout {
		normalized, err := nsURLNormalizeValidate(params.UrlOrUsernameEnc, siteContext.Config)
		if err != nil {
			slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", *params.UrlOrUsernameEnc)
			return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", *params.UrlOrUsernameEnc, err))
		}
		params.UrlOrUsernameEnc = normalized
	}

	if params.PasswordEnc == nil || *params.PasswordEnc == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("password")
	}

	encPassword, err := security.Encrypt(*params.PasswordEnc)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout password: %s", err))
	}
	params.PasswordEnc = &encPassword

	encUrlOrUsername, err := security.Encrypt(*params.UrlOrUsernameEnc)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout URL: %s", err))
	}
	params.UrlOrUsernameEnc = &encUrlOrUsername

	err = q.CreateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return nil, apiutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return nil, apiutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return nil, apiutils.ErrConstraintCheckFailedWithHint(err.Error())
		}
		return nil, apiutils.ErrBadBody
	}

	return model, nil
}

func (s *SyncSettingsCollection) CreateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ///
// Update
// ///

func (s *SyncSettingsCollection) UpdateRow(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SyncSource, model *sitemodels.SyncSource) (*sitemodels.SyncSource, error) {
	if allowed, seconds := httputils.UserUpdateSyncWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	params := sitemodels.SyncSourceToUpdate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.Subdomain
	params.Type = oldModel.Type

	var err error

	if params.UrlOrUsernameEnc != nil {
		if strings.TrimSpace(*params.UrlOrUsernameEnc) == "" {
			return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
		}
		// Trim like create does, so pasted spaces do not make the URL invalid
		params.UrlOrUsernameEnc = ptr(strings.TrimSpace(*params.UrlOrUsernameEnc))
	}

	security := siteContext.Config.Security

	// Nightscout specific validation and normalization
	if oldModel.Type == consts.SyncSourceTypeNightscout {
		if params.UrlOrUsernameEnc != nil {
			normalized, err := nsURLNormalizeValidate(params.UrlOrUsernameEnc, siteContext.Config)
			if err != nil {
				slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", *params.UrlOrUsernameEnc)
				return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", *params.UrlOrUsernameEnc, err))
			}
			params.UrlOrUsernameEnc = normalized
		}
	}

	if params.PasswordEnc != nil {
		if strings.TrimSpace(*params.PasswordEnc) == "" || *params.PasswordEnc == consts.PASSWORD_GARBAGE_VALUE {
			params.PasswordEnc = nil
		} else {
			*params.PasswordEnc, err = security.Encrypt(*params.PasswordEnc)
			if err != nil {
				return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Password: %s", err))
			}

		}
	}

	if params.UrlOrUsernameEnc != nil {
		*params.UrlOrUsernameEnc, err = security.Encrypt(*params.UrlOrUsernameEnc)
		if err != nil {
			return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt URL Or Username: %s", err))
		}
	}

	err = q.UpdateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return nil, apiutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return nil, apiutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return nil, apiutils.ErrConstraintCheckFailed
		}
		return nil, apiutils.ErrBadBody
	}

	return model, nil
}

func (s *SyncSettingsCollection) UpdateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ////
// Get Model
// /////
func (s *SyncSettingsCollection) GetRow(ctx context.Context, q *controldb.Queries, id string, siteContext *httputils.SiteContext) (*sitemodels.SyncSource, error) {
	sstype, err := consts.NewSyncSourceTypeFromString(id)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", id))
	}

	model, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{
		Type:   sstype,
		SiteID: siteContext.Subdomain,
	})

	if err != nil {
		slog.Error("Error fetching sync source", "error", err)
		if err == sql.ErrNoRows {
			return nil, nil
		}
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

func (s *SyncSettingsCollection) SearchRows(ctx context.Context, q *controldb.Queries, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SyncSource, error) {
	dbSyncSources, err := q.GetAllSyncSourcesForSite(ctx, controldb.GetAllSyncSourcesForSiteParams{
		SiteID: siteContext.Subdomain,
	})

	if err != nil {
		slog.Error("Search Sync sources", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*sitemodels.SyncSource{}
	for _, dbSyncSource := range dbSyncSources {
		model, err := sitemodels.SyncSourceFromDB(dbSyncSource, siteContext.Config.Security)
		if err != nil {
			slog.Error("Failed to convert sync source from DB", "error", err)
			continue
		}
		models = append(models, model)
	}
	return models, nil
}

func (s *SyncSettingsCollection) SearchHTMXResponse(models []*sitemodels.SyncSource, query *apiutils.Query) templ.Component {
	return site_settings_templ.SyncSettingsPartial(models)
}

// ///
// Delete
// ///

// DeleteRow removes the site's sync source of the type in the id
func (s *SyncSettingsCollection) DeleteRow(ctx context.Context, q *controldb.Queries, id string, siteContext *httputils.SiteContext) error {
	sstype, err := consts.NewSyncSourceTypeFromString(id)
	if err != nil {
		return apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", id))
	}
	return q.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{Type: sstype, SiteID: siteContext.Subdomain})
}

func (s *SyncSettingsCollection) DeleteHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SyncDeleted(model.Type)
}

///////
// UTIL METHODS
////////

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

	return ptr(normalizedUrl), nil
}

func ptr[T any](v T) *T {
	return &v
}
