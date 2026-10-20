package sitesettings

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type IdentitiesStore struct{}

var identityRoles = []consts.IdentityRole{
	consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly,
	consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite,
}

// ////
// Shared
// ////
func (s *IdentitiesStore) GetName() string {
	return "identities"
}

// ////
// CREATE
// ////
func (s *IdentitiesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.UserIdentity) (bool, *siteapiutils.HTTPError) {
	if model.Name == "" || model.Role == "" {
		return false, siteapiutils.ErrMissingRequiredFieldWithHint("name and role")
	}
	role := consts.IdentityRole(model.Role)
	if !slices.Contains(identityRoles, role) {
		return false, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Unknown role: %s", model.Role))
	}

	q, err := tx.Control()
	if err != nil {
		return false, siteapiutils.ErrInternalServerError
	}
	// Counted inside the transaction, so two creates at once cannot both pass the limit
	count, err := q.CountSiteIdentities(ctx, controldb.CountSiteIdentitiesParams{SiteID: siteContext.Subdomain})
	if err != nil {
		return false, siteapiutils.ErrInternalServerError
	}
	if maxIdentities := siteContext.Config.MaxIdentitiesPerSite; count >= int64(maxIdentities) {
		return false, siteapiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("a site can have at most %d users and devices", maxIdentities))
	}

	values, err := siteContext.ControlDB.CreateSiteIdentity(ctx, q, siteContext.Subdomain, model.Name, role)
	if err != nil {
		slog.Error("Failed to create identity", "error", err)
		// Names are unique per site
		if errors.IsSQLUniqueViolationError(err) {
			return false, siteapiutils.ErrUniqueConstraintViolation
		}
		if errors.IsSQLConstraintCheck(err) {
			return false, siteapiutils.ErrConstraintCheckFailedWithHint(err.Error())
		}
		return false, siteapiutils.ErrInternalServerError
	}

	created, herr := getIdentityByName(ctx, q, siteContext, model.Name)
	if herr != nil {
		return false, herr
	}
	*model = *created
	model.Password, model.TotpSecret, model.NSToken = values.Password, values.TotpSecret, values.NSToken
	model.SiteURL = siteContext.SiteDB.URL()
	return true, nil
}

func (s *IdentitiesStore) CreateHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.CreateIdentityResponse(model.SiteURL, model, model.Password, model.TotpSecret, model.NSToken)
}

// ///
// Update
// ///

// UpdateModel only turns an identity on or off
func (s *IdentitiesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.UserIdentity, model *sitemodels.UserIdentity) *siteapiutils.HTTPError {
	q, err := tx.Control()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	if !model.Enabled {
		if herr := checkCanRemoveIdentity(ctx, q, siteContext, oldModel); herr != nil {
			return herr
		}
	}

	// Respond with the whole identity, not just the fields that were sent
	enabled := model.Enabled
	*model = *oldModel
	model.Enabled = enabled

	err = q.UpdateSiteIdentityEnabled(ctx, sitemodels.IdentityToUpdate(model, siteContext))
	if err != nil {
		slog.Error("Error updating identity", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *IdentitiesStore) UpdateHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityUpdated(model)
}

// ////
// Get Model
// /////
func (s *IdentitiesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.UserIdentity, *siteapiutils.HTTPError) {
	identityID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, siteapiutils.ErrBadBodyWithHint(fmt.Sprintf("Invalid identity id: %s", id))
	}

	q, err := tx.Control()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	identity, err := q.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteContext.Subdomain, ID: identityID})
	if err == sql.ErrNoRows || (err == nil && identity == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching identity", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

func (s *IdentitiesStore) GetHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityRow(model)
}

// ///
// Search Models
// ///

func (s *IdentitiesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *siteapiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.UserIdentity, *siteapiutils.HTTPError) {
	q, err := tx.Control()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.Subdomain})
	if err != nil {
		slog.Error("Search identities", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	models := []*sitemodels.UserIdentity{}
	for _, identity := range identities {
		models = append(models, sitemodels.IdentityFromDB(identity, siteContext))
	}
	return models, nil
}

func (s *IdentitiesStore) SearchHTMXResponse(models []*sitemodels.UserIdentity, query *siteapiutils.Query) templ.Component {
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

// DeleteModel removes the identity, so its password and tokens stop working
func (s *IdentitiesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.UserIdentity, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	q, err := tx.Control()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	if herr := checkCanRemoveIdentity(ctx, q, siteContext, model); herr != nil {
		return herr
	}

	err = q.DeleteSiteIdentity(ctx, sitemodels.IdentityToDelete(model, siteContext))
	if err != nil {
		slog.Error("Error deleting identity", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *IdentitiesStore) DeleteHTMXResponse(model *sitemodels.UserIdentity) templ.Component {
	return site_settings_templ.IdentityDeleted()
}

///////
// UTIL METHODS
////////

func getIdentityByName(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, name string) (*sitemodels.UserIdentity, *siteapiutils.HTTPError) {
	identity, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteContext.Subdomain, Name: name})
	if err == sql.ErrNoRows || (err == nil && identity == nil) {
		return nil, siteapiutils.ErrNotFound
	}
	if err != nil {
		slog.Error("Error fetching identity", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	return sitemodels.IdentityFromDB(identity, siteContext), nil
}

// checkCanRemoveIdentity stops an identity disabling or deleting itself, or the site losing its last enabled admin
func checkCanRemoveIdentity(ctx context.Context, q *controldb.Queries, siteContext *httputils.SiteContext, identity *sitemodels.UserIdentity) *siteapiutils.HTTPError {
	if identity.CurrentIdentity {
		return siteapiutils.ErrConstraintCheckFailedWithHint("you cannot disable or delete yourself")
	}
	if consts.IdentityRole(identity.Role) != consts.IdentityRoleAdmin || !identity.Enabled {
		return nil
	}

	identities, err := q.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteContext.Subdomain})
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	admins := 0
	for _, i := range identities {
		if i.Role == consts.IdentityRoleAdmin && i.Enabled {
			admins++
		}
	}
	if admins <= 1 {
		return siteapiutils.ErrConstraintCheckFailedWithHint("a site must keep at least one enabled admin")
	}
	return nil
}
