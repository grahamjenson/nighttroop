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

type ConnectorSettingsCollection struct{}

// ////
// Shared
// ////
func (s *ConnectorSettingsCollection) GetName() string {
	return "connector_settings"
}

// GetRoute is where the collection is in its URL
func (s *ConnectorSettingsCollection) GetRoute() string {
	return "connectors"
}

// Upserts makes an update of a type of connector the site has not got yet create it
func (s *ConnectorSettingsCollection) Upserts() {}

// ////
// CREATE
// ////
func (s *ConnectorSettingsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.Connector) (*sitemodels.Connector, error) {
	// Creating spends the same limit as updating, so a site cannot save new sources to sync from any faster
	if allowed, seconds := httputils.UserUpdateConnectorsWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	var err error
	params := sitemodels.ConnectorToCreate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.SiteID
	if id == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("id/type")
	}

	sstype, err := consts.NewConnectorTypeFromString(*id)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid connector type: %s", *id))
	}
	params.Type = sstype
	if err := checkVirtualConnectorAllowed(siteContext, params.Type); err != nil {
		return nil, err
	}

	// The virtual source has no upstream; the connector job ignores these values.
	if params.Type == consts.ConnectorTypeVirtual {
		params.UrlOrUsernameEnc, params.PasswordEnc = new("virtual"), new("virtual")
	}

	if params.UrlOrUsernameEnc == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	if params.UrlOrUsernameEnc, err = cleanURLOrUsername(ctx, siteContext, params.Type, params.UrlOrUsernameEnc); err != nil {
		return nil, err
	}

	if params.Type == consts.ConnectorTypeLibre {
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

	if err := q.CreateConnector(ctx, params); err != nil {
		return nil, connectorWriteError(err)
	}

	model.Type = params.Type // so its response shows the buttons of that connector
	return model, nil
}

func (s *ConnectorSettingsCollection) CreateHTMXResponse(model *sitemodels.Connector) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ///
// Update
// ///

// UpdateRow changes the fields sent of the site's connector, keeping its type
func (s *ConnectorSettingsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.Connector, model *sitemodels.Connector) (*sitemodels.Connector, error) {
	if allowed, seconds := httputils.UserUpdateConnectorsWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	params := sitemodels.ConnectorToUpdate(model)
	// DO NOT ALLOW THE CLIENT TO SET SITEID OR TYPE
	params.SiteID = siteContext.SiteID
	params.Type = oldModel.Type
	if err := checkVirtualConnectorAllowed(siteContext, params.Type); err != nil {
		return nil, err
	}

	var err error

	if params.UrlOrUsernameEnc != nil {
		if params.UrlOrUsernameEnc, err = cleanURLOrUsername(ctx, siteContext, params.Type, params.UrlOrUsernameEnc); err != nil {
			return nil, err
		}
	}

	if params.Type == consts.ConnectorTypeLibre && params.Region != nil {
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

	if err := q.UpdateConnector(ctx, params); err != nil {
		return nil, connectorWriteError(err)
	}

	model.Type = params.Type // so its response shows the buttons of that connector
	return model, nil
}

func (s *ConnectorSettingsCollection) UpdateHTMXResponse(model *sitemodels.Connector) templ.Component {
	return site_settings_templ.SuccessfulUpdate(model)
}

// ////
// Get Model
// /////
func (s *ConnectorSettingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.Connector, error) {
	sstype, err := consts.NewConnectorTypeFromString(id)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid connector type: %s", id))
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	model, err := q.GetConnector(ctx, controldb.GetConnectorParams{
		Type:   sstype,
		SiteID: siteContext.SiteID,
	})

	// None yet is not an error: saving a connector the site has not got creates it
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching connector", "error", err)
		return nil, apiutils.ErrBadBody
	}
	if model == nil {
		return nil, nil
	}

	respMode, err := sitemodels.ConnectorFromDB(model, siteContext.Config.Security)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to convert connector from DB: %s", err))
	}
	return respMode, nil
}

// ToModel is the connector: it is decrypted as it is read, so its row is its model
func (s *ConnectorSettingsCollection) ToModel(row *sitemodels.Connector, _ *httputils.SiteContext) *sitemodels.Connector {
	return row
}

// ///
// Search Models
// ///

