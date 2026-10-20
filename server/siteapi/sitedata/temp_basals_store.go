package sitedata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxTempBasalMinutes is a day, the longest pumps set a temp basal for
const maxTempBasalMinutes = 24 * 60

// maxTempBasalUnitsPerHour is far above any pump's maximum basal rate
const maxTempBasalUnitsPerHour = 50

type TempBasalsStore struct{}

// ////
// Shared
// ////
func (s *TempBasalsStore) GetName() string {
	return "temp_basals"
}

// ////
// CREATE
// ////
func (s *TempBasalsStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteTempBasal) (bool, *siteapiutils.HTTPError) {
	if herr := checkTempBasal(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, siteapiutils.ErrInternalServerError
	}
	created, err := q.CreateTempBasal(ctx, sitemodels.TempBasalToCreate(model, siteContext))
	if err != nil {
		return false, tempBasalWriteError(err)
	}
	*model = *sitemodels.TempBasalFromDB(created)
	return true, nil
}

func (s *TempBasalsStore) CreateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes a temp basal's rate, duration, notes or reason; anything not sent stays as it is
func (s *TempBasalsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempBasal, model *sitemodels.SiteTempBasal) *siteapiutils.HTTPError {
	// Check the temp basal as it will be once the fields sent are changed
	merged := *oldModel
	if model.DurationMinutes != 0 {
		merged.DurationMinutes = model.DurationMinutes
	}
	if model.Absolute != nil {
		merged.Absolute = model.Absolute
	}
	if herr := checkTempBasal(&merged); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateTempBasal(ctx, sitemodels.TempBasalToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return siteapiutils.ErrNotFound
	}
	if err != nil {
		return tempBasalWriteError(err)
	}
	*model = *sitemodels.TempBasalFromDB(updated)
	return nil
}

func (s *TempBasalsStore) UpdateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempBasalsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteTempBasal, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	tempBasal, err := q.GetTempBasal(ctx, sitedb.GetTempBasalParams{ID: id})
	if err == sql.ErrNoRows || (err == nil && tempBasal == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching temp basal", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	return sitemodels.TempBasalFromDB(tempBasal), nil
}

// GetHTMXResponse is the temp basal's details, shown in a modal
func (s *TempBasalsStore) GetHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalModal(model)
}

// ///
// Search Models
// ///

func (s *TempBasalsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *siteapiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteTempBasal, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	tempBasals, err := q.AllTempBasals(ctx, sitequery.QueryToAllTempBasals(query))
	if err != nil {
		slog.Error("Search temp basals", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	models := []*sitemodels.SiteTempBasal{}
	for _, tempBasal := range tempBasals {
		models = append(models, sitemodels.TempBasalFromDB(tempBasal))
	}
	return models, nil
}

func (s *TempBasalsStore) SearchHTMXResponse(models []*sitemodels.SiteTempBasal, query *siteapiutils.Query) templ.Component {
	return site_data_templ.TempBasalsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempBasalsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteTempBasal, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteTempBasal(ctx, sitedb.SoftDeleteTempBasalParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting temp basal", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *TempBasalsStore) DeleteHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalDeleted()
}

///////
// UTIL METHODS
////////

// checkTempBasal keeps the rate to what a pump can deliver, and the duration to a day
func checkTempBasal(model *sitemodels.SiteTempBasal) *siteapiutils.HTTPError {
	if model.Absolute == nil || *model.Absolute < 0 || *model.Absolute > maxTempBasalUnitsPerHour {
		return siteapiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("absolute must be from 0 to %d U/hr", maxTempBasalUnitsPerHour))
	}
	if model.DurationMinutes <= 0 || model.DurationMinutes > maxTempBasalMinutes {
		return siteapiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("duration must be more than 0 and at most %d minutes", maxTempBasalMinutes))
	}
	return nil
}

func tempBasalWriteError(err error) *siteapiutils.HTTPError {
	if errors.IsSQLConstraintCheck(err) {
		return siteapiutils.ErrConstraintCheckFailedWithHint(err.Error())
	}
	slog.Error("Error writing temp basal", "error", err)
	return siteapiutils.ErrInternalServerError
}
