package sitedata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

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
func (s *GlucoseReadingsStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteGlucoseReading) (bool, *apiutils.HTTPError) {
	params := sitemodels.GlucoseReadingToCreate(model, siteContext)
	if herr := checkGlucoseMgdl(params.Mgdl, siteContext); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateGlucoseReading(ctx, params)
	if err != nil {
		return false, apiutils.WriteError(err, "glucose reading")
	}
	*model = *sitemodels.GlucoseReadingFromDB(created, siteContext)
	return true, nil
}

func (s *GlucoseReadingsStore) CreateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes a reading's value, notes or reason; anything not sent stays as it is
func (s *GlucoseReadingsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteGlucoseReading, model *sitemodels.SiteGlucoseReading) *apiutils.HTTPError {
	model.ID = oldModel.ID
	params := sitemodels.GlucoseReadingToUpdate(model, siteContext)
	if params.Mgdl != nil {
		if herr := checkGlucoseMgdl(*params.Mgdl, siteContext); herr != nil {
			return herr
		}
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	updated, err := q.UpdateGlucoseReading(ctx, params)
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "glucose reading")
	}
	*model = *sitemodels.GlucoseReadingFromDB(updated, siteContext)
	return nil
}

func (s *GlucoseReadingsStore) UpdateHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingUpdated(model)
}

// ////
// Get Model
// /////
func (s *GlucoseReadingsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteGlucoseReading, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	reading, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: id})
	return apiutils.Found(reading, err, "glucose reading", apiutils.WithContext(sitemodels.GlucoseReadingFromDB, siteContext))
}

// GetHTMXResponse is the reading's details, shown in a modal
func (s *GlucoseReadingsStore) GetHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingModal(model)
}

// ///
// Search Models
// ///

func (s *GlucoseReadingsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteGlucoseReading, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	readings, err := q.AllGlucoseReadings(ctx, sitequery.QueryToAllGlucoseReadings(query))
	return apiutils.Searched(readings, err, "glucose readings", apiutils.WithContext(sitemodels.GlucoseReadingFromDB, siteContext))
}

func (s *GlucoseReadingsStore) SearchHTMXResponse(models []*sitemodels.SiteGlucoseReading, query *apiutils.Query) templ.Component {
	return site_data_templ.GlucoseReadingsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *GlucoseReadingsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteGlucoseReading, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting glucose reading", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *GlucoseReadingsStore) DeleteHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
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
