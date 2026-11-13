package sitesettings

import (
	"context"
	"database/sql"
	"log/slog"
	"slices"
	"strconv"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/i18n"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

type IdentitiesCollection struct{}

var identityRoles = []consts.IdentityRole{
	consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly, consts.IdentityRoleClinician,
	consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite,
}

// mgdlClients are the Nightscout clients a device can be added for that need glucose in mg/dL, so a device added for one
// has its own units set to mg/dL whatever the site's. Any other device, "other", follows the site.
var mgdlClients = []string{"aaps", "trio", "loop", "xdrip", "xdrip4ios", "nightguard", "sugarpixel", "t1dcarer"}

// clientMgdlLabels are how a client needs the Nightscout API to write mg/dL, where it is not mg/dL: xDrip4iOS reads a
// BG check's glucose as mg/dL only when its units are exactly mg/dl, and otherwise uploads it again, and AAPS writes
// them mg/dl, the only spelling of mg/dL its NsUnits reads exactly. They spell mmol/L as Nightscout does too, mmol, as
// AAPS reads a profile in mmol/L as mg/dL otherwise
var clientMgdlLabels = map[string]string{"xdrip4ios": "mg/dl", "aaps": "mg/dl"}

// ////
// Shared
// ////
func (s *IdentitiesCollection) GetName() string {
	return "identities"
}

// GetRoute is where the collection is in its URL
func (s *IdentitiesCollection) GetRoute() string {
	return "identities"
}

// ////
// CREATE
// ////
func (s *IdentitiesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.UserIdentity) (*sitemodels.UserIdentity, error) {
	// A device for a client that must have a role has it, and cannot be made with another
	if forced, ok := sitemodels.ClientRoles[model.Client]; ok {
		if model.Role != "" && consts.IdentityRole(model.Role) != forced {
			return nil, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.identities.device_role", i18n.Data{"Client": model.Client, "Role": forced}))
		}
		model.Role = string(forced)
	}
	if model.Name == "" || model.Role == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("errors.missing_field.name_and_role")
	}
	role := consts.IdentityRole(model.Role)
	if !slices.Contains(identityRoles, role) {
		return nil, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.identities.unknown_role", i18n.Data{"Role": model.Role}))
	}
	if model.Client != "" && model.Client != "other" && !slices.Contains(mgdlClients, model.Client) {
		return nil, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.identities.unknown_client", i18n.Data{"Client": model.Client}))
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	// Counted inside the transaction, so two creates at once cannot both pass the limit
	count, err := q.CountSiteIdentities(ctx, controldb.CountSiteIdentitiesParams{SiteID: siteContext.SiteID})
	if err != nil {
		return nil, err
	}
	if maxIdentities := siteContext.Config.MaxIdentitiesPerSite; count >= int64(maxIdentities) {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.identities.too_many", i18n.Data{"Max": maxIdentities}))
	}
	// Each create hashes a password with bcrypt, so stop creating and deleting from using up the CPU
	if allowed, seconds := httputils.IdentityCreateWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	// The app a device is for is kept, so its settings can later be changed for every device of an app
	var client *string
	if model.Client != "" {
		client = &model.Client
	}
	values, err := siteContext.ControlDB.CreateSiteIdentity(ctx, q, siteContext.SiteID, model.Name, role, client)
	if err != nil {
		// Names are unique per site
		return nil, err
	}

	created, herr := getIdentityByName(ctx, q, siteContext, model.Name)
	if herr != nil {
		return nil, herr
	}
	if slices.Contains(mgdlClients, model.Client) {
		if err := setUnitsMGDL(ctx, q, siteContext, created.ID, model.Client); err != nil {
			return nil, err
		}
	}
	// Shown this once
	created.Password, created.TotpSecret, created.NSToken = values.Password, values.TotpSecret, values.NSToken
	created.SiteURL, created.NSURL = siteContext.SiteDB.URL(), siteContext.SiteDB.NSURL()
	return created, nil
}

// setUnitsMGDL sets the identity's own units to mg/dL, written as client needs, for a client that needs them, keeping
// its other settings
func setUnitsMGDL(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, identityID int64, client string) error {
	identity, err := q.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
	if err != nil {
		return err
	}
	level := settings.NewSettingsMap(identity, &identity.CreatedAt, new("identity"))
	if err := level.SetSettingByName("units", string(consts.UnitsMGDL), nil, new("identity")); err != nil {
		return err
	}
	if label, ok := clientMgdlLabels[client]; ok {
		if err := level.SetSettingByName("units_mgdl_label", label, nil, new("identity")); err != nil {
			return err
		}
		if err := level.SetSettingByName("units_mmol_label", "mmol", nil, new("identity")); err != nil {
			return err
		}
	}
	params := &controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: siteContext.SiteID}
	settings.SetSettings(params, level)
	return q.UpdateSiteIdentitySettings(ctx, *params)
}

