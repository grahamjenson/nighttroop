package sitedata

import (
	"context"
	"unicode/utf8"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileSwitchesCollection struct{}

// ////
// Shared
// ////
func (s *ProfileSwitchesCollection) GetName() string {
	return "profile_switches"
}

// ////
// CREATE
// ////
func (s *ProfileSwitchesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteProfileSwitch) (*sitedb.ProfileSwitch, error) {
	if model.Name == "" {
		return nil, apiutils.ErrConstraintCheckFailedWithHint("name is required")
	}
	if herr := checkProfileSwitch(model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateProfileSwitch(ctx, sitemodels.ProfileSwitchToCreate(model, siteContext))
}

func (s *ProfileSwitchesCollection) CreateHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *ProfileSwitchesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteProfileSwitch, model *sitemodels.SiteProfileSwitch) (*sitedb.ProfileSwitch, error) {
	if herr := checkProfileSwitch(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateProfileSwitch(ctx, sitemodels.ProfileSwitchToUpdate(model, siteContext))
}

func (s *ProfileSwitchesCollection) UpdateHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchUpdated(model)
}

// ////
// Get Model
// /////
func (s *ProfileSwitchesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.ProfileSwitch, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetProfileSwitch(ctx, sitedb.GetProfileSwitchParams{ID: id})
}

// ToModel is the site model of a row
func (s *ProfileSwitchesCollection) ToModel(row *sitedb.ProfileSwitch, siteContext *httputils.SiteContext) *sitemodels.SiteProfileSwitch {
	return sitemodels.ProfileSwitchFromDB(row)
}

// GetHTMXResponse is the profile switch's details, shown in a modal
func (s *ProfileSwitchesCollection) GetHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchModal(model)
}

// ///
// Search Models
// ///

func (s *ProfileSwitchesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.ProfileSwitch, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllProfileSwitches(ctx, siteapiutils.QueryToAllProfileSwitches(query))
}

func (s *ProfileSwitchesCollection) SearchHTMXResponse(models []*sitemodels.SiteProfileSwitch, query *apiutils.Query) templ.Component {
	return site_data_templ.ProfileSwitchesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ProfileSwitchesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.SoftDeleteProfileSwitch(ctx, sitedb.SoftDeleteProfileSwitchParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *ProfileSwitchesCollection) DeleteHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchDeleted()
}

///////
// UTIL METHODS
////////

// checkProfileSwitch keeps each value sent within what a pump allows, and the duration to a week
func checkProfileSwitch(model *sitemodels.SiteProfileSwitch) *apiutils.HTTPError {
	if utf8.RuneCountInString(model.Name) > 64 {
		return apiutils.ErrConstraintCheckFailedWithHint("name must be at most 64 characters")
	}
	if p := model.Percent; p != nil && (*p < 10 || *p > 500) {
		return apiutils.ErrConstraintCheckFailedWithHint("percent must be from 10 to 500")
	}
	if t := model.Timeshift; t != nil && (*t < -24*60 || *t > 24*60) {
		return apiutils.ErrConstraintCheckFailedWithHint("timeshift must be from -1440 to 1440 minutes")
	}
	if d := model.DurationMinutes; d != nil && (*d < 0 || *d > 7*24*60) {
		return apiutils.ErrConstraintCheckFailedWithHint("duration must be from 0 to 10080 minutes, 0 is until the next switch")
	}
	return nil
}