func (s *ConnectorSettingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.Connector, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	dbConnectors, err := q.GetAllConnectorsForSite(ctx, controldb.GetAllConnectorsForSiteParams{
		SiteID: siteContext.SiteID,
	})

	if err != nil {
		slog.ErrorContext(ctx, "Search Connectors", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*sitemodels.Connector{}
	for _, dbConnector := range dbConnectors {
		model, err := sitemodels.ConnectorFromDB(dbConnector, siteContext.Config.Security)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to convert connector from DB", "error", err)
			continue
		}
		if err := addRecentLogs(ctx, siteContext, model); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// recentConnectorLogs is how many of a connector's newest events and errors its form shows
const recentConnectorLogs = 5

// addRecentLogs adds the source's newest events and errors from the logs DB, the site's own only
func addRecentLogs(ctx context.Context, siteContext *httputils.SiteContext, model *sitemodels.Connector) error {
	logs, connectorType := siteContext.ControlDB.Logs.Queries, string(model.Type)
	events, err := logs.RecentConnectorEvents(ctx, logsdb.RecentConnectorEventsParams{SiteID: &siteContext.SiteID, ConnectorType: &connectorType, Limit: recentConnectorLogs})
	if err != nil {
		return err
	}
	syncErrors, err := logs.RecentConnectorErrors(ctx, logsdb.RecentConnectorErrorsParams{SiteID: &siteContext.SiteID, ConnectorType: &connectorType, Limit: recentConnectorLogs})
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

func (s *ConnectorSettingsCollection) SearchHTMXResponse(models []*sitemodels.Connector, query *apiutils.Query) templ.Component {
	return site_settings_templ.ConnectorSettingsPartial(models)
}

// ///
// Delete
// ///

// DeleteRow removes the site's connector of the type in the id
func (s *ConnectorSettingsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	sstype, err := consts.NewConnectorTypeFromString(id)
	if err != nil {
		return apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid connector type: %s", id))
	}
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}

	return q.DeleteConnector(ctx, controldb.DeleteConnectorParams{Type: sstype, SiteID: siteContext.SiteID})
}

func (s *ConnectorSettingsCollection) DeleteHTMXResponse(model *sitemodels.Connector) templ.Component {
	return site_settings_templ.ConnectorDeleted(model.Type)
}

///////
// UTIL METHODS
////////

// cleanURLOrUsername is the URL or username trimmed, as pasted spaces would make a URL invalid, with a Nightscout URL
// normalised, or why it cannot be used
func cleanURLOrUsername(ctx context.Context, siteContext *httputils.SiteContext, sstype consts.ConnectorType, urlOrUsername *string) (*string, error) {
	trimmed := strings.TrimSpace(*urlOrUsername)
	if trimmed == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("Url Or Username")
	}
	if sstype != consts.ConnectorTypeNightscout {
		return &trimmed, nil
	}
	normalized, err := nsURLNormalizeValidate(&trimmed, siteContext.Config)
	if err != nil {
		// Neither the log nor the error, which is recorded, has the URL, which can hold a login or token
		slog.ErrorContext(ctx, "Failed to normalize and validate Nightscout URL", "error", err)
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Nightscout URL is not a valid URL. %s", err))
	}
	return normalized, nil
}

// encryptCredentials encrypts in place each of the connector's URL or username and password that is set
func encryptCredentials(security *security.SecurityConfig, credentials ...*string) error {
	for _, credential := range credentials {
		if credential == nil {
			continue
		}
		encrypted, err := security.Encrypt(*credential)
		if err != nil {
			return apiutils.ErrBadBodyWithHint(fmt.Sprintf("Failed to encrypt connector credentials: %s", err))
		}
		*credential = encrypted
	}
	return nil
}

// checkVirtualConnectorAllowed refuses saving the virtual source, which makes up glucose, on a production server
func checkVirtualConnectorAllowed(siteContext *httputils.SiteContext, sstype consts.ConnectorType) error {
	if sstype == consts.ConnectorTypeVirtual && !siteContext.Config.VirtualConnectorAllowed() {
		return apiutils.ErrConstraintCheckFailedWithHint("virtual connector, which makes up glucose, is only on development, test and staging servers")
	}
	return nil
}

// connectorWriteError is the error for the control DB refusing to create or update a connector
func connectorWriteError(err error) error {
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
		return rawUrl, fmt.Errorf("Nightscout URL cannot be read as a URL.")
	}

	scheme := strings.ToLower(nsURL.Scheme)
	host := httputils.NormalizeHost(nsURL.Hostname())
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

// libreRegionValidate refuses a region not in consts.LibreRegions, so a Libre connector only calls their hosts
func libreRegionValidate(region *string) error {
	if !consts.IsLibreRegion(*region) {
		return apiutils.ErrBadBodyWithHint(fmt.Sprintf("LibreLinkUp region '%s' is not a known region.", *region))
	}
	return nil
}
