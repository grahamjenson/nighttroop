package sitedata

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type GlucoseReadingsCollection struct{}

// ////
// Shared
// ////
func (s *GlucoseReadingsCollection) GetName() string {
	return "glucose"
}

// GetRoute is where the collection is in its URL
func (s *GlucoseReadingsCollection) GetRoute() string {
	return "glucose"
}

// ////
// CREATE
// ////
func (s *GlucoseReadingsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteGlucoseReading) (*sitedb.GlucoseReading, error) {
	params := sitemodels.GlucoseReadingToCreate(model, siteContext)
	if herr := checkGlucoseMgdl(params.Mgdl, siteContext); herr != nil {
		return nil, herr
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateGlucoseReading, params)
}

// NewHTMXResponse is the modal to add a glucose reading in
func (s *GlucoseReadingsCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.GlucoseReadingNew()
}

func (s *GlucoseReadingsCollection) CreateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a reading's value, CGM fields (delta, direction, noise, filtered, unfiltered, rssi), notes or reason; anything not sent stays as it is
func (s *GlucoseReadingsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteGlucoseReading, model *sitemodels.SiteGlucoseReading) (*sitedb.GlucoseReading, error) {
	model.ID = oldModel.ID
	params := sitemodels.GlucoseReadingToUpdate(model, siteContext)
	if params.Mgdl != nil {
		if herr := checkGlucoseMgdl(*params.Mgdl, siteContext); herr != nil {
			return nil, herr
		}
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateGlucoseReading, params)
}

func (s *GlucoseReadingsCollection) UpdateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingUpdated(model)
}

// ////
// Get Model
// /////
func (s *GlucoseReadingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.GlucoseReading, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetGlucoseReading, sitedb.GetGlucoseReadingParams{ID: id})
}

// ToModel is the site model of a row
func (s *GlucoseReadingsCollection) ToModel(row *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *sitemodels.SiteGlucoseReading {
	return sitemodels.GlucoseReadingFromDB(row, siteContext)
}

// GetHTMXResponse is the reading's details, shown in a modal
func (s *GlucoseReadingsCollection) GetHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingModal(model)
}

// ///
// Search Models
// ///

func (s *GlucoseReadingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.GlucoseReading, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllGlucoseReadings, siteapiutils.QueryToAllGlucoseReadings(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *GlucoseReadingsCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).GlucoseReadingTotals)
}

func (s *GlucoseReadingsCollection) SearchHTMXResponse(models []*sitemodels.SiteGlucoseReading, query *apiutils.Query) templ.Component {
	return site_data_templ.GlucoseReadingsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *GlucoseReadingsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteGlucoseReading)
}

func (s *GlucoseReadingsCollection) DeleteHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingDeleted()
}

///////
// UTIL METHODS
////////

func checkGlucoseMgdl(mgdl float64, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	if mgdl <= 0 || mgdl > consts.MaxGlucoseMgdl {
		units := sitemodels.ContextUnits(siteContext)
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("reading must be more than 0 and at most %v %s", consts.Round(units.FromMGDL(consts.MaxGlucoseMgdl), 1), *units))
	}
	return nil
}
