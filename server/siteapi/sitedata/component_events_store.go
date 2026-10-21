package sitedata

import (
	"context"
	"database/sql"
	"log/slog"
	"slices"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/httputils"
)

type ComponentEventsStore struct{}

// ////
// Shared
// ////
func (s *ComponentEventsStore) GetName() string {
	return "component_events"
}

// ////
// CREATE
// ////
func (s *ComponentEventsStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteComponentEvent) (bool, *apiutils.HTTPError) {
	if herr := checkComponentEventType(model.EventType); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateComponentEvent(ctx, sitemodels.ComponentEventToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "component event")
	}
	tx.Created(created)
	*model = *sitemodels.ComponentEventFromDB(created)
	return true, nil
}

func (s *ComponentEventsStore) CreateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes the fields sent; anything not sent stays as it is
func (s *ComponentEventsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteComponentEvent, model *sitemodels.SiteComponentEvent) *apiutils.HTTPError {
	if model.EventType != "" {
		if herr := checkComponentEventType(model.EventType); herr != nil {
			return herr
		}
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateComponentEvent(ctx, sitemodels.ComponentEventToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "component event")
	}
	tx.Updated(updated)
	*model = *sitemodels.ComponentEventFromDB(updated)
	return nil
}

func (s *ComponentEventsStore) UpdateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *ComponentEventsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteComponentEvent, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	event, err := q.GetComponentEvent(ctx, sitedb.GetComponentEventParams{ID: id})
	return apiutils.Found(event, err, "component event", sitemodels.ComponentEventFromDB)
}

// GetHTMXResponse is the component event's details, shown in a modal
func (s *ComponentEventsStore) GetHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventModal(model)
}

// ///
// Search Models
// ///

func (s *ComponentEventsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteComponentEvent, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	events, err := q.AllComponentEvents(ctx, sitequery.QueryToAllComponentEvents(query))
	return apiutils.Searched(events, err, "component events", sitemodels.ComponentEventFromDB)
}

func (s *ComponentEventsStore) SearchHTMXResponse(models []*sitemodels.SiteComponentEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.ComponentEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ComponentEventsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteComponentEvent, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	tx.Deleted(q.GetComponentEvent(ctx, sitedb.GetComponentEventParams{ID: model.ID}))
	_, err = q.SoftDeleteComponentEvent(ctx, sitedb.SoftDeleteComponentEventParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting component event", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *ComponentEventsStore) DeleteHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventDeleted()
}

///////
// UTIL METHODS
////////

// checkComponentEventType allows only the Nightscout device events that change a sensor, site or pump part
func checkComponentEventType(eventType string) *apiutils.HTTPError {
	if !slices.Contains(sitemodels.ComponentEventTypes, eventType) {
		return apiutils.ErrConstraintCheckFailedWithHint("event_type must be one of " + strings.Join(sitemodels.ComponentEventTypes, ", "))
	}
	return nil
}
