package sitedata

import (
	"context"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// EventGroupsCollection shows and deletes event groups. They are created and changed through the Nightscout treatments API,
// which knows how to make the events in each type of group.
type EventGroupsCollection struct{}

// ////
// Shared
// ////
func (s *EventGroupsCollection) GetName() string {
	return "event_groups"
}

// GetRoute is where the collection is in its URL
func (s *EventGroupsCollection) GetRoute() string {
	return "event-groups"
}

// ToModel is the site model of a group and its events. The rows are bundles, as the Nightscout treatments show them,
// so a change is published as one.
func (s *EventGroupsCollection) ToModel(row *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) *sitemodels.SiteEventGroup {
	return sitemodels.EventGroupFromDB(row, siteContext)
}

// ////
// Get Model
// /////
func (s *EventGroupsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.EventGroupsBundle, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	row, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: id})
	if err != nil {
		return nil, err
	}
	bundle := sitedb.GetEventGroupToEventGroupsBundle(*row)
	return &bundle, nil
}

// GetHTMXResponse is the group's details and events, shown in a modal
func (s *EventGroupsCollection) GetHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupModal(model)
}

// ///
// Search Models
// ///

func (s *EventGroupsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.EventGroupsBundle, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	rows, err := q.AllEventGroups(ctx, siteapiutils.QueryToAllEventGroups(query))
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

func (s *EventGroupsCollection) SearchHTMXResponse(models []*sitemodels.SiteEventGroup, query *apiutils.Query) templ.Component {
	return site_data_templ.EventGroupsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes the group and its events, as the Nightscout API does, so syncing clients see the delete
func (s *EventGroupsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	return sitedb.SoftDeleteEventGroupAndEvents(ctx, q, id, &siteContext.IdentityModel.Name)
}

func (s *EventGroupsCollection) DeleteHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupDeleted()
}
