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
func (s *InsulinDosesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteInsulinDose) (bool, *siteapiutils.HTTPError) {
	if herr := checkInsulinDose(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, siteapiutils.ErrInternalServerError
	}
	created, err := q.CreateInsulinDose(ctx, sitemodels.InsulinDoseToCreate(model, siteContext))
	if err != nil {
		return false, insulinDoseWriteError(err)
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
func (s *InsulinDosesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteInsulinDose, model *sitemodels.SiteInsulinDose) *siteapiutils.HTTPError {
	if model.Dose != 0 {
		if herr := checkInsulinDose(model); herr != nil {
			return herr
		}
	}
	q, err := tx.Site()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateInsulinDose(ctx, sitemodels.InsulinDoseToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return siteapiutils.ErrNotFound
	}
	if err != nil {
		return insulinDoseWriteError(err)
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
func (s *InsulinDosesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteInsulinDose, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	insulinDose, err := q.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id})
	if err == sql.ErrNoRows || (err == nil && insulinDose == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching insulin dose", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	return sitemodels.InsulinDoseFromDB(insulinDose), nil
}

// GetHTMXResponse is the insulin dose's details, shown in a modal
func (s *InsulinDosesStore) GetHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseModal(model)
}

// ///
// Search Models
// ///

func (s *InsulinDosesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *siteapiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteInsulinDose, *siteapiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, siteapiutils.ErrInternalServerError
	}
	insulinDoses, err := q.AllInsulinDoses(ctx, sitequery.QueryToAllInsulinDoses(query))
	if err != nil {
		slog.Error("Search insulin doses", "error", err)
		return nil, siteapiutils.ErrInternalServerError
	}
	models := []*sitemodels.SiteInsulinDose{}
	for _, insulinDose := range insulinDoses {
		models = append(models, sitemodels.InsulinDoseFromDB(insulinDose))
	}
	return models, nil
}

func (s *InsulinDosesStore) SearchHTMXResponse(models []*sitemodels.SiteInsulinDose, query *siteapiutils.Query) templ.Component {
	return site_data_templ.InsulinDosesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *InsulinDosesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteInsulinDose, siteContext *httputils.SiteContext) *siteapiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return siteapiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting insulin dose", "error", err)
		return siteapiutils.ErrInternalServerError
	}
	return nil
}

func (s *InsulinDosesStore) DeleteHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseDeleted()
}

///////
// UTIL METHODS
////////

func checkInsulinDose(model *sitemodels.SiteInsulinDose) *siteapiutils.HTTPError {
	if model.Dose <= 0 || model.Dose > maxInsulinDoseUnits {
		return siteapiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("dose must be more than 0 and at most %d units", maxInsulinDoseUnits))
	}
	return nil
}

func insulinDoseWriteError(err error) *siteapiutils.HTTPError {
	if errors.IsSQLConstraintCheck(err) {
		return siteapiutils.ErrConstraintCheckFailedWithHint(err.Error())
	}
	slog.Error("Error writing insulin dose", "error", err)
	return siteapiutils.ErrInternalServerError
}
