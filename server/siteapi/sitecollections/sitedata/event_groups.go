package sitedata

import (
	"context"
	"net/http"
	"slices"

	"nighttroop.t1d.nz/utils/apiutils"
	utilerrors "nighttroop.t1d.nz/utils/errors"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// EventGroupsCollection shows, groups and deletes event groups. The Nightscout treatments API makes a group's events with
// it; the site groups events already made, each in at most one group.
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
// CREATE
// ////

// CreateRow groups the events picked, none of which can be in another group
func (s *EventGroupsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteEventGroup) (*sitedb.EventGroupsBundle, error) {
	if model.GroupType == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("errors.missing_field.group_type")
	}
	if herr := checkEventsPicked(model.GroupType, model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	group, err := q.CreateEventGroup(ctx, sitemodels.EventGroupToCreate(model, siteContext))
	if err != nil {
		return nil, groupingError(err)
	}
	return s.GetRow(ctx, collectionTx, group.ID, siteContext)
}

func (s *EventGroupsCollection) CreateHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow sets the group's events to those picked, leaving those taken out ungrouped, and its type, time, notes or
// reason when sent
func (s *EventGroupsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteEventGroup, model *sitemodels.SiteEventGroup) (*sitedb.EventGroupsBundle, error) {
	groupType := model.GroupType
	if groupType == "" {
		groupType = oldModel.GroupType
	}
	if herr := checkEventsPicked(groupType, model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	if _, err := q.RegroupEventGroup(ctx, sitemodels.EventGroupToRegroup(model, siteContext)); err != nil {
		return nil, groupingError(err)
	}
	return s.GetRow(ctx, collectionTx, model.ID, siteContext)
}

func (s *EventGroupsCollection) UpdateHTMXResponse(model *sitemodels.SiteEventGroup) templ.Component {
	return site_data_templ.EventGroupUpdated(model)
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

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *EventGroupsCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).EventGroupTotals)
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

///////
// UTIL METHODS
////////

// checkEventsPicked makes sure a group has an event, unless it is a change, which has none, and only of the kinds its
// type can have
func checkEventsPicked(groupType string, model *sitemodels.SiteEventGroup) *apiutils.HTTPError {
	kinds, ok := sitemodels.EventGroupTypeKinds[nightscoutmodels.TreatmentEventType(groupType)]
	if !ok {
		return &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "Unknown event group type " + groupType}
	}
	picked := false
	for field, id := range model.EventPicks() {
		if id != nil && !slices.Contains(kinds, field) {
			return &apiutils.HTTPError{Status: http.StatusBadRequest, Message: groupType + " cannot have " + field}
		}
		picked = picked || id != nil
	}
	if !picked && !slices.Contains(sitemodels.ChangeTypes, groupType) {
		return &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "Pick at least one event to group"}
	}
	return nil
}

// groupingError says which event cannot be grouped: one already in another group, as each event's column is unique,
// or one that is not there
func groupingError(err error) error {
	switch {
	case utilerrors.IsSQLUniqueViolationError(err):
		return &apiutils.HTTPError{Status: http.StatusConflict, Message: "An event picked is already in another event group"}
	case utilerrors.IsSQLConstraintForeignKey(err):
		return &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "An event picked does not exist"}
	}
	return err
}
