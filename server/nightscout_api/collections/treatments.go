package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	converter func(c *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record
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

func (c TreatmentCollection) Convert(dr *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c TreatmentCollection) Validate(nsModel nightscout.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	// Validate the input
	if nsModel.EventType == nil || *nsModel.EventType == nightscout.Unknown {
		eventTypeError := "eventType: null"
		if nsModel.EventType != nil {
			eventTypeError = "eventType: " + nsModel.EventType.String()
		}
		slog.Error("treatment EventType is empty or unknown", "error", eventTypeError)
		return nsutils.BadDataBodyErrorWithHint("EventType is nil or unknown: " + eventTypeError)
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

func (c TreatmentCollection) ValidatePatch(nsModel nightscout.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}
	return nil
}

func (c TreatmentCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscout.Treatment, cc nightscout.ConverterContext, overrideIdentifier *string) (*sitedb.EventGroupsBundle, error) {

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

func (c TreatmentCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscout.Treatment) (*sitedb.EventGroupsBundle, error) {
	eventGroups, rerr := q.SearchEventGroups(ctx, sitedb.SearchEventGroupsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		TypeEq:   nsModel.EventType.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(eventGroups) > 0 {
		egb := sitedb.GetEventGroupToEventGroupsBundle(sitedb.SearchEventGroupToGetEventGroup(*eventGroups[0]))
		return &egb, nil
	}
	return nil, nil
}

func (c TreatmentCollection) GetDBFunc(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: identifier, IncludeDeleted: includeDeleted})
	if err != nil {
		return nil, err
	}

	egb := sitedb.GetEventGroupToEventGroupsBundle(*dbModel)

	return &egb, nil
}

func (c TreatmentCollection) PatchDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscout.Treatment, identifier string, dbModel *sitedb.EventGroupsBundle, cc nightscout.ConverterContext) (*sitedb.EventGroupsBundle, error) {
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

func (c TreatmentCollection) HardDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return hardDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) SoftDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return softDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) HardDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {

	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString, IncludeDeleted: true})
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

func (c TreatmentCollection) SoftDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString})
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
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	treatments, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(treatments, c.converter, siteContext, query.Raw), nil
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

