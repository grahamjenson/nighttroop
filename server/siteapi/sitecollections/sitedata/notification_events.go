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

// defaultNotificationType is the type of a notification sent without one
const defaultNotificationType = "announcement"

type NotificationEventsCollection struct{}

// ////
// Shared
// ////
func (s *NotificationEventsCollection) GetName() string {
	return "notification_events"
}

// GetRoute is where the collection is in its URL
func (s *NotificationEventsCollection) GetRoute() string {
	return "notifications"
}

// ////
// CREATE
// ////
func (s *NotificationEventsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteNotificationEvent) (*sitedb.NotificationEvent, error) {
	if model.NotificationType == "" {
		model.NotificationType = defaultNotificationType
	}
	if model.Title == "" {
		return nil, apiutils.ErrConstraintCheckFailedWithHint("title is required")
	}
	if herr := checkNotificationEvent(model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateNotificationEvent(ctx, sitemodels.NotificationEventToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add a notification in
func (s *NotificationEventsCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.NotificationEventNew()
}

func (s *NotificationEventsCollection) CreateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *NotificationEventsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteNotificationEvent, model *sitemodels.SiteNotificationEvent) (*sitedb.NotificationEvent, error) {
	if herr := checkNotificationEvent(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateNotificationEvent(ctx, sitemodels.NotificationEventToUpdate(model, siteContext))
}

func (s *NotificationEventsCollection) UpdateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *NotificationEventsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.NotificationEvent, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetNotificationEvent(ctx, sitedb.GetNotificationEventParams{ID: id})
}

// ToModel is the site model of a row
func (s *NotificationEventsCollection) ToModel(row *sitedb.NotificationEvent, siteContext *httputils.SiteContext) *sitemodels.SiteNotificationEvent {
	return sitemodels.NotificationEventFromDB(row)
}

// GetHTMXResponse is the notification's details, shown in a modal
func (s *NotificationEventsCollection) GetHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventModal(model)
}

// ///
// Search Models
// ///

func (s *NotificationEventsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.NotificationEvent, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllNotificationEvents(ctx, siteapiutils.QueryToAllNotificationEvents(query))
}

func (s *NotificationEventsCollection) SearchHTMXResponse(models []*sitemodels.SiteNotificationEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.NotificationEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *NotificationEventsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteNotificationEvent)
}

func (s *NotificationEventsCollection) DeleteHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventDeleted()
}

///////
// UTIL METHODS
////////

// checkNotificationEvent keeps the title short enough to show and the severity within its scale
func checkNotificationEvent(model *sitemodels.SiteNotificationEvent) *apiutils.HTTPError {
	if len(model.Title) > 128 {
		return apiutils.ErrConstraintCheckFailedWithHint("title must be at most 128 characters")
	}
	return apiutils.CheckRange("severity", &model.Severity, 0, 5, "")
}
