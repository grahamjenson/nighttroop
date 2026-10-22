package sitedata

import (
	"context"
	"slices"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
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
func (s *ComponentEventsStore) CreateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, model *sitemodels.SiteComponentEvent) (*sitedb.ComponentEvent, error) {
	if herr := checkComponentEventType(model.EventType); herr != nil {
		return nil, herr
	}
	return q.CreateComponentEvent(ctx, sitemodels.ComponentEventToCreate(model, siteContext))
}

func (s *ComponentEventsStore) CreateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *ComponentEventsStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteComponentEvent, model *sitemodels.SiteComponentEvent) (*sitedb.ComponentEvent, error) {
	if model.EventType != "" {
		if herr := checkComponentEventType(model.EventType); herr != nil {
			return nil, herr
		}
	}
	model.ID = oldModel.ID
	return q.UpdateComponentEvent(ctx, sitemodels.ComponentEventToUpdate(model, siteContext))
}

func (s *ComponentEventsStore) UpdateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *ComponentEventsStore) GetRow(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.ComponentEvent, error) {
	return q.GetComponentEvent(ctx, sitedb.GetComponentEventParams{ID: id})
}

// ToModel is the site model of a row
func (s *ComponentEventsStore) ToModel(row *sitedb.ComponentEvent, siteContext *httputils.SiteContext) *sitemodels.SiteComponentEvent {
	return sitemodels.ComponentEventFromDB(row)
}

// GetHTMXResponse is the component event's details, shown in a modal
func (s *ComponentEventsStore) GetHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventModal(model)
}

// ///
// Search Models
// ///

func (s *ComponentEventsStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query) ([]*sitedb.ComponentEvent, error) {
	return q.AllComponentEvents(ctx, sitequery.QueryToAllComponentEvents(query))
}

func (s *ComponentEventsStore) SearchHTMXResponse(models []*sitemodels.SiteComponentEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.ComponentEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ComponentEventsStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteComponentEvent(ctx, sitedb.SoftDeleteComponentEventParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
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
