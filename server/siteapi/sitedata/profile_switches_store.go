package sitedata

import (
	"context"
	"database/sql"
	"log/slog"
	"unicode/utf8"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileSwitchesStore struct{}

// ////
// Shared
// ////
func (s *ProfileSwitchesStore) GetName() string {
	return "profile_switches"
}

// ////
// CREATE
// ////
func (s *ProfileSwitchesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteProfileSwitch) (bool, *apiutils.HTTPError) {
	if model.Name == "" {
		return false, apiutils.ErrConstraintCheckFailedWithHint("name is required")
	}
	if herr := checkProfileSwitch(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateProfileSwitch(ctx, sitemodels.ProfileSwitchToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "profile switch")
	}
	tx.Created(created)
	*model = *sitemodels.ProfileSwitchFromDB(created)
	return true, nil
}

func (s *ProfileSwitchesStore) CreateHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes the fields sent; anything not sent stays as it is
func (s *ProfileSwitchesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteProfileSwitch, model *sitemodels.SiteProfileSwitch) *apiutils.HTTPError {
	if herr := checkProfileSwitch(model); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateProfileSwitch(ctx, sitemodels.ProfileSwitchToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "profile switch")
	}
	tx.Updated(updated)
	*model = *sitemodels.ProfileSwitchFromDB(updated)
	return nil
}

func (s *ProfileSwitchesStore) UpdateHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchUpdated(model)
}

// ////
// Get Model
// /////
func (s *ProfileSwitchesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteProfileSwitch, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	profileSwitch, err := q.GetProfileSwitch(ctx, sitedb.GetProfileSwitchParams{ID: id})
	return apiutils.Found(profileSwitch, err, "profile switch", sitemodels.ProfileSwitchFromDB)
}

// GetHTMXResponse is the profile switch's details, shown in a modal
func (s *ProfileSwitchesStore) GetHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchModal(model)
}

// ///
// Search Models
// ///

func (s *ProfileSwitchesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteProfileSwitch, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	profileSwitches, err := q.AllProfileSwitches(ctx, sitequery.QueryToAllProfileSwitches(query))
	return apiutils.Searched(profileSwitches, err, "profile switches", sitemodels.ProfileSwitchFromDB)
}

func (s *ProfileSwitchesStore) SearchHTMXResponse(models []*sitemodels.SiteProfileSwitch, query *apiutils.Query) templ.Component {
	return site_data_templ.ProfileSwitchesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ProfileSwitchesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteProfileSwitch, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	tx.Deleted(q.GetProfileSwitch(ctx, sitedb.GetProfileSwitchParams{ID: model.ID}))
	_, err = q.SoftDeleteProfileSwitch(ctx, sitedb.SoftDeleteProfileSwitchParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting profile switch", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *ProfileSwitchesStore) DeleteHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
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
