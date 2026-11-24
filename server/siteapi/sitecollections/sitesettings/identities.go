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

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/can"
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
var mgdlClients = []string{"aaps", consts.AAPSClient, "trio", "loop", "xdrip", "xdrip4ios", "sugarpixel", "t1dcarer"}

// followClients are the Nightscout clients a device can be added for that show glucose in the units of their
// Nightscout status, so a device added for one follows the site: Nightguard reads sgv as mg/dL and shows it in
// settings.units
var followClients = []string{"nightguard"}

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
	if forced, ok := can.ClientRoles[model.Client]; ok {
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
	if model.Client != "" && model.Client != "other" && !slices.Contains(mgdlClients, model.Client) && !slices.Contains(followClients, model.Client) {
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
		if err := setUnitsMGDL(ctx, q, siteContext, created.ID); err != nil {
			return nil, err
		}
	}
	// Shown this once
	created.Password, created.TotpSecret, created.NSToken = values.Password, values.TotpSecret, values.NSToken
	created.SiteURL, created.NSURL = siteContext.SiteDB.URL(), siteContext.SiteDB.NSURL()
	return created, nil
}

// setUnitsMGDL sets the identity's own units to mg/dL, for a client that needs them, keeping its other settings
func setUnitsMGDL(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, identityID int64) error {
	level := &settings.SettingMap{}
	level.SetUnits(new(consts.UnitsMGDL))
	return db.SaveSiteSettings(ctx, siteContext.SiteDB.Queries(), q, siteContext.SiteID, identityID, level, siteContext.IdentityModel.Name)
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
	identity, err := siteapiutils.ControlQuery(ctx, collectionTx, (*controldb.Queries).GetSiteIdentityByID, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.SiteID, ID: identityID})
	if err != nil {
		return nil, err
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

// ToModel is the identity: it is converted as it is read, so its row is its model
func (s *IdentitiesCollection) ToModel(row *sitemodels.UserIdentity, _ *httputils.SiteContext) *sitemodels.UserIdentity {
	return row
}

// Hides a change with an identity's credentials, as a new identity's row has, from the identities' stream, which
// another admin or tab may have open: they are shown once, to whoever made the identity, so it is sent as {}
func (s *IdentitiesCollection) Hides(event events.Event, _ *httputils.SiteContext) bool {
	identity, ok := event.Model.(*sitemodels.UserIdentity)
	return ok && (identity.Password != "" || identity.TotpSecret != nil || identity.NSToken != nil)
}

func (s *IdentitiesCollection) GetHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityRow(model)
}

// ///
// Search Models
// ///

func (s *IdentitiesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.UserIdentity, error) {
	identities, err := siteapiutils.ControlQuery(ctx, collectionTx, (*controldb.Queries).AllSiteIdentities, controldb.AllSiteIdentitiesParams{SiteID: siteContext.SiteID})
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
	// The identity's settings go with it
	if err := db.DeleteIdentitySettings(ctx, siteContext.SiteDB.Queries(), model.ID); err != nil {
		return err
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
	if !can.IdentityAdministerSite(consts.IdentityRole(identity.Role)) || !identity.Enabled {
		return nil
	}

	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.SiteID})
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	admins := 0
	for _, i := range identities {
		if can.IdentityAdministerSite(i.Role) && i.Enabled {
			admins++
		}
	}
	if admins <= 1 {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.identities.keep_an_admin"))
	}
	return nil
}
