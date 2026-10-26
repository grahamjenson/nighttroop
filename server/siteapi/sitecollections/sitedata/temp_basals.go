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

// maxTempBasalMinutes is a day, the longest pumps set a temp basal for
const maxTempBasalMinutes = 24 * 60

// maxTempBasalUnitsPerHour is far above any pump's maximum basal rate
const maxTempBasalUnitsPerHour = 50

type TempBasalsCollection struct{}

// ////
// Shared
// ////
func (s *TempBasalsCollection) GetName() string {
	return "temp_basals"
}

// GetRoute is where the collection is in its URL
func (s *TempBasalsCollection) GetRoute() string {
	return "temp_basals"
}

// ////
// CREATE
// ////
func (s *TempBasalsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteTempBasal) (*sitedb.TempBasal, error) {
	if herr := checkTempBasal(model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateTempBasal(ctx, sitemodels.TempBasalToCreate(model, siteContext))
}

func (s *TempBasalsCollection) CreateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a temp basal's rate, duration, notes or reason; anything not sent stays as it is
func (s *TempBasalsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteTempBasal, model *sitemodels.SiteTempBasal) (*sitedb.TempBasal, error) {
	// Check the temp basal as it will be once the fields sent are changed
	merged := *oldModel
	if model.DurationMinutes != 0 {
		merged.DurationMinutes = model.DurationMinutes
	}
	if model.Absolute != nil {
		merged.Absolute = model.Absolute
	}
	if herr := checkTempBasal(&merged); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateTempBasal(ctx, sitemodels.TempBasalToUpdate(model, siteContext))
}

func (s *TempBasalsCollection) UpdateHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalUpdated(model)
}

// ////
// Get Model
// /////
func (s *TempBasalsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.TempBasal, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetTempBasal(ctx, sitedb.GetTempBasalParams{ID: id})
}

// ToModel is the site model of a row
func (s *TempBasalsCollection) ToModel(row *sitedb.TempBasal, siteContext *httputils.SiteContext) *sitemodels.SiteTempBasal {
	return sitemodels.TempBasalFromDB(row)
}

// GetHTMXResponse is the temp basal's details, shown in a modal
func (s *TempBasalsCollection) GetHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalModal(model)
}

// ///
// Search Models
// ///

func (s *TempBasalsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.TempBasal, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllTempBasals(ctx, siteapiutils.QueryToAllTempBasals(query))
}

func (s *TempBasalsCollection) SearchHTMXResponse(models []*sitemodels.SiteTempBasal, query *apiutils.Query) templ.Component {
	return site_data_templ.TempBasalsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *TempBasalsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.SoftDeleteTempBasal(ctx, sitedb.SoftDeleteTempBasalParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *TempBasalsCollection) DeleteHTMXResponse(model *sitemodels.SiteTempBasal) templ.Component {
	return site_data_templ.TempBasalDeleted()
}

///////
// UTIL METHODS
////////

// checkTempBasal keeps the rate to what a pump can deliver, and the duration to a day
func checkTempBasal(model *sitemodels.SiteTempBasal) *apiutils.HTTPError {
	if model.Absolute == nil || *model.Absolute < 0 || *model.Absolute > maxTempBasalUnitsPerHour {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("absolute must be from 0 to %d U/hr", maxTempBasalUnitsPerHour))
	}
	if model.DurationMinutes <= 0 || model.DurationMinutes > maxTempBasalMinutes {
		return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("duration must be more than 0 and at most %d minutes", maxTempBasalMinutes))
	}
	return nil
}
