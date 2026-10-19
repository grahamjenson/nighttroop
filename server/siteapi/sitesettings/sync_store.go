package sitesettings

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	ntsync "nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"

	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/nightscoutapi"

	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type SyncSettingsStore struct{}

// ////
// Shared
// ////
func (s *SyncSettingsStore) GetName() string {
	return "sync_settings"
}

// ////
// CREATE
// ////
func (s *SyncSettingsStore) CreateModel(ctx context.Context, id *string, siteContext *httputils.SiteContext, model *sitemodels.SyncSource) (bool, *siteapiutils.HTTPError) {
	var err error
	params := sitemodels.SyncSourceToCreate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.Subdomain
	if id == nil {
		return false, siteapiutils.ErrMissingRequiredFieldWithHint("id/type")
	}

	sstype, err := consts.NewSyncSourceTypeFromString(*id)
	if err != nil {
		return false, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", *id))
	}
	params.Type = sstype

	// The virtual source has no upstream; the sync job ignores these values.
	if params.Type == consts.SyncSourceTypeVirtual {
		params.UrlOrUsernameEnc, params.PasswordEnc = ptr("virtual"), ptr("virtual")
	}

	if params.UrlOrUsernameEnc == nil || strings.TrimSpace(*params.UrlOrUsernameEnc) == "" {
		return false, siteapiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	params.UrlOrUsernameEnc = ptr(strings.TrimSpace(*params.UrlOrUsernameEnc))

	security := siteContext.Config.Security

	// Nightscout specific validation and normalization
	if params.Type == consts.SyncSourceTypeNightscout {
		normalized, err := nsURLNormalizeValidate(params.UrlOrUsernameEnc, siteContext.Config)
		if err != nil {
			slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", *params.UrlOrUsernameEnc)
			return false, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", *params.UrlOrUsernameEnc, err))
		}
		params.UrlOrUsernameEnc = normalized
	}

	if params.PasswordEnc == nil || *params.PasswordEnc == "" {
		return false, siteapiutils.ErrMissingRequiredFieldWithHint("password")
	}

	if model.Enabled {
		if herr := verifySyncSource(ctx, params.Type, *params.UrlOrUsernameEnc, *params.PasswordEnc, params.Region); herr != nil {
			return false, herr
		}
	}

	encPassword, err := security.Encrypt(*params.PasswordEnc)
	if err != nil {
		return false, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout password: %s", err))
	}
	params.PasswordEnc = &encPassword

	encUrlOrUsername, err := security.Encrypt(*params.UrlOrUsernameEnc)
	if err != nil {
		return false, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Nightscout URL: %s", err))
	}
	params.UrlOrUsernameEnc = &encUrlOrUsername

	err = siteContext.ControlDB.Queries.CreateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return false, siteapiutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return false, siteapiutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return false, siteapiutils.ErrConstraintCheckFailedWithHint(err.Error())
		}
		return false, siteapiutils.ErrBadBody
	}

	if sstype == consts.SyncSourceTypeVirtual && model.Enabled {
		if err := seedVirtual(ctx, siteContext); err != nil {
			return false, err
		}
	}
	return true, nil
}

// seedVirtual writes the last 30 minutes of virtual data as soon as the
// virtual source is enabled, so the site has something to show right away.
func seedVirtual(ctx context.Context, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	store := nightscoutapi.NewNightTroopAPIStore(siteContext.ControlDB, siteContext.Config)
	added, err := ntsync.WriteVirtualData(ctx, store, siteContext, time.Now(), 30*time.Minute)
	if err != nil {
		slog.Error("Failed to seed virtual sync data", "error", err, "site_id", siteContext.Subdomain)
		return siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to create virtual data: %s", err))
	}
	slog.Info("Seeded virtual sync data", "site_id", siteContext.Subdomain, "added", added)
	return nil
}

func (s *SyncSettingsStore) CreateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ///
// Update
// ///

