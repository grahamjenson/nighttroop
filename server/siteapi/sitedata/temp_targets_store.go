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
func (s *TempTargetsStore) CreateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, model *sitemodels.SiteTempTarget) (*sitedb.TempTarget, error) {
	if herr := checkTempTarget(model, siteContext); herr != nil {
		return nil, herr
	}
	return q.CreateTempTarget(ctx, sitemodels.TempTargetToCreate(model, siteContext))
}

func (s *TempTargetsStore) CreateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a temp target's targets, duration, notes or reason; anything not sent stays as it is
func (s *TempTargetsStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempTarget, model *sitemodels.SiteTempTarget) (*sitedb.TempTarget, error) {
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
		return nil, herr
	}
	model.ID = oldModel.ID
	return q.UpdateTempTarget(ctx, sitemodels.TempTargetToUpdate(model, siteContext))
}

func (s *TempTargetsStore) UpdateHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempTargetsStore) GetRow(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.TempTarget, error) {
	return q.GetTempTarget(ctx, sitedb.GetTempTargetParams{ID: id})
}

// ToModel is the site model of a row
func (s *TempTargetsStore) ToModel(row *sitedb.TempTarget, siteContext *httputils.SiteContext) *sitemodels.SiteTempTarget {
	return sitemodels.TempTargetFromDB(row, siteContext)
}

// GetHTMXResponse is the temp target's details, shown in a modal
func (s *TempTargetsStore) GetHTMXResponse(model *sitemodels.SiteTempTarget) templ.Component {
	return site_data_templ.TempTargetModal(model)
}

// ///
// Search Models
// ///

func (s *TempTargetsStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query) ([]*sitedb.TempTarget, error) {
	return q.AllTempTargets(ctx, sitequery.QueryToAllTempTargets(query))
}

func (s *TempTargetsStore) SearchHTMXResponse(models []*sitemodels.SiteTempTarget, query *apiutils.Query) templ.Component {
	return site_data_templ.TempTargetsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempTargetsStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteTempTarget(ctx, sitedb.SoftDeleteTempTargetParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
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
	if model.LowTarget <= 0 || units.ToMGDL(model.HighTarget) > consts.MaxGlucoseMgdl || model.LowTarget > model.HighTarget {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("targets must be more than 0, at most %v %s, and low must be at most high", consts.Round(units.FromMGDL(consts.MaxGlucoseMgdl), 1), *units))
	}
	if model.DurationMinutes <= 0 || model.DurationMinutes > maxTempTargetMinutes {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("duration must be more than 0 and at most %d minutes", maxTempTargetMinutes))
	}
	return nil
}
