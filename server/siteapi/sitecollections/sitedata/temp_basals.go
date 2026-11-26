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

type TempBasalsCollection struct{}

// ////
// Shared
// ////
func (s *TempBasalsCollection) GetName() string {
	return "temp_basals"
}

// GetRoute is where the collection is in its URL
func (s *TempBasalsCollection) GetRoute() string {
	return "temp-basals"
}

// ////
// CREATE
// ////
// CreateRow adds a temp basal of a rate, as the site keeps one typed in
func (s *TempBasalsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteTempBasal) (*sitedb.TempBasal, error) {
	// A rate of 0 is sent, a suspended basal
	if model.Rate == nil {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.required", i18n.Data{"Field": "rate"}))
	}
	if err := rateOrMultiplier(model); err != nil {
		return nil, err
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateTempBasal, sitemodels.TempBasalToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add a temp basal in
func (s *TempBasalsCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.TempBasalNew()
}

func (s *TempBasalsCollection) CreateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a temp basal's rate or multiplier, duration, notes or reason; anything not sent stays as it is. A
// rate sent makes it a rate, and a multiplier a multiplier.
func (s *TempBasalsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempBasal, model *sitemodels.SiteTempBasal) (*sitedb.TempBasal, error) {
	if err := rateOrMultiplier(model); err != nil {
		return nil, err
	}
	model.ID = oldModel.ID
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateTempBasal, sitemodels.TempBasalToUpdate(model, siteContext))
}

// rateOrMultiplier refuses a temp basal sent with both a rate and a multiplier, as it is one or the other
func rateOrMultiplier(model *sitemodels.SiteTempBasal) error {
	if model.Rate != nil && model.Multiplier != nil {
		return apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.data.rate_or_multiplier"))
	}
	return nil
}

func (s *TempBasalsCollection) UpdateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempBasalsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.TempBasal, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetTempBasal, sitedb.GetTempBasalParams{ID: id})
}

// ToModel is the site model of a row
func (s *TempBasalsCollection) ToModel(row *sitedb.TempBasal, siteContext *httputils.SiteContext) *sitemodels.SiteTempBasal {
	return sitemodels.TempBasalFromDB(row)
}

// FillModels gives the temp basals their scheduled basal, so they show their rate and percent at their start
func (s *TempBasalsCollection) FillModels(ctx context.Context, collectionTx *siteapiutils.CollectionTx, models []*sitemodels.SiteTempBasal, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	return sitemodels.FillScheduledBasals(ctx, q, siteContext.GetSiteTimezone(), models)
}

// GetHTMXResponse is the temp basal's details, shown in a modal
func (s *TempBasalsCollection) GetHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalModal(model)
}

// ///
// Search Models
// ///

func (s *TempBasalsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.TempBasal, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllTempBasals, siteapiutils.QueryToAllTempBasals(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *TempBasalsCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).TempBasalTotals)
}

func (s *TempBasalsCollection) SearchHTMXResponse(models []*sitemodels.SiteTempBasal, query *apiutils.Query) templ.Component {
	return site_data_templ.TempBasalsTable(models, query.Skip, query.Limit, query.Before)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempBasalsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteTempBasal)
}

func (s *TempBasalsCollection) DeleteHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalDeleted()
}
