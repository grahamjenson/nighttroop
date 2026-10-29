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
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ComponentEventsCollection struct{}

// ////
// Shared
// ////
func (s *ComponentEventsCollection) GetName() string {
	return "component_events"
}

// GetRoute is where the collection is in its URL
func (s *ComponentEventsCollection) GetRoute() string {
	return "component-events"
}

// ////
// CREATE
// ////
func (s *ComponentEventsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteComponentEvent) (*sitedb.ComponentEvent, error) {
	if herr := checkComponentEventType(model.EventType); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateComponentEvent(ctx, sitemodels.ComponentEventToCreate(model, siteContext))
}

func (s *ComponentEventsCollection) CreateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *ComponentEventsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteComponentEvent, model *sitemodels.SiteComponentEvent) (*sitedb.ComponentEvent, error) {
	if model.EventType != "" {
		if herr := checkComponentEventType(model.EventType); herr != nil {
			return nil, herr
		}
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateComponentEvent(ctx, sitemodels.ComponentEventToUpdate(model, siteContext))
}

func (s *ComponentEventsCollection) UpdateHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *ComponentEventsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.ComponentEvent, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetComponentEvent, sitedb.GetComponentEventParams{ID: id})
}

// ToModel is the site model of a row
func (s *ComponentEventsCollection) ToModel(row *sitedb.ComponentEvent, siteContext *httputils.SiteContext) *sitemodels.SiteComponentEvent {
	return sitemodels.ComponentEventFromDB(row)
}

// GetHTMXResponse is the component event's details, shown in a modal
func (s *ComponentEventsCollection) GetHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
	return site_data_templ.ComponentEventModal(model)
}

// ///
// Search Models
// ///

func (s *ComponentEventsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.ComponentEvent, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllComponentEvents(ctx, siteapiutils.QueryToAllComponentEvents(query))
}

func (s *ComponentEventsCollection) SearchHTMXResponse(models []*sitemodels.SiteComponentEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.ComponentEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ComponentEventsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	_, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).SoftDeleteComponentEvent, sitedb.SoftDeleteComponentEventParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *ComponentEventsCollection) DeleteHTMXResponse(model *sitemodels.SiteComponentEvent) templ.Component {
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