func EventGroupSearch(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*sitedb.EventGroupsBundle, *nsutils.HTTPError) {
	eventGroups, rerr := siteContext.SiteDB.Queries.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if rerr != nil {
		slog.Error("failed to query event groups from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	res := []*sitedb.EventGroupsBundle{}
	for _, eg := range eventGroups {
		if eg != nil {
			egb := sitedb.GetEventGroupToEventGroupsBundle(sitedb.SearchEventGroupToGetEventGroup(*eg))
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
func createUpdateEventGroup(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscout.ConverterContext) error {

	// GlucoseReading
	if bundle.DeleteGlucoseReading() {
		_, err := q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: *bundle.EventGroup.GlucoseReadingID})
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
		_, err := q.HardDeleteInsulinDose(ctx, sitedb.HardDeleteInsulinDoseParams{ID: *bundle.EventGroup.InsulinDoseID})
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
		_, err := q.HardDeleteFoodIntake(ctx, sitedb.HardDeleteFoodIntakeParams{ID: *bundle.EventGroup.FoodIntakeID})
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
		_, err := q.HardDeleteTempBasal(ctx, sitedb.HardDeleteTempBasalParams{ID: *bundle.EventGroup.TempBasalID})
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
		_, err := q.HardDeleteTempTarget(ctx, sitedb.HardDeleteTempTargetParams{ID: *bundle.EventGroup.TempTargetID})
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
		_, err := q.HardDeleteProfileSwitch(ctx, sitedb.HardDeleteProfileSwitchParams{ID: *bundle.EventGroup.ProfileSwitchID})
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
		_, err := q.HardDeleteComponentEvent(ctx, sitedb.HardDeleteComponentEventParams{ID: *bundle.EventGroup.ComponentEventID})
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
		_, err := q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: *bundle.EventGroup.ActivityID})
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
		_, err := q.HardDeleteNotificationEvent(ctx, sitedb.HardDeleteNotificationEventParams{ID: *bundle.EventGroup.NotificationEventID})
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

func hardDeleteEventGroup(ctx context.Context, q *sitedb.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.HardDeleteEventGroup(ctx, sitedb.HardDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to hard delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to hard delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return nil, err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.HardDeleteInsulinDose(ctx, sitedb.HardDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return nil, err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.HardDeleteFoodIntake(ctx, sitedb.HardDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.HardDeleteTempBasal(ctx, sitedb.HardDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.HardDeleteTempTarget(ctx, sitedb.HardDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.HardDeleteProfileSwitch(ctx, sitedb.HardDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ComponentEventID != nil {
		if res, err := q.HardDeleteComponentEvent(ctx, sitedb.HardDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to delete component event by ID", "error", err, "componentEventID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NotificationEventID != nil {
		if res, err := q.HardDeleteNotificationEvent(ctx, sitedb.HardDeleteNotificationEventParams{ID: *eventGroup.NotificationEventID}); err != nil {
			slog.Error("failed to delete notification event by ID", "error", err, "notificationEventID", *eventGroup.NotificationEventID, "result", res)
			return nil, err
		}
	}

	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

func softDeleteEventGroup(ctx context.Context, q *sitedb.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.SoftDeleteEventGroup(ctx, sitedb.SoftDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to soft delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to soft delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to soft delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return nil, err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to soft delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return nil, err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.SoftDeleteFoodIntake(ctx, sitedb.SoftDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to soft delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.SoftDeleteTempBasal(ctx, sitedb.SoftDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to soft delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return nil, err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.SoftDeleteTempTarget(ctx, sitedb.SoftDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to soft delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.SoftDeleteProfileSwitch(ctx, sitedb.SoftDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to soft delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ComponentEventID != nil {
		if res, err := q.SoftDeleteComponentEvent(ctx, sitedb.SoftDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to soft delete component event by ID", "error", err, "componentEventID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to soft delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NotificationEventID != nil {
		if res, err := q.SoftDeleteNotificationEvent(ctx, sitedb.SoftDeleteNotificationEventParams{ID: *eventGroup.NotificationEventID}); err != nil {
			slog.Error("failed to soft delete notification event by ID", "error", err, "notificationEventID", *eventGroup.NotificationEventID, "result", res)
			return nil, err
		}
	}

	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

// Converts an incomming treatment to EventGroups
func TreatmentEventGroupBundle(ctx context.Context, treatment nightscout.Treatment) sitedb.EventGroupsBundle {
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
		return sitedb.EventGroupsBundle{
			EventGroup:        eventGroup,
			GlucoseReading:    glucoseReading,
			NotificationEvent: notificationEvent,
		}
	case nightscout.BGCheck:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
		}
	case nightscout.BolusWizard:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.CarbCorrection:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			FoodIntake: foodIntake,
		}
	case nightscout.ComboBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.CorrectionBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			InsulinDose:    insulinDose,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscout.DADAlert:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
		}
	case nightscout.Exercise:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			Activity:   activity,
		}
	case nightscout.InsulinChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.MealBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.Note:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ProfileSwitch:  profileSwitch,
			Activity:       activity,
			ComponentEvent: componentEvent,
		}
	case nightscout.OpenAPSOffline:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.ProfileSwitch:
		return sitedb.EventGroupsBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: profileSwitch,
		}
	case nightscout.PumpBatteryChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Question:
		return sitedb.EventGroupsBundle{
			EventGroup:        eventGroup,
			NotificationEvent: notificationEvent,
		}
	case nightscout.SensorChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SensorStart:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SiteChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.SnackBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscout.TempBasal:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			TempBasal:      tempBasal,
			ComponentEvent: componentEvent,
		}
	case nightscout.TemporaryTarget:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			TempTarget: tempTarget,
		}
	case nightscout.CannulaFill:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Prime:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.ReservoirChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscout.Calibration:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscout.Bolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			InsulinDose:    insulinDose,
			ComponentEvent: componentEvent,
		}
	}

	return sitedb.EventGroupsBundle{EventGroup: eventGroup}
}

// Merge a DB Event Group with one created from the
func MergeUpdateBundle(dbModel *sitedb.EventGroupsBundle, updateBundle sitedb.EventGroupsBundle) {
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

func TreatmentToFoodIntake(source nightscout.Treatment) *sitedb.FoodIntake {
	// Check some needed fields are assigned
	if isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions) {
		return ptr(nightscout.PartialTreatmentToFoodIntake(source))
	}
	return nil
}

///////
// Glucose Reading
///////

func TreatmentToGlucoseReading(source nightscout.Treatment) *sitedb.GlucoseReading {
	target := nightscout.PartialTreatmentToGlucoseReading(source)
	target.ReadingType = "sgv"
	target.Units = source.Units

	if source.Glucose != nil {
		target.Mgdl = target.Units.ToMGDL(float64(*source.Glucose))
	}

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

func TreatmentToInsulinDose(source nightscout.Treatment) *sitedb.InsulinDose {
	if isAssigned(source.Insulin) {
		return ptr(nightscout.PartialTreatmentToInsulinDose(source))
	}
	return nil
}

///////
// NotificationEvents
///////

func TreatmentToNotificationEvent(source nightscout.Treatment, eventType nightscout.TreatmentEventType) *sitedb.NotificationEvent {
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

func TreatmentToProfileSwitch(source nightscout.Treatment) *sitedb.ProfileSwitch {
	target := nightscout.PartialTreatmentToProfileSwitch(source)

	// TODO: This is the bit that will need more work, with the json of profile
	if source.Profile != nil {
		target.Name = source.Profile.String()
	}

	if source.ProfileJSON != nil {
		var profile nightscout.Profile
		err := json.Unmarshal([]byte(*source.ProfileJSON), &profile)
		if err == nil {

			profile.ConvertToMGDL()

			target.Dia = profile.Dia.Float64p()
			target.Timezone = profile.Timezone.Stringp()

			// Set units
			target.Units = profile.Units
			target.CarbsHr = profile.CarbsHr.Float64p()
			target.Delay = profile.Delay.Float64p()

			basalBytes, _ := json.Marshal(profile.Basal)
			carbRatioBytes, _ := json.Marshal(profile.CarbRatio)
			sensBytes, _ := json.Marshal(profile.Sens)
			targetLowBytes, _ := json.Marshal(profile.TargetLow)
			targetHighBytes, _ := json.Marshal(profile.TargetHigh)

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

func TreatmentToTempBasal(source nightscout.Treatment) *sitedb.TempBasal {
	if isAssigned(source.Absolute, source.Rate, source.Percent, source.DurationMinutes) {
		return ptr(nightscout.PartialTreatmentToTempBasal(source))
	}
	return nil
}

///////
// Temp Targets
///////

func TreatmentToTempTarget(source nightscout.Treatment) *sitedb.TempTarget {

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

func TreatmentToComponentEvent(source nightscout.Treatment) *sitedb.ComponentEvent {
	if isAssigned(source.PumpID, source.PumpSerial, source.PumpType) {
		return ptr(nightscout.PartialTreatmentToComponentEvent(source))
	}
	return nil
}

func TreatmentToActivity(source nightscout.Treatment) *sitedb.Activity {
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

func DBModelToTreatment(source *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) nightscout.Treatment {
	treatment := nightscout.EventGroupToTreatment(source.EventGroup)

	nightscout.UpdateGlucoseReadingToTreatment(source.GlucoseReading, siteContext, treatment)
	nightscout.UpdateInsulinDoseToTreatment(source.InsulinDose, treatment)
	nightscout.UpdateFoodIntakeToTreatment(source.FoodIntake, treatment)
	nightscout.UpdateTempBasalToTreatment(source.TempBasal, treatment)
	nightscout.UpdateTempTargetToTreatment(source.TempTarget, treatment)
	nightscout.UpdateProfileSwitchToTreatment(source.ProfileSwitch, siteContext, treatment)
	nightscout.UpdateComponentEventToTreatment(source.ComponentEvent, treatment)
	nightscout.UpdateActivityToTreatment(source.Activity, treatment)
	nightscout.UpdateNotificationEventToTreatment(source.NotificationEvent, treatment)

	return *treatment
}
