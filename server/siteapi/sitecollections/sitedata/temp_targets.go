package sitedata

import (
	"context"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/i18n"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TempTargetsCollection struct{}

// ////
// Shared
// ////
func (s *TempTargetsCollection) GetName() string {
	return "temp_targets"
}

// GetRoute is where the collection is in its URL
func (s *TempTargetsCollection) GetRoute() string {
	return "temp-targets"
}

// ////
// CREATE
// ////
func (s *TempTargetsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteTempTarget) (*sitedb.TempTarget, error) {
	if herr := checkTempTarget(model); herr != nil {
		return nil, herr
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateTempTarget, sitemodels.TempTargetToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add a temp target in
func (s *TempTargetsCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.TempTargetNew()
}

func (s *TempTargetsCollection) CreateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a temp target's targets, duration, notes or reason; anything not sent stays as it is
func (s *TempTargetsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempTarget, model *sitemodels.SiteTempTarget) (*sitedb.TempTarget, error) {
	// Check the temp target as it will be once the fields sent are changed
	merged := *oldModel
	if model.LowTarget != 0 {
		merged.LowTarget = model.LowTarget
	}
	if model.HighTarget != 0 {
		merged.HighTarget = model.HighTarget
	}
	if herr := checkTempTarget(&merged); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateTempTarget, sitemodels.TempTargetToUpdate(model, siteContext))
}

func (s *TempTargetsCollection) UpdateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempTargetsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.TempTarget, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetTempTarget, sitedb.GetTempTargetParams{ID: id})
}

// ToModel is the site model of a row
func (s *TempTargetsCollection) ToModel(row *sitedb.TempTarget, siteContext *httputils.SiteContext) *sitemodels.SiteTempTarget {
	return sitemodels.TempTargetFromDB(row, siteContext)
}

// GetHTMXResponse is the temp target's details, shown in a modal
func (s *TempTargetsCollection) GetHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetModal(model)
}

// ///
// Search Models
// ///

func (s *TempTargetsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.TempTarget, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllTempTargets, siteapiutils.QueryToAllTempTargets(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *TempTargetsCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).TempTargetTotals)
}

func (s *TempTargetsCollection) SearchHTMXResponse(models []*sitemodels.SiteTempTarget, query *apiutils.Query) templ.Component {
	return site_data_templ.TempTargetsTable(models, query.Skip, query.Limit, query.Before)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempTargetsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteTempTarget)
}

func (s *TempTargetsCollection) DeleteHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetDeleted()
}

///////
// UTIL METHODS
////////

// checkTempTarget keeps the low target at most the high
func checkTempTarget(model *sitemodels.SiteTempTarget) *apiutils.HTTPError {
	if model.LowTarget > model.HighTarget {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.targets"))
	}
	return nil
}
