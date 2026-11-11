package sitedb

import (
	"context"
	"database/sql"
)

// SoftDeleteEventGroupAndEvents soft deletes a group and every event in it, so none of its events are left behind.
// It is sql.ErrNoRows when there is no such group, or it is already deleted.
func SoftDeleteEventGroupAndEvents(ctx context.Context, q *Queries, id string, updatedBy *string) error {
	eventGroup, err := q.SoftDeleteEventGroup(ctx, SoftDeleteEventGroupParams{ID: id, UpdatedByIdentityName: updatedBy})
	if err != nil {
		return err
	}
	if eventGroup == nil {
		return sql.ErrNoRows
	}
	deletes := []struct {
		id     *string
		delete func(id string) (sql.Result, error)
	}{
		{eventGroup.GlucoseReadingID, func(id string) (sql.Result, error) {
			return q.SoftDeleteGlucoseReading(ctx, SoftDeleteGlucoseReadingParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.InsulinDoseID, func(id string) (sql.Result, error) {
			return q.SoftDeleteInsulinDose(ctx, SoftDeleteInsulinDoseParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.FoodIntakeID, func(id string) (sql.Result, error) {
			return q.SoftDeleteFoodIntake(ctx, SoftDeleteFoodIntakeParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.TempBasalID, func(id string) (sql.Result, error) {
			return q.SoftDeleteTempBasal(ctx, SoftDeleteTempBasalParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.TempTargetID, func(id string) (sql.Result, error) {
			return q.SoftDeleteTempTarget(ctx, SoftDeleteTempTargetParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.ProfileSwitchID, func(id string) (sql.Result, error) {
			return q.SoftDeleteProfileSwitch(ctx, SoftDeleteProfileSwitchParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.ActivityID, func(id string) (sql.Result, error) {
			return q.SoftDeleteActivity(ctx, SoftDeleteActivityParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
		{eventGroup.NotificationEventID, func(id string) (sql.Result, error) {
			return q.SoftDeleteNotificationEvent(ctx, SoftDeleteNotificationEventParams{ID: id, UpdatedByIdentityName: updatedBy})
		}},
	}
	for _, d := range deletes {
		if d.id == nil {
			continue
		}
		if _, err := d.delete(*d.id); err != nil {
			return err
		}
	}
	return nil
}
