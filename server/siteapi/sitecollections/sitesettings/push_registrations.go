package sitesettings

import (
	"context"
	"errors"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// PushRegistrationsCollection is the signed in identity's own apps on devices that pushes are sent to: an app
// registers its token after its identity signs in. A token is encrypted and never returned; registering one again,
// from this or another identity, moves it there.
type PushRegistrationsCollection struct{}

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

// ToModel is the registration without its token. Its row has the identity it is for, so its changes are streamed to
// that identity alone.
func (s *PushRegistrationsCollection) ToModel(row *controldb.PushRegistration, _ *httputils.SiteContext) *sitemodels.PushRegistration {
	return sitemodels.PushRegistrationFromDB(row)
}

// ////
// CREATE
// ////

// CreateRow registers the app's token for the signed in identity, of the site the request is for
func (s *PushRegistrationsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.PushRegistration) (*controldb.PushRegistration, error) {
	if model.Token == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("token")
	}
	registration := push.Registration{Provider: model.Provider, App: string(model.App), BundleID: model.BundleID,
		Token: *model.Token, Production: model.Production}
	if err := push.CheckRegistration(registration); err != nil {
		return nil, apiutils.ErrBadBodyWithHint(err.Error())
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	registered, err := push.Register(ctx, q, siteContext.Config.Security, siteContext.SiteID, siteContext.IdentityModel.ID, registration)
	if errors.Is(err, push.ErrTooManyRegistrations) {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(err.Error())
	}
	if err != nil {
		return nil, err
	}
	return registered, nil
}

func (s *PushRegistrationsCollection) CreateHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return templ.NopComponent
}

// ////
// Get Model
// /////
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
		IdentityID: siteContext.IdentityModel.ID})
}

func (s *PushRegistrationsCollection) GetHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return templ.NopComponent
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

func (s *PushRegistrationsCollection) SearchHTMXResponse(models []*sitemodels.PushRegistration, query *apiutils.Query) templ.Component {
	return templ.NopComponent
}

// ///
// Delete
// ///

// DeleteRow stops pushes to the app, as when it signs out
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
		IdentityID: &siteContext.IdentityModel.ID})
	return err
}

func (s *PushRegistrationsCollection) DeleteHTMXResponse(model *sitemodels.PushRegistration) templ.Component {
	return templ.NopComponent
}
