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
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxTempTargetMinutes is a week, far longer than any temp target is set for
const maxTempTargetMinutes = 7 * 24 * 60

type TempTargetsStore struct{}

// ////
// Shared
// ////
func (s *TempTargetsStore) GetName() string {
	return "temp_targets"
}

// ////
// CREATE
// ////
func (s *TempTargetsStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteTempTarget) (bool, *apiutils.HTTPError) {
	if herr := checkTempTarget(model, siteContext); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateTempTarget(ctx, sitemodels.TempTargetToCreate(model, siteContext))
	if err != nil {
		return false, tempTargetWriteError(err)
	}
	*model = *sitemodels.TempTargetFromDB(created, siteContext)
	return true, nil
}

func (s *TempTargetsStore) CreateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes a temp target's targets, duration, notes or reason; anything not sent stays as it is
func (s *TempTargetsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempTarget, model *sitemodels.SiteTempTarget) *apiutils.HTTPError {
	// Check the temp target as it will be once the fields sent are changed
	merged := *oldModel
	if model.DurationMinutes != 0 {
		merged.DurationMinutes = model.DurationMinutes
	}
	if model.LowTarget != 0 {
		merged.LowTarget = model.LowTarget
	}
	if model.HighTarget != 0 {
		merged.HighTarget = model.HighTarget
	}
	if herr := checkTempTarget(&merged, siteContext); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateTempTarget(ctx, sitemodels.TempTargetToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return tempTargetWriteError(err)
	}
	*model = *sitemodels.TempTargetFromDB(updated, siteContext)
	return nil
}

func (s *TempTargetsStore) UpdateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempTargetsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteTempTarget, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	tempTarget, err := q.GetTempTarget(ctx, sitedb.GetTempTargetParams{ID: id})
	return apiutils.Found(tempTarget, err, "temp target", apiutils.WithContext(sitemodels.TempTargetFromDB, siteContext))
}

// GetHTMXResponse is the temp target's details, shown in a modal
func (s *TempTargetsStore) GetHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetModal(model)
}

// ///
// Search Models
// ///

func (s *TempTargetsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteTempTarget, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	tempTargets, err := q.AllTempTargets(ctx, sitequery.QueryToAllTempTargets(query))
	return apiutils.Searched(tempTargets, err, "temp targets", apiutils.WithContext(sitemodels.TempTargetFromDB, siteContext))
}

func (s *TempTargetsStore) SearchHTMXResponse(models []*sitemodels.SiteTempTarget, query *apiutils.Query) templ.Component {
	return site_data_templ.TempTargetsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempTargetsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteTempTarget, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteTempTarget(ctx, sitedb.SoftDeleteTempTargetParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting temp target", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *TempTargetsStore) DeleteHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetDeleted()
}

///////
// UTIL METHODS
////////

// checkTempTarget keeps targets within real glucose values, low at most high, and the duration to a week.
// The targets are in the settings' units.
func checkTempTarget(model *sitemodels.SiteTempTarget, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	units := sitemodels.ContextUnits(siteContext)
	if model.LowTarget <= 0 || units.ToMGDL(model.HighTarget) > maxGlucoseMgdl || model.LowTarget > model.HighTarget {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("targets must be more than 0, at most %v %s, and low must be at most high", consts.Round(units.FromMGDL(maxGlucoseMgdl), 1), *units))
	}
	if model.DurationMinutes <= 0 || model.DurationMinutes > maxTempTargetMinutes {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("duration must be more than 0 and at most %d minutes", maxTempTargetMinutes))
	}
	return nil
}

func tempTargetWriteError(err error) *apiutils.HTTPError {
	if errors.IsSQLConstraintCheck(err) {
		return apiutils.ErrConstraintCheckFailedWithHint(err.Error())
	}
	slog.Error("Error writing temp target", "error", err)
	return apiutils.ErrInternalServerError
}
