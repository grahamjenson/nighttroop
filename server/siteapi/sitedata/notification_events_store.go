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
func (s *NotificationEventsStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteNotificationEvent) (bool, *apiutils.HTTPError) {
	if model.NotificationType == "" {
		model.NotificationType = defaultNotificationType
	}
	if model.Title == "" {
		return false, apiutils.ErrConstraintCheckFailedWithHint("title is required")
	}
	if herr := checkNotificationEvent(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateNotificationEvent(ctx, sitemodels.NotificationEventToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "notification")
	}
	*model = *sitemodels.NotificationEventFromDB(created)
	return true, nil
}

func (s *NotificationEventsStore) CreateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes the fields sent; anything not sent stays as it is
func (s *NotificationEventsStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteNotificationEvent, model *sitemodels.SiteNotificationEvent) *apiutils.HTTPError {
	if herr := checkNotificationEvent(model); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateNotificationEvent(ctx, sitemodels.NotificationEventToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "notification")
	}
	*model = *sitemodels.NotificationEventFromDB(updated)
	return nil
}

func (s *NotificationEventsStore) UpdateHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventUpdated(model)
}

// ////
// Get Model
// /////
func (s *NotificationEventsStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteNotificationEvent, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	event, err := q.GetNotificationEvent(ctx, sitedb.GetNotificationEventParams{ID: id})
	return apiutils.Found(event, err, "notification", sitemodels.NotificationEventFromDB)
}

// GetHTMXResponse is the notification's details, shown in a modal
func (s *NotificationEventsStore) GetHTMXResponse(model *sitemodels.SiteNotificationEvent) templ.Component {
	return site_data_templ.NotificationEventModal(model)
}

// ///
// Search Models
// ///

func (s *NotificationEventsStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteNotificationEvent, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	events, err := q.AllNotificationEvents(ctx, sitequery.QueryToAllNotificationEvents(query))
	return apiutils.Searched(events, err, "notifications", sitemodels.NotificationEventFromDB)
}

func (s *NotificationEventsStore) SearchHTMXResponse(models []*sitemodels.SiteNotificationEvent, query *apiutils.Query) templ.Component {
	return site_data_templ.NotificationEventsTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *NotificationEventsStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteNotificationEvent, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteNotificationEvent(ctx, sitedb.SoftDeleteNotificationEventParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting notification", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
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
