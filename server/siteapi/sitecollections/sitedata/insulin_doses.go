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
	"nighttroop.t1d.nz/utils/httputils"
)

// maxInsulinDoseUnits is far above any single dose a pump or pen gives
const maxInsulinDoseUnits = 100

type InsulinDosesCollection struct{}

// ////
// Shared
// ////
func (s *InsulinDosesCollection) GetName() string {
	return "insulin_doses"
}

// GetRoute is where the collection is in its URL
func (s *InsulinDosesCollection) GetRoute() string {
	return "insulin"
}

// ////
// CREATE
// ////
func (s *InsulinDosesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteInsulinDose) (*sitedb.InsulinDose, error) {
	if herr := checkInsulinDose(model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateInsulinDose(ctx, sitemodels.InsulinDoseToCreate(model, siteContext))
}

func (s *InsulinDosesCollection) CreateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a dose's units, type, notes or reason; anything not sent stays as it is
func (s *InsulinDosesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteInsulinDose, model *sitemodels.SiteInsulinDose) (*sitedb.InsulinDose, error) {
	if model.Dose != 0 {
		if herr := checkInsulinDose(model); herr != nil {
			return nil, herr
		}
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateInsulinDose(ctx, sitemodels.InsulinDoseToUpdate(model, siteContext))
}

func (s *InsulinDosesCollection) UpdateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseUpdated(model)
}

// ////
// Get Model
// /////
func (s *InsulinDosesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.InsulinDose, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetInsulinDose, sitedb.GetInsulinDoseParams{ID: id})
}

// ToModel is the site model of a row
func (s *InsulinDosesCollection) ToModel(row *sitedb.InsulinDose, siteContext *httputils.SiteContext) *sitemodels.SiteInsulinDose {
	return sitemodels.InsulinDoseFromDB(row)
}

// GetHTMXResponse is the insulin dose's details, shown in a modal
func (s *InsulinDosesCollection) GetHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseModal(model)
}

// ///
// Search Models
// ///

func (s *InsulinDosesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.InsulinDose, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllInsulinDoses(ctx, siteapiutils.QueryToAllInsulinDoses(query))
}

func (s *InsulinDosesCollection) SearchHTMXResponse(models []*sitemodels.SiteInsulinDose, query *apiutils.Query) templ.Component {
	return site_data_templ.InsulinDosesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *InsulinDosesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	_, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).SoftDeleteInsulinDose, sitedb.SoftDeleteInsulinDoseParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *InsulinDosesCollection) DeleteHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
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
