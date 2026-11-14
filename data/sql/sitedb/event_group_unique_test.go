package sitedb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// Each event is in at most one event group, while any number of groups have none of an event, as the unique indexes
// hold only the groups that have it
func TestEventGroupEventsUnique(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx, by := site.SiteDB.Queries(), context.Background(), "test"
	group := func(id string, set func(*sitedb.CreateEventGroupParams)) error {
		params := sitedb.CreateEventGroupParams{ID: id, GroupType: "note", OccurredAt: 1, CreatedByIdentityName: by}
		set(&params)
		_, err := q.CreateEventGroup(ctx, params)
		return err
	}

	for _, id := range []string{"none-1", "none-2", "none-3"} {
		require.NoError(t, group(id, func(*sitedb.CreateEventGroupParams) {}), "a group with no event")
	}

	tests := []struct {
		name   string
		create func(id string) error
		set    func(p *sitedb.CreateEventGroupParams, id *string)
	}{
		{"glucose_reading_id", func(id string) error {
			_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 100, ReadingType: "mbg", OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.GlucoseReadingID = id }},
		{"insulin_dose_id", func(id string) error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.InsulinDoseID = id }},
		{"food_intake_id", func(id string) error {
			_, err := q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: id, OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.FoodIntakeID = id }},
		{"temp_basal_id", func(id string) error {
			_, err := q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: id, DurationMinutes: 30, Rate: 1, OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.TempBasalID = id }},
		{"temp_target_id", func(id string) error {
			_, err := q.CreateTempTarget(ctx, sitedb.CreateTempTargetParams{ID: id, DurationMinutes: 30, HighTargetMgdl: 140, LowTargetMgdl: 120, OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.TempTargetID = id }},
		{"profile_switch_id", func(id string) error {
			_, err := q.CreateProfileSwitch(ctx, sitedb.CreateProfileSwitchParams{ID: id, Name: id, OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.ProfileSwitchID = id }},
		{"activity_id", func(id string) error {
			_, err := q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: id, ActivityType: "exercise", OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.ActivityID = id }},
		{"notification_event_id", func(id string) error {
			_, err := q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: id, Title: "title", NotificationType: "note", OccurredAt: 1, CreatedByIdentityName: by})
			return err
		}, func(p *sitedb.CreateEventGroupParams, id *string) { p.NotificationEventID = id }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := "event-" + tt.name
			require.NoError(t, tt.create(event))
			with := func(p *sitedb.CreateEventGroupParams) { tt.set(p, &event) }
			require.NoError(t, group("first-"+tt.name, with))
			err := group("second-"+tt.name, with)
			require.Error(t, err, "a second group with the same event")
			require.True(t, utilerrors.IsSQLUniqueViolationError(err), err.Error())
			require.Contains(t, err.Error(), "event_groups."+tt.name)
		})
	}
}
