package sitesettings

import (
	"context"
	"fmt"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// maxPushCredentials is the most push credentials a site can have, far more than a family's own builds of apps need
const maxPushCredentials = 5

// PushCredentialsCollection is the site admin's credentials for pushes to apps they build themselves, such as Loop,
// which only the site's identities use. Its secret is encrypted and never returned; to change one, delete it and add
// it again.
type PushCredentialsCollection struct{}

// ////
// Shared
// ////
func (s *PushCredentialsCollection) GetName() string {
	return "push_credentials"
}

// GetRoute is where the collection is in its URL
func (s *PushCredentialsCollection) GetRoute() string {
	return "push-credentials"
}

// ToModel is the credential: it is converted as it is read, so its row is its model
func (s *PushCredentialsCollection) ToModel(row *sitemodels.PushCredential, _ *httputils.SiteContext) *sitemodels.PushCredential {
	return row
}

// ////
// CREATE
// ////

// CreateRow checks the credential and keeps it, its secret encrypted, for the site the request is for
func (s *PushCredentialsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.PushCredential) (*sitemodels.PushCredential, error) {
	if model.Name == "" || len(model.Name) > 64 || model.Secret == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("name (at most 64 characters) and secret")
	}
	if err := push.Check(push.Credential{Provider: model.Provider, TeamOrProjectID: model.TeamOrProjectID, KeyID: model.KeyID,
		Secret: *model.Secret, Apps: model.Apps, BundleIDs: model.BundleIDs}); err != nil {
		return nil, apiutils.ErrBadBodyWithHint(err.Error())
	}
	secretEnc, err := siteContext.Config.Security.Encrypt(*model.Secret)
	if err != nil {
		return nil, err
	}

	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	// Counted inside the transaction, so two creates at once cannot both pass the limit
	count, err := q.CountPushCredentials(ctx, controldb.CountPushCredentialsParams{SiteID: &siteContext.SiteID})
	if err != nil {
		return nil, err
	}
	if count >= maxPushCredentials {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("a site can have at most %d push credentials", maxPushCredentials))
	}
	params := sitemodels.PushCredentialToCreate(model, siteContext)
	params.SecretEnc = &secretEnc
	credential, err := q.CreatePushCredential(ctx, params)
	if err != nil {
		return nil, err
	}
	return sitemodels.PushCredentialFromDB(credential), nil
}

func (s *PushCredentialsCollection) CreateHTMXResponse(model *sitemodels.PushCredential) templ.Component {
	return templ.NopComponent
}

// ////
// Get Model
// /////
func (s *PushCredentialsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.PushCredential, error) {
	credentialID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	credential, err := q.GetPushCredential(ctx, controldb.GetPushCredentialParams{ID: credentialID, SiteID: &siteContext.SiteID})
	if err != nil {
		return nil, err
	}
	return sitemodels.PushCredentialFromDB(credential), nil
}

func (s *PushCredentialsCollection) GetHTMXResponse(model *sitemodels.PushCredential) templ.Component {
	return templ.NopComponent
}

// ///
// Search Models
// ///

// SearchRows is the site's credentials, newest first
func (s *PushCredentialsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.PushCredential, error) {
	q, err := collectionTx.Control()
	if err != nil {
		return nil, err
	}
	credentials, err := q.SearchPushCredentials(ctx, controldb.SearchPushCredentialsParams{SiteID: &siteContext.SiteID,
		Skip: query.Skip, Limit: query.Limit})
	if err != nil {
		return nil, err
	}
	models := make([]*sitemodels.PushCredential, 0, len(credentials))
	for _, credential := range credentials {
		models = append(models, sitemodels.PushCredentialFromDB(credential))
	}
	return models, nil
}

func (s *PushCredentialsCollection) SearchHTMXResponse(models []*sitemodels.PushCredential, query *apiutils.Query) templ.Component {
	return templ.NopComponent
}

// ///
// Delete
// ///

// DeleteRow soft deletes the credential, removing its secret, so pushes no longer use it
func (s *PushCredentialsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	credentialID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return apiutils.ErrNotFound
	}
	q, err := collectionTx.Control()
	if err != nil {
		return err
	}
	_, err = q.DeletePushCredential(ctx, controldb.DeletePushCredentialParams{ID: credentialID, SiteID: &siteContext.SiteID})
	return err
}

func (s *PushCredentialsCollection) DeleteHTMXResponse(model *sitemodels.PushCredential) templ.Component {
	return templ.NopComponent
}
