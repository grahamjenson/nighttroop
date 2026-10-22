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
func (s *InsulinDosesStore) CreateRow(ctx context.Context, q *sitedb.Queries, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteInsulinDose) (*sitedb.InsulinDose, error) {
	if herr := checkInsulinDose(model); herr != nil {
		return nil, herr
	}
	return q.CreateInsulinDose(ctx, sitemodels.InsulinDoseToCreate(model, siteContext))
}

func (s *InsulinDosesStore) CreateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a dose's units, type, notes or reason; anything not sent stays as it is
func (s *InsulinDosesStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteInsulinDose, model *sitemodels.SiteInsulinDose) (*sitedb.InsulinDose, error) {
	if model.Dose != 0 {
		if herr := checkInsulinDose(model); herr != nil {
			return nil, herr
		}
	}
	model.ID = oldModel.ID
	return q.UpdateInsulinDose(ctx, sitemodels.InsulinDoseToUpdate(model, siteContext))
}

func (s *InsulinDosesStore) UpdateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseUpdated(model)
}

// ////
// Get Model
// /////
func (s *InsulinDosesStore) GetRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) (*sitedb.InsulinDose, error) {
	return q.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id})
}

// ToModel is the site model of a row
func (s *InsulinDosesStore) ToModel(row *sitedb.InsulinDose, siteContext *httputils.SiteContext) *sitemodels.SiteInsulinDose {
	return sitemodels.InsulinDoseFromDB(row)
}

// GetHTMXResponse is the insulin dose's details, shown in a modal
func (s *InsulinDosesStore) GetHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseModal(model)
}

// ///
// Search Models
// ///

func (s *InsulinDosesStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.InsulinDose, error) {
	return q.AllInsulinDoses(ctx, sitequery.QueryToAllInsulinDoses(query))
}

func (s *InsulinDosesStore) SearchHTMXResponse(models []*sitemodels.SiteInsulinDose, query *apiutils.Query) templ.Component {
	return site_data_templ.InsulinDosesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *InsulinDosesStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
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
