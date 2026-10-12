package settings

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/models/site"

	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/site/siteutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type SyncSettingsStore struct{}

func (s *SyncSettingsStore) CreateModel(ctx context.Context, siteContext *httputils.SiteContext, model *site.SyncSource) (bool, *siteutils.HTTPError) {
	params := site.SyncSourceToCreate(model)
	params.SiteID = siteContext.Subdomain

	params.UrlOrUsername = strings.TrimSpace(params.UrlOrUsername)
	if params.UrlOrUsername == "" {
		return false, siteutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}

	security := siteContext.Config.Security
	var err error

	// Nightscout specific validation and normalization
	if params.Type == consts.SyncSourceTypeNightscout {
		params.UrlOrUsername, err = nsURLNormalizeValidate(params.UrlOrUsername, siteContext.Config)
		if err != nil {
			slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", params.UrlOrUsername)
			return false, siteutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", params.UrlOrUsername, err))
		}
	}

	if params.Password == "" {
		return false, siteutils.ErrMissingRequiredFieldWithHint("password")
	}

	params.Password, err = security.Encrypt(params.Password)
	if err != nil {
		return false, siteutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout password: %s", err))
	}

	params.UrlOrUsername, err = security.Encrypt(params.UrlOrUsername)
	if err != nil {
		return false, siteutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout URL: %s", err))
	}

	err = siteContext.ControlDB.Queries.CreateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return false, siteutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return false, siteutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return false, siteutils.ErrConstraintCheckFailed
		}
		return false, siteutils.ErrBadBody
	}
	return true, nil
}

func (s *SyncSettingsStore) CreateHTMXResponse(model *site.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

func (s *SyncSettingsStore) GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*site.SyncSource, *siteutils.HTTPError) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, siteutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid id: %s", id))
	}

	model, err := siteContext.ControlDB.Queries.GetSyncSource(ctx, control.GetSyncSourceParams{
		ID:     int64(idInt),
		SiteID: siteContext.Subdomain,
	})

	if err != nil {
		return nil, siteutils.ErrBadBody
	}
	if model == nil {
		return nil, nil
	}

	return site.SyncSourceFromDB(model), nil
}

func (s *SyncSettingsStore) UpdateModel(ctx context.Context, id string, siteContext *httputils.SiteContext, oldModel *site.SyncSource, model *site.SyncSource) *siteutils.HTTPError {
	if allowed, seconds := httputils.UserUpdateSyncWithinRateLimit(siteContext); !allowed {
		return siteutils.ErrRateLimitExceeded(seconds)
	}

	params := site.SyncSourceToUpdate(model)
	params.SiteID = siteContext.Subdomain

	var err error

	if params.UrlOrUsername != nil && strings.TrimSpace(*params.UrlOrUsername) == "" {
		return siteutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}

	security := siteContext.Config.Security

	// Nightscout specific validation and normalization
	if oldModel.Type == consts.SyncSourceTypeNightscout {
		if params.UrlOrUsername != nil {
			*params.UrlOrUsername, err = nsURLNormalizeValidate(*params.UrlOrUsername, siteContext.Config)
			if err != nil {
				slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", params.UrlOrUsername)
				return siteutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", params.UrlOrUsername, err))
			}
		}
	}

	if params.Password != nil {
		if strings.TrimSpace(*params.Password) == "" || *params.Password == consts.PASSWORD_GARBAGE_VALUE {
			params.Password = nil
		} else {
			*params.Password, err = security.Encrypt(*params.Password)
			if err != nil {
				return siteutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Password: %s", err))
			}

		}
	}

	if params.UrlOrUsername != nil {
		*params.UrlOrUsername, err = security.Encrypt(*params.UrlOrUsername)
		if err != nil {
			return siteutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt URL Or Username: %s", err))
		}
	}

	err = siteContext.ControlDB.Queries.UpdateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return siteutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return siteutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return siteutils.ErrConstraintCheckFailed
		}
		return siteutils.ErrBadBody
	}

	return nil
}

func (s *SyncSettingsStore) UpdateHTMXResponse(model *site.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

func (s *SyncSettingsStore) GetName() string {
	return "sync_settings"
}

func GetSyncSettingsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

		ctx := r.Context()

		if acceptType.API() {
			// TODO return JSON of current settings, devices, tokens, last seen, etc
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "This will return the current sync settings in JSON format."}`))
			return
		}

		dbSyncSources, err := controlDB.Queries.GetAllSyncSourcesForSite(ctx, control.GetAllSyncSourcesForSiteParams{
			SiteID: siteContext.Subdomain,
		})

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		modelSyncSources := []*site.SyncSource{}
		for _, dbSyncSource := range dbSyncSources {
			modelSyncSource := site.SyncSourceFromDB(dbSyncSource)
			modelSyncSource.Password = consts.PASSWORD_GARBAGE_VALUE
			modelSyncSource.UrlOrUsername, err = config.Security.Decrypt(modelSyncSource.UrlOrUsername)
			if err != nil {
				slog.Error("Failed to decrypt sync source URL or username from DB", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			modelSyncSources = append(modelSyncSources, modelSyncSource)

		}

		if acceptType == httputils.AcceptTypeHTML {
			err = site_settings_templ.SyncSettingsFull(modelSyncSources).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTML template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = site_settings_templ.SyncSettingsPartial(modelSyncSources).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTMX template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

func DELETESyncSettingsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetSyncSettingsReportsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

///////
// UTIL METHODS
////////

func nsURLNormalizeValidate(rawUrl string, config *config.ServerConfig) (string, error) {
	errMsg := ""
	nsURL, err := url.Parse(rawUrl)
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
		errMsg += "Nightscout URL must use HTTPS."
	}
	if path != "" && path != "/" {
		errMsg += "Nightscout URL cannot have a path. "
	}
	if params != "" {
		errMsg += "Nightscout URL cannot have query parameters. "
	}

	err = httputils.IsNightscoutURLBlocked(host, config)
	if err != nil {
		errMsg += "Nightscout URL is blocked." + err.Error()
	}

	if errMsg != "" {
		return rawUrl, fmt.Errorf("%s", errMsg)
	}

	// Normalize the URL by removing default Ports

	normalizedUrl := fmt.Sprintf("%s://%s", scheme, host)
	// if the port is specified and is not the default port
	if port != "" {
		normalizedUrl = fmt.Sprintf("%s://%s:%s", scheme, host, port)
	}

	return normalizedUrl, nil
}
