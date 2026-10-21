package sitedata

import (
	"context"
	"database/sql"
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

// EventGroupsStore shows and deletes event groups. They are created and changed through the Nightscout treatments API,
// which knows how to make the events in each type of group.
type EventGroupsStore struct{}

// ////
// Shared
// ////
func (s *EventGroupsStore) GetName() string {
	return "event_groups"
}

// eventGroupFromRow is the site model of a group and its events, read with GetEventGroup
func eventGroupFromRow(siteContext *httputils.SiteContext) func(*sitedb.GetEventGroupRow) *sitemodels.SiteEventGroup {
	return func(row *sitedb.GetEventGroupRow) *sitemodels.SiteEventGroup {
		bundle := sitedb.GetEventGroupToEventGroupsBundle(*row)
		return sitemodels.EventGroupFromDB(&bundle, siteContext)
	}
}

// ////
// Get Model
// /////
func (s *EventGroupsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteEventGroup, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	eventGroup, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: id})
	return apiutils.Found(eventGroup, err, "event group", eventGroupFromRow(siteContext))
}

// GetHTMXResponse is the group's details and events, shown in a modal
func (s *EventGroupsStore) GetHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupModal(model)
}

// ///
// Search Models
// ///

func (s *EventGroupsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteEventGroup, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	rows, err := q.AllEventGroups(ctx, sitequery.QueryToAllEventGroups(query))
	return apiutils.Searched(rows, err, "event groups", func(row *sitedb.AllEventGroupsRow) *sitemodels.SiteEventGroup {
		getRow := sitedb.AllEventGroupToGetEventGroup(*row)
		return eventGroupFromRow(siteContext)(&getRow)
	})
}

func (s *EventGroupsStore) SearchHTMXResponse(models []*sitemodels.SiteEventGroup, query *apiutils.Query) templ.Component {
	return site_data_templ.EventGroupsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes the group and its events, as the Nightscout API does, so syncing clients see the delete
func (s *EventGroupsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteEventGroup, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	err = sitedb.SoftDeleteEventGroupAndEvents(ctx, q, model.ID, &siteContext.IdentityModel.Name)
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		slog.Error("Error deleting event group", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *EventGroupsStore) DeleteHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupDeleted()
}