func (s *IdentitiesCollection) CreateHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.CreateIdentityResponse(model.SiteURL, model, model.Password, model.TotpSecret, model.NSToken)
}

// ///
// Update
// ///

// UpdateRow only turns an identity on or off
func (s *IdentitiesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.UserIdentity, model *sitemodels.UserIdentity) (*sitemodels.UserIdentity, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	if !model.Enabled {
		if herr := checkCanRemoveIdentity(ctx, q, siteContext, oldModel); herr != nil {
			return nil, herr
		}
	}

	// Respond with the whole identity, not just the fields that were sent
	enabled := model.Enabled
	*model = *oldModel
	model.Enabled = enabled

	return model, q.UpdateSiteIdentityEnabled(ctx, sitemodels.IdentityToUpdate(model, siteContext))
}

func (s *IdentitiesCollection) UpdateHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityUpdated(model)
}

// ////
// Get Model
// /////
func (s *IdentitiesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.UserIdentity, error) {
	identityID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.identities.bad_id", i18n.Data{"ID": id}))
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	identity, err := q.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
	if err != nil {
		return nil, err
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

// ToModel is the identity: it is converted as it is read, so its row is its model
func (s *IdentitiesCollection) ToModel(row *sitemodels.UserIdentity, _ *httputils.SiteContext) *sitemodels.UserIdentity {
	return row
}

func (s *IdentitiesCollection) GetHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityRow(model)
}

// ///
// Search Models
// ///

func (s *IdentitiesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.UserIdentity, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.SiteID})
	if err != nil {
		return nil, err
	}
	models := make([]*sitemodels.UserIdentity, 0, len(identities))
	for _, identity := range identities {
		models = append(models, sitemodels.IdentityFromDB(identity, siteContext))
	}
	return models, nil
}

func (s *IdentitiesCollection) SearchHTMXResponse(models []*sitemodels.UserIdentity, query *apiutils.Query) templ.Component {
	users, devices := []*sitemodels.UserIdentity{}, []*sitemodels.UserIdentity{}
	for _, model := range models {
		switch consts.IdentityRole(model.Role).GetType() {
		case consts.IdentityTypeAdmin:
			// Admins are listed first
			users = append([]*sitemodels.UserIdentity{model}, users...)
		case consts.IdentityTypeUser:
			users = append(users, model)
		case consts.IdentityTypeDevice:
			devices = append(devices, model)
		}
	}
	return site_settings_templ.IdentitesTable(users, devices)
}

// ///
// Delete
// ///

// DeleteRow removes the identity, so its password and tokens stop working
func (s *IdentitiesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {

	model, err := s.GetRow(ctx, collectionTx, id, siteContext)
	if err != nil {
		return err
	}
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}

	if herr := checkCanRemoveIdentity(ctx, q, siteContext, model); herr != nil {
		return herr
	}
	return q.DeleteSiteIdentity(ctx, sitemodels.IdentityToDelete(model, siteContext))
}

func (s *IdentitiesCollection) DeleteHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityDeleted()
}

///////
// UTIL METHODS
////////

func getIdentityByName(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, name string) (*sitemodels.UserIdentity, *apiutils.HTTPError) {
	identity, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteContext.SiteID, Name: name})
	if err == sql.ErrNoRows || (err == nil && identity == nil) {
		return nil, apiutils.ErrNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching identity", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

// checkCanRemoveIdentity stops an identity disabling or deleting itself, or the site losing its last enabled admin
func checkCanRemoveIdentity(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, identity *sitemodels.UserIdentity) *apiutils.HTTPError {
	if identity.CurrentIdentity {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.identities.not_yourself"))
	}
	if consts.IdentityRole(identity.Role) != consts.IdentityRoleAdmin || !identity.Enabled {
		return nil
	}

	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.SiteID})
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	admins := 0
	for _, i := range identities {
		if i.Role == consts.IdentityRoleAdmin && i.Enabled {
			admins++
		}
	}
	if admins <= 1 {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.identities.keep_an_admin"))
	}
	return nil
}
