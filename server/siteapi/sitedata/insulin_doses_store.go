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
	"nighttroop.t1d.nz/utils/httputils"
)

// maxInsulinDoseUnits is far above any single dose a pump or pen gives
const maxInsulinDoseUnits = 100

type InsulinDosesStore struct{}

// ////
// Shared
// ////
func (s *InsulinDosesStore) GetName() string {
	return "insulin_doses"
}

// ////
// CREATE
// ////
func (s *InsulinDosesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteInsulinDose) (bool, *apiutils.HTTPError) {
	if herr := checkInsulinDose(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateInsulinDose(ctx, sitemodels.InsulinDoseToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "insulin dose")
	}
	*model = *sitemodels.InsulinDoseFromDB(created)
	return true, nil
}

func (s *InsulinDosesStore) CreateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes a dose's units, type, notes or reason; anything not sent stays as it is
func (s *InsulinDosesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteInsulinDose, model *sitemodels.SiteInsulinDose) *apiutils.HTTPError {
	if model.Dose != 0 {
		if herr := checkInsulinDose(model); herr != nil {
			return herr
		}
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateInsulinDose(ctx, sitemodels.InsulinDoseToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "insulin dose")
	}
	*model = *sitemodels.InsulinDoseFromDB(updated)
	return nil
}

func (s *InsulinDosesStore) UpdateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseUpdated(model)
}

// ////
// Get Model
// /////
func (s *InsulinDosesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteInsulinDose, *apiutils.HTTPError) {
	return getModel(ctx, tx, (*sitedb.Queries).GetInsulinDose, sitedb.GetInsulinDoseParams{ID: id}, "insulin dose", sitemodels.InsulinDoseFromDB)
}

// GetHTMXResponse is the insulin dose's details, shown in a modal
func (s *InsulinDosesStore) GetHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseModal(model)
}

// ///
// Search Models
// ///

func (s *InsulinDosesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteInsulinDose, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	insulinDoses, err := q.AllInsulinDoses(ctx, sitequery.QueryToAllInsulinDoses(query))
	return apiutils.Searched(insulinDoses, err, "insulin doses", sitemodels.InsulinDoseFromDB)
}

func (s *InsulinDosesStore) SearchHTMXResponse(models []*sitemodels.SiteInsulinDose, query *apiutils.Query) templ.Component {
	return site_data_templ.InsulinDosesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *InsulinDosesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteInsulinDose, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting insulin dose", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *InsulinDosesStore) DeleteHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseDeleted()
}

///////
// UTIL METHODS
////////

func checkInsulinDose(model *sitemodels.SiteInsulinDose) *apiutils.HTTPError {
	if model.Dose <= 0 || model.Dose > maxInsulinDoseUnits {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("dose must be more than 0 and at most %d units", maxInsulinDoseUnits))
	}
	return nil
}
