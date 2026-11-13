package sitedata

import (
	"cmp"
	"context"
	"unicode/utf8"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/i18n"

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

// GetRoute is where the collection is in its URL
func (s *ProfileSwitchesCollection) GetRoute() string {
	return "profile-switches"
}

// ////
// CREATE
// ////
func (s *ProfileSwitchesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteProfileSwitch) (*sitedb.ProfileSwitch, error) {
	if model.Name == "" {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.name_required"))
	}
	if herr := checkProfileSwitch(model); herr != nil {
		return nil, herr
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateProfileSwitch, sitemodels.ProfileSwitchToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add a profile switch in
func (s *ProfileSwitchesCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.ProfileSwitchNew()
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
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateProfileSwitch, sitemodels.ProfileSwitchToUpdate(model, siteContext))
}

func (s *ProfileSwitchesCollection) UpdateHTMXResponse(model *sitemodels.SiteProfileSwitch) templ.Component {
	return site_data_templ.ProfileSwitchUpdated(model)
}

// ////
// Get Model
// /////
func (s *ProfileSwitchesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.ProfileSwitch, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetProfileSwitch, sitedb.GetProfileSwitchParams{ID: id})
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
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllProfileSwitches, siteapiutils.QueryToAllProfileSwitches(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *ProfileSwitchesCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).ProfileSwitchTotals)
}

func (s *ProfileSwitchesCollection) SearchHTMXResponse(models []*sitemodels.SiteProfileSwitch, query *apiutils.Query) templ.Component {
	return site_data_templ.ProfileSwitchesTable(models, query.Skip, query.Limit, query.Before)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ProfileSwitchesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteProfileSwitch)
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
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.name_too_long"))
	}
	if herr := cmp.Or(
		apiutils.CheckRange("errors.data.multiplier", model.Multiplier, 0.1, 5),
		apiutils.CheckRange("errors.data.timeshift", model.Timeshift, -24*60, 24*60),
	); herr != nil {
		return herr
	}
	if d := model.DurationMinutes; d != nil && (*d < 0 || *d > 7*24*60) {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.profile_switch_duration"))
	}
	return nil
}
