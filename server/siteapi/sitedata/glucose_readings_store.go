package sitedata

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxGlucoseMgdl is far above any real reading, so a typo or a made up value is refused
const maxGlucoseMgdl = 1000

type GlucoseReadingsStore struct{}

// ////
// Shared
// ////
func (s *GlucoseReadingsStore) GetName() string {
	return "glucose_readings"
}

// ////
// CREATE
// ////
func (s *GlucoseReadingsStore) CreateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, model *sitemodels.SiteGlucoseReading) (*sitedb.GlucoseReading, error) {
	params := sitemodels.GlucoseReadingToCreate(model, siteContext)
	if herr := checkGlucoseMgdl(params.Mgdl, siteContext); herr != nil {
		return nil, herr
	}
	return q.CreateGlucoseReading(ctx, params)
}

func (s *GlucoseReadingsStore) CreateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a reading's value, notes or reason; anything not sent stays as it is
func (s *GlucoseReadingsStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteGlucoseReading, model *sitemodels.SiteGlucoseReading) (*sitedb.GlucoseReading, error) {
	model.ID = oldModel.ID
	params := sitemodels.GlucoseReadingToUpdate(model, siteContext)
	if params.Mgdl != nil {
		if herr := checkGlucoseMgdl(*params.Mgdl, siteContext); herr != nil {
			return nil, herr
		}
	}
	return q.UpdateGlucoseReading(ctx, params)
}

func (s *GlucoseReadingsStore) UpdateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingUpdated(model)
}

// ////
// Get Model
// /////
func (s *GlucoseReadingsStore) GetRow(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.GlucoseReading, error) {
	return q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: id})
}

// ToModel is the site model of a row
func (s *GlucoseReadingsStore) ToModel(row *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *sitemodels.SiteGlucoseReading {
	return sitemodels.GlucoseReadingFromDB(row, siteContext)
}

// GetHTMXResponse is the reading's details, shown in a modal
func (s *GlucoseReadingsStore) GetHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingModal(model)
}

// ///
// Search Models
// ///

func (s *GlucoseReadingsStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query) ([]*sitedb.GlucoseReading, error) {
	return q.AllGlucoseReadings(ctx, sitequery.QueryToAllGlucoseReadings(query))
}

func (s *GlucoseReadingsStore) SearchHTMXResponse(models []*sitemodels.SiteGlucoseReading, query *apiutils.Query) templ.Component {
	return site_data_templ.GlucoseReadingsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *GlucoseReadingsStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *GlucoseReadingsStore) DeleteHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingDeleted()
}

///////
// UTIL METHODS
////////

func checkGlucoseMgdl(mgdl float64, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	if mgdl <= 0 || mgdl > maxGlucoseMgdl {
		units := sitemodels.ContextUnits(siteContext)
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("reading must be more than 0 and at most %v %s", consts.Round(units.FromMGDL(maxGlucoseMgdl), 1), *units))
	}
	return nil
}
