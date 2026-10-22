package sitedata

import (
	"context"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
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

// ToModel is the site model of a group and its events. The rows are bundles, as the Nightscout treatments show them,
// so a change is published as one.
func (s *EventGroupsStore) ToModel(row *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) *sitemodels.SiteEventGroup {
	return sitemodels.EventGroupFromDB(row, siteContext)
}

// ////
// Get Model
// /////
func (s *EventGroupsStore) GetRow(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.EventGroupsBundle, error) {
	row, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: id})
	if err != nil {
		return nil, err
	}
	bundle := sitedb.GetEventGroupToEventGroupsBundle(*row)
	return &bundle, nil
}

// GetHTMXResponse is the group's details and events, shown in a modal
func (s *EventGroupsStore) GetHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupModal(model)
}

// ///
// Search Models
// ///

func (s *EventGroupsStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query) ([]*sitedb.EventGroupsBundle, error) {
	rows, err := q.AllEventGroups(ctx, sitequery.QueryToAllEventGroups(query))
	if err != nil {
		return nil, err
	}
	bundles := make([]*sitedb.EventGroupsBundle, 0, len(rows))
	for _, row := range rows {
		bundle := sitedb.GetEventGroupToEventGroupsBundle(sitedb.AllEventGroupToGetEventGroup(*row))
		bundles = append(bundles, &bundle)
	}
	return bundles, nil
}

func (s *EventGroupsStore) SearchHTMXResponse(models []*sitemodels.SiteEventGroup, query *apiutils.Query) templ.Component {
	return site_data_templ.EventGroupsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes the group and its events, as the Nightscout API does, so syncing clients see the delete
func (s *EventGroupsStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	return sitedb.SoftDeleteEventGroupAndEvents(ctx, q, id, &siteContext.IdentityModel.Name)
}

func (s *EventGroupsStore) DeleteHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupDeleted()
}
