package sitedata

import (
	"cmp"
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

// maxInsulinDoseUnits is far above any single dose a pump or pen gives; as much again below 0 takes back a dose
// entered by mistake, from insulin on board
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
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateInsulinDose, sitemodels.InsulinDoseToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add an insulin dose in
func (s *InsulinDosesCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.InsulinDoseNew()
}

func (s *InsulinDosesCollection) CreateHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a dose's units, type, notes or reason; anything not sent stays as it is
func (s *InsulinDosesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteInsulinDose, model *sitemodels.SiteInsulinDose) (*sitedb.InsulinDose, error) {
	// Check the dose as it will be once the fields sent are changed
	merged := *model
	if merged.Dose == 0 {
		merged.Dose = oldModel.Dose
	}
	if herr := checkInsulinDose(&merged); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateInsulinDose, sitemodels.InsulinDoseToUpdate(model, siteContext))
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
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllInsulinDoses, siteapiutils.QueryToAllInsulinDoses(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *InsulinDosesCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).InsulinDoseTotals)
}

func (s *InsulinDosesCollection) SearchHTMXResponse(models []*sitemodels.SiteInsulinDose, query *apiutils.Query) templ.Component {
	return site_data_templ.InsulinDosesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *InsulinDosesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteInsulinDose)
}

func (s *InsulinDosesCollection) DeleteHTMXResponse(model *sitemodels.SiteInsulinDose) templ.Component {
	return site_data_templ.InsulinDoseDeleted()
}

///////
// UTIL METHODS
////////

func checkInsulinDose(model *sitemodels.SiteInsulinDose) *apiutils.HTTPError {
	if model.Dose == 0 || model.Dose < -maxInsulinDoseUnits || model.Dose > maxInsulinDoseUnits {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("dose must not be 0, and must be from -%d to %d units", maxInsulinDoseUnits, maxInsulinDoseUnits))
	}
	return cmp.Or(apiutils.CheckRange("split_now", model.SplitNow, 0, 1, ""), apiutils.CheckRange("split_ext", model.SplitExt, 0, 1, ""))
}
