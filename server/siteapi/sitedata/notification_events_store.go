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

// defaultNotificationType is the type of a notification sent without one
const defaultNotificationType = "announcement"

type NotificationEventsStore struct{}

// ////
// Shared
// ////
func (s *NotificationEventsStore) GetName() string {
	return "notification_events"
}

// ////
// CREATE
// ////
func (s *NotificationEventsStore) CreateRow(ctx context.Context, q *sitedb.Queries, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteNotificationEvent) (*sitedb.NotificationEvent, error) {
	if model.NotificationType == "" {
		model.NotificationType = defaultNotificationType
	}
	if model.Title == "" {
		return nil, apiutils.ErrConstraintCheckFailedWithHint("title is required")
	}
	if herr := checkNotificationEvent(model); herr != nil {
		return nil, herr
	}
	return q.CreateNotificationEvent(ctx, sitemodels.NotificationEventToCreate(model, siteContext))
}

func (s *NotificationEventsStore) CreateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *NotificationEventsStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteNotificationEvent, model *sitemodels.SiteNotificationEvent) (*sitedb.NotificationEvent, error) {
	if herr := checkNotificationEvent(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	return q.UpdateNotificationEvent(ctx, sitemodels.NotificationEventToUpdate(model, siteContext))
}

func (s *NotificationEventsStore) UpdateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *NotificationEventsStore) GetRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) (*sitedb.NotificationEvent, error) {
	return q.GetNotificationEvent(ctx, sitedb.GetNotificationEventParams{ID: id})
}

// ToModel is the site model of a row
func (s *NotificationEventsStore) ToModel(row *sitedb.NotificationEvent, siteContext *httputils.SiteContext) *sitemodels.SiteNotificationEvent {
	return sitemodels.NotificationEventFromDB(row)
}

// GetHTMXResponse is the notification's details, shown in a modal
func (s *NotificationEventsStore) GetHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventModal(model)
}

// ///
// Search Models
// ///

func (s *NotificationEventsStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.NotificationEvent, error) {
	return q.AllNotificationEvents(ctx, sitequery.QueryToAllNotificationEvents(query))
}

func (s *NotificationEventsStore) SearchHTMXResponse(models []*sitemodels.SiteNotificationEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.NotificationEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *NotificationEventsStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteNotificationEvent(ctx, sitedb.SoftDeleteNotificationEventParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *NotificationEventsStore) DeleteHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
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
	if model.Severity < 0 || model.Severity > 5 {
		return apiutils.ErrConstraintCheckFailedWithHint("severity must be from 0 to 5")
	}
	return nil
}
