package sitesettings

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type IdentitiesCollection struct{}

var identityRoles = []consts.IdentityRole{
	consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly,
	consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite,
}

// ////
// Shared
// ////
func (s *IdentitiesCollection) GetName() string {
	return "identities"
}

// ////
// CREATE
// ////
func (s *IdentitiesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.UserIdentity) (*sitemodels.UserIdentity, error) {
	if model.Name == "" || model.Role == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("name and role")
	}
	role := consts.IdentityRole(model.Role)
	if !slices.Contains(identityRoles, role) {
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Unknown role: %s", model.Role))
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}

	// Counted inside the transaction, so two creates at once cannot both pass the limit
	count, err := q.CountSiteIdentities(ctx, controldb.CountSiteIdentitiesParams{SiteID: siteContext.Subdomain})
	if err != nil {
		return nil, err
	}
	if maxIdentities := siteContext.Config.MaxIdentitiesPerSite; count >= int64(maxIdentities) {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("a site can have at most %d users and devices", maxIdentities))
	}
	// Each create hashes a password with bcrypt, so stop creating and deleting from using up the CPU
	if allowed, seconds := httputils.IdentityCreateWithinRateLimit(siteContext); !allowed {
		return nil, apiutils.ErrRateLimitExceeded(seconds)
	}

	values, err := siteContext.ControlDB.CreateSiteIdentity(ctx, q, siteContext.Subdomain, model.Name, role)
	if err != nil {
		// Names are unique per site
		return nil, err
	}

	created, herr := getIdentityByName(ctx, q, siteContext, model.Name)
	if herr != nil {
		return nil, herr
	}
	// Shown this once
	created.Password, created.TotpSecret, created.NSToken = values.Password, values.TotpSecret, values.NSToken
	created.SiteURL = siteContext.SiteDB.URL()
	return created, nil
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
		return nil, apiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid identity id: %s", id))
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	identity, err := q.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.Subdomain, ID: identityID})
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
	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.Subdomain})
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
	identity, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteContext.Subdomain, Name: name})
	if err == sql.ErrNoRows || (err == nil && identity == nil) {
		return nil, apiutils.ErrNotFound
	}
	if err != nil {
		slog.Error("Error fetching identity", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

// checkCanRemoveIdentity stops an identity disabling or deleting itself, or the site losing its last enabled admin
func checkCanRemoveIdentity(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, identity *sitemodels.UserIdentity) *apiutils.HTTPError {
	if identity.CurrentIdentity {
		return apiutils.ErrConstraintCheckFailedWithHint("you cannot disable or delete yourself")
	}
	if consts.IdentityRole(identity.Role) != consts.IdentityRoleAdmin || !identity.Enabled {
		return nil
	}

	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.Subdomain})
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
		return apiutils.ErrConstraintCheckFailedWithHint("a site must keep at least one enabled admin")
	}
	return nil
}
