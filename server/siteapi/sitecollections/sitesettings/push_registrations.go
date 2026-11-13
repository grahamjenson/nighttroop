package sitesettings

import (
	"context"
	"errors"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// PushRegistrationsCollection is the signed in identity's own apps on devices that pushes are sent to: an app
// registers its token after its identity signs in. A token is encrypted and never returned; registering one again,
// from this or another identity, moves it there. An admin can also get and delete any of the site's, see PushIdentity.
type PushRegistrationsCollection struct{}

// PushIdentity is whose push registrations the signed in identity may get, test and delete: an admin's site's, and
// anyone else's own
func PushIdentity(siteContext *httputils.SiteContext) *int64 {
	if siteContext.IdentityModel.Role == consts.IdentityRoleAdmin {
		return nil
	}
	return &siteContext.IdentityModel.ID
}

// ////
// Shared
// ////
func (s *PushRegistrationsCollection) GetName() string {
	return "push_registrations"
}

// GetRoute is where the collection is in its URL
func (s *PushRegistrationsCollection) GetRoute() string {
	return "push-registrations"
}

// ToModel is the registration without its token, named for the signed in identity when it is theirs. Its row has the
// identity it is for, so its changes are streamed to that identity alone.
func (s *PushRegistrationsCollection) ToModel(row *controldb.PushRegistration, siteContext *httputils.SiteContext) *sitemodels.PushRegistration {
	registration := sitemodels.PushRegistrationFromDB(row)
	if row.IdentityID == siteContext.IdentityModel.ID {
		registration.IdentityName = siteContext.IdentityModel.Name
	}
	return registration
}

// ////
// CREATE
// ////

// CreateRow registers the app's token for the signed in identity, of the site the request is for
func (s *PushRegistrationsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.PushRegistration) (*controldb.PushRegistration, error) {
	if model.Token == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("errors.missing_field.token")
	}
	registration := push.Registration{Provider: model.Provider, App: string(model.App), BundleID: model.BundleID,
		Token: *model.Token, Production: model.Production}
	if err := push.CheckRegistration(registration); err != nil {
		return nil, apiutils.ErrBadBodyWithHint(err)
	}
	// Loop and Trio take remote commands from their registration, so one that cannot change the site's data registers neither
	if (model.App == consts.PushAppLoop || model.App == consts.PushAppTrio) && !can.SeeDeviceTokens(siteContext.IdentityModel) {
		return nil, apiutils.ErrForbidden
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	registered, err := push.Register(ctx, q, siteContext.Config.Security, siteContext.SiteID, siteContext.IdentityModel.ID, registration)
	if errors.Is(err, push.ErrTooManyRegistrations) {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(err)
	}
	if err != nil {
		return nil, err
	}
	return registered, nil
}

func (s *PushRegistrationsCollection) CreateHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return site_settings_templ.PushDeviceRow(model)
}

// ////
// Get Model
// /////

// GetRow is a registration the signed in identity may see, see PushIdentity
func (s *PushRegistrationsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*controldb.PushRegistration, error) {
	registrationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	return q.GetPushRegistration(ctx, controldb.GetPushRegistrationParams{ID: registrationID, SiteID: siteContext.SiteID,
		IdentityID: PushIdentity(siteContext)})
}

func (s *PushRegistrationsCollection) GetHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return site_settings_templ.PushDeviceRow(model)
}

// ///
// Search Models
// ///

// SearchRows is the signed in identity's registrations, newest first
func (s *PushRegistrationsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*controldb.PushRegistration, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	return q.SearchPushRegistrations(ctx, controldb.SearchPushRegistrationsParams{SiteID: siteContext.SiteID,
		IdentityID: &siteContext.IdentityModel.ID, Skip: query.Skip, Limit: query.Limit})
}

// SearchHTMXResponse is Push Devices with the identity's own registrations, as anyone but an admin sees it
func (s *PushRegistrationsCollection) SearchHTMXResponse(models []*sitemodels.PushRegistration, query *apiutils.Query) templ.Component {
	return site_settings_templ.PushDevices(models, false)
}

// ///
// Delete
// ///

// DeleteRow stops pushes to the app, as when it signs out or is deleted on Push Devices, see PushIdentity
func (s *PushRegistrationsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	registrationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return apiutils.ErrNotFound
	}
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}
	_, err = q.DeletePushRegistration(ctx, controldb.DeletePushRegistrationParams{ID: registrationID, SiteID: siteContext.SiteID,
		IdentityID: PushIdentity(siteContext)})
	return err
}

func (s *PushRegistrationsCollection) DeleteHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return site_settings_templ.PushDeviceDeleted()
}
