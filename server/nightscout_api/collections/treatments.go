package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	converter func(c *site.EventGroupsBundle, raw bool) nsutils.Record
}

func NewTreatmentCollection() TreatmentCollection {
	return TreatmentCollection{
		converter: nsutils.DBModelToRecord(nsutils.Treatments, DBModelToTreatment),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c TreatmentCollection) Name() string {
	return "treatments"
}

func (c TreatmentCollection) Convert(dr *site.EventGroupsBundle, raw bool) nsutils.Record {
	return c.converter(dr, raw)
}

func (c TreatmentCollection) Validate(nsModel nightscout.Treatment, dbModel *site.EventGroupsBundle) *nsutils.HTTPError {
	// Validate the input
	if nsModel.EventType == nil || *nsModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}

	// Validate the change
	if dbModel == nil {
		return nil
	}

	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("cannot change event_type attribute on treatment", "existingEventType", dbModel.EventGroup.GroupType, "newEventType", nsModel.EventType.String())
		return nsutils.ErrCannotChangeEventType
	}

	return nil
}

func (c TreatmentCollection) ValidatePatch(nsModel nightscout.Treatment, dbModel *site.EventGroupsBundle) *nsutils.HTTPError {
	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}
	return nil
}

func (c TreatmentCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.EventGroupsBundle, error) {

	tebg := TreatmentEventGroupBundle(ctx, nsModel)
	err := createUpdateEventGroup(ctx, q, &tebg, cc)
	if err != nil {
		return nil, err
	}

	createEventGroupParams := nightscout.EventGroupCreate(tebg.EventGroup, cc)

	if overrideIdentifier != nil {
		createEventGroupParams.ID = *overrideIdentifier
	}
	newEventGroup, err := q.CreateEventGroup(ctx, createEventGroupParams)
	if err != nil {
		slog.Error("failed to create event group in database", "error", err, "createParams", tebg)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", tebg)
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	tebg.EventGroup = *newEventGroup

	return &tebg, nil
}

func (c TreatmentCollection) GetDuplicateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment) (*site.EventGroupsBundle, error) {
	eventGroups, rerr := q.SearchEventGroups(ctx, site.SearchEventGroupsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		TypeEq:   nsModel.EventType.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(eventGroups) > 0 {
		egb := site.GetEventGroupToEventGroupsBundle(site.SearchEventGroupToGetEventGroup(*eventGroups[0]))
		return &egb, nil
	}
	return nil, nil
}

func (c TreatmentCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, site.GetEventGroupParams{ID: identifier})
	if err != nil {
		return nil, err
	}

	egb := site.GetEventGroupToEventGroupsBundle(*dbModel)

	return &egb, nil
}

func (c TreatmentCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, identifier string, dbModel *site.EventGroupsBundle, cc nightscout.ConverterContext) (*site.EventGroupsBundle, error) {
	// This will be a partial update

	nsModel.EventType = ptr(nightscout.TreatmentEventType(dbModel.EventGroup.GroupType)) // Ensure the event type is set to the existing value

	updateBundle := TreatmentEventGroupBundle(ctx, nsModel)

	MergeUpdateBundle(dbModel, updateBundle)

	// Do something with the update bundle to merge it into the dbModel

	err := createUpdateEventGroup(ctx, q, dbModel, cc)
	if err != nil {
		return nil, err
	}

	updateEventGroupParams := nightscout.EventGroupUpdate(dbModel.EventGroup, cc)
	newEventGroup, err := q.UpdateEventGroup(ctx, updateEventGroupParams)
	if err != nil {
		slog.Error("failed to update event group in database", "error", err)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.Error("failed to update event group in database, EventGroup is nil")
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	dbModel.EventGroup = *newEventGroup

	return dbModel, nil
}

func (c TreatmentCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return hardDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return softDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) HardDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, site.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString})
	if err != nil {
		return nil, err
	}

	for _, eg := range eventGroups {
		if eg != nil {
			if _, err := hardDeleteEventGroup(ctx, q, eg.ID); err != nil {
				return nil, err
			}

		}
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

func (c TreatmentCollection) SoftDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, site.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString})
	if err != nil {
		return nil, err
	}

	for _, eg := range eventGroups {
		if eg != nil {
			if _, err := softDeleteEventGroup(ctx, q, eg.ID); err != nil {
				return nil, err
			}

		}
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c TreatmentCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c TreatmentCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw)
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	treatments, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(treatments, c.converter, query.Raw), nil
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