func (s *SyncSettingsStore) UpdateModel(ctx context.Context, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SyncSource, model *sitemodels.SyncSource) *siteapiutils.HTTPError {
	if allowed, seconds := httputils.UserUpdateSyncWithinRateLimit(siteContext); !allowed {
		return siteapiutils.ErrRateLimitExceeded(seconds)
	}

	params := sitemodels.SyncSourceToUpdate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.Subdomain
	params.Type = oldModel.Type

	var err error

	if params.UrlOrUsernameEnc != nil && strings.TrimSpace(*params.UrlOrUsernameEnc) == "" {
		return siteapiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}

	security := siteContext.Config.Security

	// Nightscout specific validation and normalization
	if oldModel.Type == consts.SyncSourceTypeNightscout {
		if params.UrlOrUsernameEnc != nil {
			normalized, err := nsURLNormalizeValidate(params.UrlOrUsernameEnc, siteContext.Config)
			if err != nil {
				slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", *params.UrlOrUsernameEnc)
				return siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", *params.UrlOrUsernameEnc, err))
			}
			params.UrlOrUsernameEnc = normalized
		}
	}

	if model.Enabled {
		if herr := verifyUpdatedSyncSource(ctx, siteContext, oldModel.Type, params); herr != nil {
			return herr
		}
	}

	if params.PasswordEnc != nil {
		if strings.TrimSpace(*params.PasswordEnc) == "" || *params.PasswordEnc == consts.PASSWORD_GARBAGE_VALUE {
			params.PasswordEnc = nil
		} else {
			*params.PasswordEnc, err = security.Encrypt(*params.PasswordEnc)
			if err != nil {
				return siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt Password: %s", err))
			}

		}
	}

	if params.UrlOrUsernameEnc != nil {
		*params.UrlOrUsernameEnc, err = security.Encrypt(*params.UrlOrUsernameEnc)
		if err != nil {
			return siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt URL Or Username: %s", err))
		}
	}

	err = siteContext.ControlDB.Queries.UpdateSyncSource(ctx, params)
	if err != nil {
		if errors.IsSQLUniqueViolationError(err) {
			return siteapiutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintNotNull(err) {
			return siteapiutils.ErrMissingRequiredField
		}
		if errors.IsSQLConstraintCheck(err) {
			return siteapiutils.ErrConstraintCheckFailed
		}
		return siteapiutils.ErrBadBody
	}

	if oldModel.Type == consts.SyncSourceTypeVirtual && model.Enabled {
		if err := seedVirtual(ctx, siteContext); err != nil {
			return err
		}
	}

	return nil
}

func (s *SyncSettingsStore) UpdateHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ////
// Get Model
// /////
func (s *SyncSettingsStore) GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*sitemodels.SyncSource, *siteapiutils.HTTPError) {
	sstype, err := consts.NewSyncSourceTypeFromString(id)
	if err != nil {
		return nil, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid sync source type: %s", id))
	}

	model, err := siteContext.ControlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{
		Type:   sstype,
		SiteID: siteContext.Subdomain,
	})

	if err != nil {
		slog.Error("Error fetching sync source", "error", err)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, siteapiutils.ErrBadBody
	}
	if model == nil {
		return nil, nil
	}

	respMode, err := sitemodels.SyncSourceFromDB(model, siteContext.Config.Security)
	if err != nil {
		return nil, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to convert sync source from DB: %s", err))
	}
	return respMode, nil
}

// ///
// Search Models
// ///

func (s *SyncSettingsStore) SearchModels(ctx context.Context, query *siteapiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SyncSource, *siteapiutils.HTTPError) {
	dbSyncSources, err := siteContext.ControlDB.Queries.GetAllSyncSourcesForSite(ctx, controldb.GetAllSyncSourcesForSiteParams{
		SiteID: siteContext.Subdomain,
	})

	if err != nil {
		slog.Error("Search Sync sources", "error", err)
		return nil, siteapiutils.ErrInternalServerError
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

func (s *SyncSettingsStore) SearchHTMXResponse(models []*sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SyncSettingsPartial(models)
}

// ///
// Delete
// ///

func (s *SyncSettingsStore) DeleteModel(ctx context.Context, model *sitemodels.SyncSource, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	err := siteContext.ControlDB.Queries.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{
		Type:   model.Type,
		SiteID: model.SiteID,
	})
	if err != nil {
		slog.Error("Error deleting sync source", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *SyncSettingsStore) DeleteHTMXResponse(model *sitemodels.SyncSource) templ.Component {
	return site_settings_templ.SuccessfulDelete()
}

///////
// UTIL METHODS
////////

// verifySyncSource connects to the source with these details and fails if
// they do not work, so broken credentials are never saved.
func verifySyncSource(ctx context.Context, sstype consts.SyncSourceType, urlOrUsername, password string, region *string) *siteapiutils.HTTPError {
	var err error
	switch sstype {
	case consts.SyncSourceTypeNightscout:
		err = nightscout_client.NewClient(urlOrUsername, password).Verify()
	case consts.SyncSourceTypeDexcom:
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		outsideUS := region == nil || consts.NewDexcomRegionFromString(*region) == consts.DexcomRegionOther
		err = dexcom_client.NewClient(urlOrUsername, password, outsideUS).Login(ctx)
	}
	if err != nil {
		return siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Could not connect to %s with these details: %s", sstype, err))
	}
	return nil
}

// verifyUpdatedSyncSource verifies an update, taking any detail it leaves out
// (the form sends a placeholder for an unchanged password) from the stored source.
func verifyUpdatedSyncSource(ctx context.Context, siteContext *httputils.SiteContext, sstype consts.SyncSourceType, params controldb.UpdateSyncSourceParams) *siteapiutils.HTTPError {
	password := params.PasswordEnc
	if password != nil && (strings.TrimSpace(*password) == "" || *password == consts.PASSWORD_GARBAGE_VALUE) {
		password = nil
	}
	if sstype == consts.SyncSourceTypeVirtual || (params.UrlOrUsernameEnc == nil && password == nil && params.Region == nil) {
		return nil
	}
	stored, err := siteContext.ControlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{Type: sstype, SiteID: siteContext.Subdomain})
	if err != nil {
		return siteapiutils.ErrBadBody
	}
	security := siteContext.Config.Security
	if params.UrlOrUsernameEnc == nil {
		params.UrlOrUsernameEnc, err = sitemodels.Decrypt(stored.UrlOrUsernameEnc, security)
	}
	if err == nil && password == nil {
		password, err = sitemodels.Decrypt(stored.PasswordEnc, security)
	}
	if err != nil || params.UrlOrUsernameEnc == nil || password == nil {
		return siteapiutils.ErrMissingRequiredFieldWithHint("Url Or Username and password")
	}
	if params.Region == nil {
		params.Region = stored.Region
	}
	return verifySyncSource(ctx, sstype, *params.UrlOrUsernameEnc, *password, params.Region)
}

func nsURLNormalizeValidate(rawUrl *string, config *config.ServerConfig) (*string, error) {
	if rawUrl == nil {
		return nil, fmt.Errorf("URL is nil")
	}
	errMsg := ""
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

	return ptr(normalizedUrl), nil
}

func ptr[T any](v T) *T {
	return &v
}
