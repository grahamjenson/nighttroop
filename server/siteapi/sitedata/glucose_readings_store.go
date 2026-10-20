package sitedata

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
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
// Get Model
// /////
func (s *GlucoseReadingsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteGlucoseReading, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	reading, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: id})
	if err == sql.ErrNoRows || (err == nil && reading == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching glucose reading", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	return sitemodels.GlucoseReadingFromDB(reading), nil
}

// GetHTMXResponse is the reading's details, shown in a modal
func (s *GlucoseReadingsStore) GetHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingModal(model)
}

// ///
// Search Models
// ///

func (s *GlucoseReadingsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *siteapiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteGlucoseReading, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	readings, err := q.AllGlucoseReadings(ctx, sitequery.QueryToAllGlucoseReadings(query))
	if err != nil {
		slog.Error("Search glucose readings", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	models := []*sitemodels.SiteGlucoseReading{}
	for _, reading := range readings {
		models = append(models, sitemodels.GlucoseReadingFromDB(reading))
	}
	return models, nil
}

func (s *GlucoseReadingsStore) SearchHTMXResponse(models []*sitemodels.SiteGlucoseReading, query *siteapiutils.Query) templ.Component {
	return site_data_templ.GlucoseReadingsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *GlucoseReadingsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteGlucoseReading, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting glucose reading", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *GlucoseReadingsStore) DeleteHTMXResponse(model *sitemodels.SiteGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingDeleted()
}