// Replace will replace (delete, then insert) if it exists, or create it if it doesnt
func (c TreatmentCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}

func (c TreatmentCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	// We will always repalce because it is way harder otherwise
	return MetaPatch(c, ctx, siteContext, identifier, patch, raw)
}

func EventGroupSearch(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*site.EventGroupsBundle, *nsutils.HTTPError) {
	eventGroups, rerr := siteContext.SiteDB.Queries.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if rerr != nil {
		slog.Error("failed to query event groups from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	res := []*site.EventGroupsBundle{}
	for _, eg := range eventGroups {
		if eg != nil {
			egb := site.GetEventGroupToEventGroupsBundle(site.SearchEventGroupToGetEventGroup(*eg))
			res = append(res, &egb)
		}
	}

	return res, nil

}

// //////////////////////
// Private Methods
// ///////////////////////

// createUpdateEventGroup takes an event bundle and tries to sync it to the DB
// 1. Delete if the EventGroup has an model ID but no model
// 2. Create if the EventGroup has no ID but has a model
// 3. Update if the EventGroup has an ID and a model
//
// TODO: This means events are always updated, we should selectivly update.
func createUpdateEventGroup(ctx context.Context, q *site.Queries, bundle *site.EventGroupsBundle, cc nightscout.ConverterContext) error {

	// GlucoseReading
	if bundle.DeleteGlucoseReading() {
		_, err := q.HardDeleteGlucoseReading(ctx, site.HardDeleteGlucoseReadingParams{ID: *bundle.EventGroup.GlucoseReadingID})
		if err != nil {
			slog.Error("failed to delete glucose reading from database", "error", err)
			return err
		}
		bundle.EventGroup.GlucoseReadingID = nil
	} else if bundle.CreateGlucoseReading() {
		dbModel, err := q.CreateGlucoseReading(ctx, nightscout.GlucoseReadingCreate(*bundle.GlucoseReading, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create glucose reading in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.GlucoseReadingID = &dbModel.ID
		bundle.GlucoseReading = dbModel
	} else if bundle.UpdateGlucoseReading() {
		dbModel, err := q.UpdateGlucoseReading(ctx, nightscout.GlucoseReadingUpdate(*bundle.GlucoseReading, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update glucose reading in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.GlucoseReadingID = &dbModel.ID
		bundle.GlucoseReading = dbModel
	}

	// InsulinDose
	if bundle.DeleteInsulinDose() {
		_, err := q.HardDeleteInsulinDose(ctx, site.HardDeleteInsulinDoseParams{ID: *bundle.EventGroup.InsulinDoseID})
		if err != nil {
			slog.Error("failed to delete insulin dose from database", "error", err)
			return err
		}
		bundle.EventGroup.InsulinDoseID = nil
	} else if bundle.CreateInsulinDose() {
		dbModel, err := q.CreateInsulinDose(ctx, nightscout.InsulinDoseCreate(*bundle.InsulinDose, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create insulin dose in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.InsulinDoseID = &dbModel.ID
		bundle.InsulinDose = dbModel
	} else if bundle.UpdateInsulinDose() {
		dbModel, err := q.UpdateInsulinDose(ctx, nightscout.InsulinDoseUpdate(*bundle.InsulinDose, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update insulin dose in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.InsulinDoseID = &dbModel.ID
		bundle.InsulinDose = dbModel
	}

	if bundle.DeleteFoodIntake() {
		_, err := q.HardDeleteFoodIntake(ctx, site.HardDeleteFoodIntakeParams{ID: *bundle.EventGroup.FoodIntakeID})
		if err != nil {
			slog.Error("failed to delete food intake from database", "error", err)
			return err
		}
		bundle.EventGroup.FoodIntakeID = nil
	} else if bundle.CreateFoodIntake() {
		dbModel, err := q.CreateFoodIntake(ctx, nightscout.FoodIntakeCreate(*bundle.FoodIntake, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create food intake in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.FoodIntakeID = &dbModel.ID
		bundle.FoodIntake = dbModel
	} else if bundle.UpdateFoodIntake() {
		dbModel, err := q.UpdateFoodIntake(ctx, nightscout.FoodIntakeUpdate(*bundle.FoodIntake, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update food intake in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.FoodIntakeID = &dbModel.ID
		bundle.FoodIntake = dbModel
	}

	// TempBasal
	if bundle.DeleteTempBasal() {
		_, err := q.HardDeleteTempBasal(ctx, site.HardDeleteTempBasalParams{ID: *bundle.EventGroup.TempBasalID})
		if err != nil {
			slog.Error("failed to delete temp basal from database", "error", err)
			return err
		}
		bundle.EventGroup.TempBasalID = nil
	} else if bundle.CreateTempBasal() {
		dbModel, err := q.CreateTempBasal(ctx, nightscout.TempBasalCreate(*bundle.TempBasal, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp basal in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempBasalID = &dbModel.ID
		bundle.TempBasal = dbModel
	} else if bundle.UpdateTempBasal() {
		dbModel, err := q.UpdateTempBasal(ctx, nightscout.TempBasalUpdate(*bundle.TempBasal, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update temp basal in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempBasalID = &dbModel.ID
		bundle.TempBasal = dbModel
	}

	// TempTarget
	if bundle.DeleteTempTarget() {
		_, err := q.HardDeleteTempTarget(ctx, site.HardDeleteTempTargetParams{ID: *bundle.EventGroup.TempTargetID})
		if err != nil {
			slog.Error("failed to delete temp target from database", "error", err)
			return err
		}
		bundle.EventGroup.TempTargetID = nil
	} else if bundle.CreateTempTarget() {
		dbModel, err := q.CreateTempTarget(ctx, nightscout.TempTargetCreate(*bundle.TempTarget, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp target in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempTargetID = &dbModel.ID
		bundle.TempTarget = dbModel
	} else if bundle.UpdateTempTarget() {
		dbModel, err := q.UpdateTempTarget(ctx, nightscout.TempTargetUpdate(*bundle.TempTarget, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update temp target in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempTargetID = &dbModel.ID
		bundle.TempTarget = dbModel
	}

	// ProfileSwitch
	if bundle.DeleteProfileSwitch() {
		_, err := q.HardDeleteProfileSwitch(ctx, site.HardDeleteProfileSwitchParams{ID: *bundle.EventGroup.ProfileSwitchID})
		if err != nil {
			slog.Error("failed to delete profile switch from database", "error", err)
			return err
		}
		bundle.EventGroup.ProfileSwitchID = nil
	} else if bundle.CreateProfileSwitch() {
		dbModel, err := q.CreateProfileSwitch(ctx, nightscout.ProfileSwitchCreate(*bundle.ProfileSwitch, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create profile switch in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ProfileSwitchID = &dbModel.ID
		bundle.ProfileSwitch = dbModel
	} else if bundle.UpdateProfileSwitch() {
		dbModel, err := q.UpdateProfileSwitch(ctx, nightscout.ProfileSwitchUpdate(*bundle.ProfileSwitch, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update profile switch in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ProfileSwitchID = &dbModel.ID
		bundle.ProfileSwitch = dbModel
	}

	// ComponentEvent
	if bundle.DeleteComponentEvent() {
		_, err := q.HardDeleteComponentEvent(ctx, site.HardDeleteComponentEventParams{ID: *bundle.EventGroup.ComponentEventID})
		if err != nil {
			slog.Error("failed to delete component event from database", "error", err)
			return err
		}
		bundle.EventGroup.ComponentEventID = nil
	} else if bundle.CreateComponentEvent() {
		dbModel, err := q.CreateComponentEvent(ctx, nightscout.ComponentEventCreate(*bundle.ComponentEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create component event in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ComponentEventID = &dbModel.ID
		bundle.ComponentEvent = dbModel
	} else if bundle.UpdateComponentEvent() {
		dbModel, err := q.UpdateComponentEvent(ctx, nightscout.ComponentEventUpdate(*bundle.ComponentEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update component event in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ComponentEventID = &dbModel.ID
		bundle.ComponentEvent = dbModel
	}

	// Activity
	if bundle.DeleteActivity() {
		_, err := q.HardDeleteActivity(ctx, site.HardDeleteActivityParams{ID: *bundle.EventGroup.ActivityID})
		if err != nil {
			slog.Error("failed to delete activity from database", "error", err)
			return err

		}
		bundle.EventGroup.ActivityID = nil
	} else if bundle.CreateActivity() {
		dbModel, err := q.CreateActivity(ctx, nightscout.ActivityCreate(*bundle.Activity, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create activity in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ActivityID = &dbModel.ID
		bundle.Activity = dbModel
	} else if bundle.UpdateActivity() {
		dbModel, err := q.UpdateActivity(ctx, nightscout.ActivityUpdate(*bundle.Activity, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update activity in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ActivityID = &dbModel.ID
		bundle.Activity = dbModel
	}

	// note
	if bundle.DeleteNotificationEvent() {
		_, err := q.HardDeleteNotificationEvent(ctx, site.HardDeleteNotificationEventParams{ID: *bundle.EventGroup.NotificationEventID})
		if err != nil {
			slog.Error("failed to delete note from database", "error", err)
			return err
		}
		bundle.EventGroup.NotificationEventID = nil
	} else if bundle.CreateNotificationEvent() {
		dbModel, err := q.CreateNotificationEvent(ctx, nightscout.NotificationEventCreate(*bundle.NotificationEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.NotificationEventID = &dbModel.ID
		bundle.NotificationEvent = dbModel
	} else if bundle.UpdateNotificationEvent() {
		dbModel, err := q.UpdateNotificationEvent(ctx, nightscout.NotificationEventUpdate(*bundle.NotificationEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to update note in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.NotificationEventID = &dbModel.ID
		bundle.NotificationEvent = dbModel
	}

	return nil
}

func hardDeleteEventGroup(ctx context.Context, q *site.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.HardDeleteEventGroup(ctx, site.HardDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to hard delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to hard delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.HardDeleteGlucoseReading(ctx, site.HardDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return nil, err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.HardDeleteInsulinDose(ctx, site.HardDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return nil, err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.HardDeleteFoodIntake(ctx, site.HardDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.HardDeleteTempBasal(ctx, site.HardDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.HardDeleteTempTarget(ctx, site.HardDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.HardDeleteProfileSwitch(ctx, site.HardDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ComponentEventID != nil {
		if res, err := q.HardDeleteComponentEvent(ctx, site.HardDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to delete component event by ID", "error", err, "componentEventID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.HardDeleteActivity(ctx, site.HardDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NotificationEventID != nil {
		if res, err := q.HardDeleteNotificationEvent(ctx, site.HardDeleteNotificationEventParams{ID: *eventGroup.NotificationEventID}); err != nil {
			slog.Error("failed to delete notification event by ID", "error", err, "notificationEventID", *eventGroup.NotificationEventID, "result", res)
			return nil, err
		}
	}

	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

func softDeleteEventGroup(ctx context.Context, q *site.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.SoftDeleteEventGroup(ctx, site.SoftDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to soft delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to soft delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.SoftDeleteGlucoseReading(ctx, site.SoftDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to soft delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return nil, err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.SoftDeleteInsulinDose(ctx, site.SoftDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to soft delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return nil, err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.SoftDeleteFoodIntake(ctx, site.SoftDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to soft delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.SoftDeleteTempBasal(ctx, site.SoftDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to soft delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.SoftDeleteTempTarget(ctx, site.SoftDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to soft delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.SoftDeleteProfileSwitch(ctx, site.SoftDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to soft delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ComponentEventID != nil {
		if res, err := q.SoftDeleteComponentEvent(ctx, site.SoftDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to soft delete component event by ID", "error", err, "componentEventID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.SoftDeleteActivity(ctx, site.SoftDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to soft delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NotificationEventID != nil {
		if res, err := q.SoftDeleteNotificationEvent(ctx, site.SoftDeleteNotificationEventParams{ID: *eventGroup.NotificationEventID}); err != nil {
			slog.Error("failed to soft delete notification event by ID", "error", err, "notificationEventID", *eventGroup.NotificationEventID, "result", res)
			return nil, err
		}
	}

	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

// Converts an incomming treatment to EventGroups
func TreatmentEventGroupBundle(ctx context.Context, treatment nightscout.Treatment) site.EventGroupsBundle {
	eventGroup := nightscout.TreatmentToEventGroup(treatment)

	glucoseReading := TreatmentToGlucoseReading(treatment)
	insulinDose := TreatmentToInsulinDose(treatment)
	foodIntake := TreatmentToFoodIntake(treatment)
	tempBasal := TreatmentToTempBasal(treatment)
	tempTarget := TreatmentToTempTarget(treatment)
	profileSwitch := TreatmentToProfileSwitch(treatment)
	componentEvent := TreatmentToComponentEvent(treatment)
	activity := TreatmentToActivity(treatment)
	notificationEvent := TreatmentToNotificationEvent(treatment, *treatment.EventType)

	//////
	// Limit the values to only the ones that are relevant to the event type
	//////

	switch *treatment.EventType {
	case nightscout.Announcement:
		return site.EventGroupsBundle{
			EventGroup:        eventGroup,
			NotificationEvent: notificationEvent,
		}
	case nightscout.BGCheck:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
		}
	case nightscout.BolusWizard:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.CarbCorrection:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			FoodIntake: foodIntake,
		}
	case nightscout.ComboBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.CorrectionBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			InsulinDose:    insulinDose,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscout.DADAlert:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
		}
	case nightscout.Exercise:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Activity:   activity,
		}
	case nightscout.InsulinChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.MealBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.Note:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ProfileSwitch:  profileSwitch,
			Activity:       activity,
			ComponentEvent: componentEvent,
		}
	case nightscout.OpenAPSOffline:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.ProfileSwitch:
		return site.EventGroupsBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: profileSwitch,
		}
	case nightscout.PumpBatteryChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Question:
		return site.EventGroupsBundle{
			EventGroup:        eventGroup,
			NotificationEvent: notificationEvent,
		}
	case nightscout.SensorChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SensorStart:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SiteChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SnackBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.TempBasal:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			TempBasal:      tempBasal,
			ComponentEvent: componentEvent,
		}
	case nightscout.TemporaryTarget:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			TempTarget: tempTarget,
		}
	case nightscout.CannulaFill:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Prime:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.ReservoirChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Calibration:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscout.Bolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			InsulinDose:    insulinDose,
			ComponentEvent: componentEvent,
		}
	}

	return site.EventGroupsBundle{EventGroup: eventGroup}
}

// Merge a DB Event Group with one created from the
func MergeUpdateBundle(dbModel *site.EventGroupsBundle, updateBundle site.EventGroupsBundle) {
	nightscout.UpdateToEventGroup(updateBundle.EventGroup, &dbModel.EventGroup)

	if updateBundle.InsulinDose != nil {
		if dbModel.InsulinDose == nil {
			dbModel.InsulinDose = updateBundle.InsulinDose
		} else {
			nightscout.UpdateToInsulinDose(updateBundle.InsulinDose, dbModel.InsulinDose)
		}
	}

	if updateBundle.GlucoseReading != nil {
		if dbModel.GlucoseReading == nil {
			dbModel.GlucoseReading = updateBundle.GlucoseReading
		} else {
			nightscout.UpdateToGlucoseReading(updateBundle.GlucoseReading, dbModel.GlucoseReading)
		}
	}

	if updateBundle.FoodIntake != nil {
		if dbModel.FoodIntake == nil {
			dbModel.FoodIntake = updateBundle.FoodIntake
		} else {
			nightscout.UpdateToFoodIntake(updateBundle.FoodIntake, dbModel.FoodIntake)
		}
	}
	if updateBundle.TempBasal != nil {
		if dbModel.TempBasal == nil {
			dbModel.TempBasal = updateBundle.TempBasal
		} else {
			nightscout.UpdateToTempBasal(updateBundle.TempBasal, dbModel.TempBasal)
		}
	}
	if updateBundle.TempTarget != nil {
		if dbModel.TempTarget == nil {
			dbModel.TempTarget = updateBundle.TempTarget
		} else {
			nightscout.UpdateToTempTarget(updateBundle.TempTarget, dbModel.TempTarget)
		}
	}
	if updateBundle.ProfileSwitch != nil {
		if dbModel.ProfileSwitch == nil {
			dbModel.ProfileSwitch = updateBundle.ProfileSwitch
		} else {
			nightscout.UpdateToProfileSwitch(updateBundle.ProfileSwitch, dbModel.ProfileSwitch)
		}
	}
	if updateBundle.ComponentEvent != nil {
		if dbModel.ComponentEvent == nil {
			dbModel.ComponentEvent = updateBundle.ComponentEvent
		} else {
			nightscout.UpdateToComponentEvent(updateBundle.ComponentEvent, dbModel.ComponentEvent)
		}
	}
	if updateBundle.Activity != nil {
		if dbModel.Activity == nil {
			dbModel.Activity = updateBundle.Activity
		} else {
			nightscout.UpdateToActivity(updateBundle.Activity, dbModel.Activity)
		}
	}
	if updateBundle.NotificationEvent != nil {
		if dbModel.NotificationEvent == nil {
			dbModel.NotificationEvent = updateBundle.NotificationEvent
		} else {
			nightscout.UpdateToNotificationEvent(updateBundle.NotificationEvent, dbModel.NotificationEvent)
		}
	}
}

////////////////////////////////////////
// TREATMENT TO CREATE/UPDATE
////////////////////////////////////////

///////
// Food Intake
///////

func TreatmentToFoodIntake(source nightscout.Treatment) *site.FoodIntake {
	// Check some needed fields are assigned
	if isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions) {
		return ptr(nightscout.PartialTreatmentToFoodIntake(source))
	}
	return nil
}

///////
// Glucose Reading
///////

func TreatmentToGlucoseReading(source nightscout.Treatment) *site.GlucoseReading {
	target := nightscout.PartialTreatmentToGlucoseReading(source)
	target.ReadingType = "sgv"

	if source.GlucoseType != nil {
		switch *source.GlucoseType {
		case nightscout.Sensor:
			target.ReadingType = "sgv"
		case nightscout.Finger:
			target.ReadingType = "mbg"
		case nightscout.Meter:
			target.ReadingType = "mbg"
		}
	}
	if isAssigned(source.Glucose) {
		return &target
	}
	return nil
}

///////
// Insulin Dose
///////

func TreatmentToInsulinDose(source nightscout.Treatment) *site.InsulinDose {
	if isAssigned(source.Insulin) {
		return ptr(nightscout.PartialTreatmentToInsulinDose(source))
	}
	return nil
}

///////
// NotificationEvents
///////

func TreatmentToNotificationEvent(source nightscout.Treatment, eventType nightscout.TreatmentEventType) *site.NotificationEvent {
	target := nightscout.PartialTreatmentToNotificationEvent(source)

	if eventType == nightscout.Announcement {
		target.Severity = 0
		target.NotificationType = "announcement"
		target.Title = "Announcement"
	} else if eventType == nightscout.Question {
		target.Severity = 1
		target.NotificationType = "question"
		target.Title = "Question"
	} else {
		target.Severity = 0
		target.NotificationType = "announcement"
		target.Title = "Treatment Event"
	}

	// Always is assigned
	if source.Name != nil {
		target.Title = source.Name.String()
	}

	return &target
}

///////
// Profile Switch
///////

func TreatmentToProfileSwitch(source nightscout.Treatment) *site.ProfileSwitch {
	target := nightscout.PartialTreatmentToProfileSwitch(source)

	// TODO: This is the bit that will need more work, with the json of profile
	if source.Profile != nil {
		target.Name = source.Profile.String()
	}

	if source.ProfileJSON != nil {
		var profileJSON nightscout.Profile
		err := json.Unmarshal([]byte(*source.ProfileJSON), &profileJSON)
		if err == nil {

			target.Dia = profileJSON.Dia.Float64p()
			target.Timezone = profileJSON.Timezone.Stringp()
			target.Units = profileJSON.Units.Stringp()

			basalBytes, _ := json.Marshal(profileJSON.Basal)
			carbRatioBytes, _ := json.Marshal(profileJSON.CarbRatio)
			sensBytes, _ := json.Marshal(profileJSON.Sens)
			targetLowBytes, _ := json.Marshal(profileJSON.TargetLow)
			targetHighBytes, _ := json.Marshal(profileJSON.TargetHigh)

			if basalBytes != nil {
				target.BasalRatesJson = ptr(string(basalBytes))
			} else {
				target.BasalRatesJson = ptr("[]")
			}

			if carbRatioBytes != nil {
				target.CarbRatiosJson = ptr(string(carbRatioBytes))
			} else {
				target.CarbRatiosJson = ptr("[]")
			}

			if sensBytes != nil {
				target.InsulinSensitivitiesJson = ptr(string(sensBytes))
			} else {
				target.InsulinSensitivitiesJson = ptr("[]")
			}

			if targetLowBytes != nil {
				target.TargetsLowJson = ptr(string(targetLowBytes))
			} else {
				target.TargetsLowJson = ptr("[]")
			}

			if targetHighBytes != nil {
				target.TargetsHighJson = ptr(string(targetHighBytes))
			} else {
				target.TargetsHighJson = ptr("[]")
			}
		}

	}
	if isAssigned(source.Name, source.Profile, source.ProfileJSON) {
		return &target
	}
	return nil
}

///////
// Temp Basal
///////

func TreatmentToTempBasal(source nightscout.Treatment) *site.TempBasal {
	if isAssigned(source.Absolute, source.Rate, source.Percent, source.DurationMinutes) {
		return ptr(nightscout.PartialTreatmentToTempBasal(source))
	}
	return nil
}

///////
// Temp Targets
///////

func TreatmentToTempTarget(source nightscout.Treatment) *site.TempTarget {

	if len(source.CorrectionRange) == 2 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[1]
	} else if len(source.CorrectionRange) == 1 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[0]
	}

	target := nightscout.PartialTreatmentToTempTarget(source)

	if source.HighTarget != nil {
		target.HighTarget = source.HighTarget.Float64()
	} else if source.LowTarget != nil {
		// If Target Top is missing then we assume both targets are the same
		target.HighTarget = source.LowTarget.Float64()
	}

	if source.LowTarget != nil {
		target.LowTarget = source.LowTarget.Float64()
	} else if source.HighTarget != nil {
		// If Target Bottom is missing then we assume both targets are the same
		target.LowTarget = source.HighTarget.Float64()
	}

	if isAssigned(source.HighTarget, source.LowTarget, source.DurationMinutes) {
		return &target
	}
	return nil
}

func TreatmentToComponentEvent(source nightscout.Treatment) *site.ComponentEvent {
	if isAssigned(source.PumpID, source.PumpSerial, source.PumpType) {
		return ptr(nightscout.PartialTreatmentToComponentEvent(source))
	}
	return nil
}

func TreatmentToActivity(source nightscout.Treatment) *site.Activity {
	if isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate) {
		return ptr(nightscout.PartialTreatmentToActivity(source))
	}
	return nil
}

func isAssigned(values ...any) bool {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func:
			if !v.IsNil() {
				return true
			}
		default:
			return true
		}
	}

	return false
}

func DBModelToTreatment(source *site.EventGroupsBundle) nightscout.Treatment {
	treatment := nightscout.EventGroupToTreatment(source.EventGroup)

	nightscout.UpdateGlucoseReadingToTreatment(source.GlucoseReading, treatment)
	nightscout.UpdateInsulinDoseToTreatment(source.InsulinDose, treatment)
	nightscout.UpdateFoodIntakeToTreatment(source.FoodIntake, treatment)
	nightscout.UpdateTempBasalToTreatment(source.TempBasal, treatment)
	nightscout.UpdateTempTargetToTreatment(source.TempTarget, treatment)
	nightscout.UpdateProfileSwitchToTreatment(source.ProfileSwitch, treatment)
	nightscout.UpdateComponentEventToTreatment(source.ComponentEvent, treatment)
	nightscout.UpdateActivityToTreatment(source.Activity, treatment)
	nightscout.UpdateNotificationEventToTreatment(source.NotificationEvent, treatment)

	return *treatment
}
