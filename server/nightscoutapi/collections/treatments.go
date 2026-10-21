package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
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

func (c TreatmentCollection) Validate(nsModel nightscoutmodels.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	// Validate the input
	if nsModel.EventType == nil || *nsModel.EventType == nightscoutmodels.Unknown {
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

func (c TreatmentCollection) ValidatePatch(nsModel nightscoutmodels.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}
	return nil
}

func (c TreatmentCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.EventGroupsBundle, error) {

	tebg := TreatmentEventGroupBundle(ctx, nsModel)
	err := createUpdateEventGroup(ctx, q, &tebg, cc)
	if err != nil {
		return nil, err
	}

	createEventGroupParams := nightscoutmodels.EventGroupCreate(tebg.EventGroup, cc)

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

func (c TreatmentCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment) (*sitedb.EventGroupsBundle, error) {
	// Without a date every filter below is empty and would match the newest row
	if nsModel.Date == nil {
		return nil, nil
	}
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

func (c TreatmentCollection) PatchDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, identifier string, dbModel *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) (*sitedb.EventGroupsBundle, error) {
	// This will be a partial update

	nsModel.EventType = ptr(nightscoutmodels.TreatmentEventType(dbModel.EventGroup.GroupType)) // Ensure the event type is set to the existing value

	updateBundle := TreatmentEventGroupBundle(ctx, nsModel)

	MergeUpdateBundle(dbModel, updateBundle)

	// Do something with the update bundle to merge it into the dbModel

	err := createUpdateEventGroup(ctx, q, dbModel, cc)
	if err != nil {
		return nil, err
	}

	updateEventGroupParams := nightscoutmodels.EventGroupUpdate(dbModel.EventGroup, cc)
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
func createUpdateEventGroup(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {

	// GlucoseReading
	if bundle.DeleteGlucoseReading() {
		_, err := q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: *bundle.EventGroup.GlucoseReadingID})
		if err != nil {
			slog.Error("failed to delete glucose reading from database", "error", err)
			return err
		}
		bundle.EventGroup.GlucoseReadingID = nil
	} else if bundle.CreateGlucoseReading() {
		dbModel, err := q.CreateGlucoseReading(ctx, nightscoutmodels.GlucoseReadingCreate(*bundle.GlucoseReading, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create glucose reading in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.GlucoseReadingID = &dbModel.ID
		bundle.GlucoseReading = dbModel
	} else if bundle.UpdateGlucoseReading() {
		dbModel, err := q.UpdateGlucoseReading(ctx, nightscoutmodels.GlucoseReadingUpdate(*bundle.GlucoseReading, cc))
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
		dbModel, err := q.CreateInsulinDose(ctx, nightscoutmodels.InsulinDoseCreate(*bundle.InsulinDose, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create insulin dose in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.InsulinDoseID = &dbModel.ID
		bundle.InsulinDose = dbModel
	} else if bundle.UpdateInsulinDose() {
		dbModel, err := q.UpdateInsulinDose(ctx, nightscoutmodels.InsulinDoseUpdate(*bundle.InsulinDose, cc))
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
		dbModel, err := q.CreateFoodIntake(ctx, nightscoutmodels.FoodIntakeCreate(*bundle.FoodIntake, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create food intake in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.FoodIntakeID = &dbModel.ID
		bundle.FoodIntake = dbModel
	} else if bundle.UpdateFoodIntake() {
		dbModel, err := q.UpdateFoodIntake(ctx, nightscoutmodels.FoodIntakeUpdate(*bundle.FoodIntake, cc))
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
		dbModel, err := q.CreateTempBasal(ctx, nightscoutmodels.TempBasalCreate(*bundle.TempBasal, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp basal in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempBasalID = &dbModel.ID
		bundle.TempBasal = dbModel
	} else if bundle.UpdateTempBasal() {
		dbModel, err := q.UpdateTempBasal(ctx, nightscoutmodels.TempBasalUpdate(*bundle.TempBasal, cc))
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
		dbModel, err := q.CreateTempTarget(ctx, nightscoutmodels.TempTargetCreate(*bundle.TempTarget, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp target in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.TempTargetID = &dbModel.ID
		bundle.TempTarget = dbModel
	} else if bundle.UpdateTempTarget() {
		dbModel, err := q.UpdateTempTarget(ctx, nightscoutmodels.TempTargetUpdate(*bundle.TempTarget, cc))
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
		dbModel, err := q.CreateProfileSwitch(ctx, nightscoutmodels.ProfileSwitchCreate(*bundle.ProfileSwitch, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create profile switch in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ProfileSwitchID = &dbModel.ID
		bundle.ProfileSwitch = dbModel
	} else if bundle.UpdateProfileSwitch() {
		dbModel, err := q.UpdateProfileSwitch(ctx, nightscoutmodels.ProfileSwitchUpdate(*bundle.ProfileSwitch, cc))
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
		dbModel, err := q.CreateComponentEvent(ctx, nightscoutmodels.ComponentEventCreate(*bundle.ComponentEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create component event in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ComponentEventID = &dbModel.ID
		bundle.ComponentEvent = dbModel
	} else if bundle.UpdateComponentEvent() {
		dbModel, err := q.UpdateComponentEvent(ctx, nightscoutmodels.ComponentEventUpdate(*bundle.ComponentEvent, cc))
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
		dbModel, err := q.CreateActivity(ctx, nightscoutmodels.ActivityCreate(*bundle.Activity, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create activity in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.ActivityID = &dbModel.ID
		bundle.Activity = dbModel
	} else if bundle.UpdateActivity() {
		dbModel, err := q.UpdateActivity(ctx, nightscoutmodels.ActivityUpdate(*bundle.Activity, cc))
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
		dbModel, err := q.CreateNotificationEvent(ctx, nightscoutmodels.NotificationEventCreate(*bundle.NotificationEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.NotificationEventID = &dbModel.ID
		bundle.NotificationEvent = dbModel
	} else if bundle.UpdateNotificationEvent() {
		dbModel, err := q.UpdateNotificationEvent(ctx, nightscoutmodels.NotificationEventUpdate(*bundle.NotificationEvent, cc))
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
	if err := sitedb.SoftDeleteEventGroupAndEvents(ctx, q, identifier, nil); err != nil {
		slog.Error("failed to soft delete event group by ID", "error", err, "identifier", identifier)
		return nil, err
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

// Converts an incomming treatment to EventGroups
func TreatmentEventGroupBundle(ctx context.Context, treatment nightscoutmodels.Treatment) sitedb.EventGroupsBundle {
	eventGroup := nightscoutmodels.TreatmentToEventGroup(treatment)

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
	case nightscoutmodels.Announcement:
		return sitedb.EventGroupsBundle{
			EventGroup:        eventGroup,
			GlucoseReading:    glucoseReading,
			NotificationEvent: notificationEvent,
		}
	case nightscoutmodels.BGCheck:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
		}
	case nightscoutmodels.BolusWizard:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.CarbCorrection:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			FoodIntake: foodIntake,
		}
	case nightscoutmodels.ComboBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.CorrectionBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			InsulinDose:    insulinDose,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.DADAlert:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
		}
	case nightscoutmodels.Exercise:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			Activity:   activity,
		}
	case nightscoutmodels.InsulinChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.MealBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.Note:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ProfileSwitch:  profileSwitch,
			Activity:       activity,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.OpenAPSOffline:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.ProfileSwitch:
		return sitedb.EventGroupsBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: profileSwitch,
		}
	case nightscoutmodels.PumpBatteryChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.Question:
		return sitedb.EventGroupsBundle{
			EventGroup:        eventGroup,
			NotificationEvent: notificationEvent,
		}
	case nightscoutmodels.SensorChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.SensorStart:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.SiteChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.SnackBolus:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.TempBasal:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			TempBasal:      tempBasal,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.TemporaryTarget:
		return sitedb.EventGroupsBundle{
			EventGroup: eventGroup,
			TempTarget: tempTarget,
		}
	case nightscoutmodels.CannulaFill:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.Prime:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.ReservoirChange:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.Calibration:
		return sitedb.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			ComponentEvent: componentEvent,
		}
	case nightscoutmodels.Bolus:
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
	nightscoutmodels.UpdateToEventGroup(updateBundle.EventGroup, &dbModel.EventGroup)

	if updateBundle.InsulinDose != nil {
		if dbModel.InsulinDose == nil {
			dbModel.InsulinDose = updateBundle.InsulinDose
		} else {
			nightscoutmodels.UpdateToInsulinDose(updateBundle.InsulinDose, dbModel.InsulinDose)
		}
	}

	if updateBundle.GlucoseReading != nil {
		if dbModel.GlucoseReading == nil {
			dbModel.GlucoseReading = updateBundle.GlucoseReading
		} else {
			nightscoutmodels.UpdateToGlucoseReading(updateBundle.GlucoseReading, dbModel.GlucoseReading)
		}
	}

	if updateBundle.FoodIntake != nil {
		if dbModel.FoodIntake == nil {
			dbModel.FoodIntake = updateBundle.FoodIntake
		} else {
			nightscoutmodels.UpdateToFoodIntake(updateBundle.FoodIntake, dbModel.FoodIntake)
		}
	}
	if updateBundle.TempBasal != nil {
		if dbModel.TempBasal == nil {
			dbModel.TempBasal = updateBundle.TempBasal
		} else {
			nightscoutmodels.UpdateToTempBasal(updateBundle.TempBasal, dbModel.TempBasal)
		}
	}
	if updateBundle.TempTarget != nil {
		if dbModel.TempTarget == nil {
			dbModel.TempTarget = updateBundle.TempTarget
		} else {
			nightscoutmodels.UpdateToTempTarget(updateBundle.TempTarget, dbModel.TempTarget)
		}
	}
	if updateBundle.ProfileSwitch != nil {
		if dbModel.ProfileSwitch == nil {
			dbModel.ProfileSwitch = updateBundle.ProfileSwitch
		} else {
			nightscoutmodels.UpdateToProfileSwitch(updateBundle.ProfileSwitch, dbModel.ProfileSwitch)
		}
	}
	if updateBundle.ComponentEvent != nil {
		if dbModel.ComponentEvent == nil {
			dbModel.ComponentEvent = updateBundle.ComponentEvent
		} else {
			nightscoutmodels.UpdateToComponentEvent(updateBundle.ComponentEvent, dbModel.ComponentEvent)
		}
	}
	if updateBundle.Activity != nil {
		if dbModel.Activity == nil {
			dbModel.Activity = updateBundle.Activity
		} else {
			nightscoutmodels.UpdateToActivity(updateBundle.Activity, dbModel.Activity)
		}
	}
	if updateBundle.NotificationEvent != nil {
		if dbModel.NotificationEvent == nil {
			dbModel.NotificationEvent = updateBundle.NotificationEvent
		} else {
			nightscoutmodels.UpdateToNotificationEvent(updateBundle.NotificationEvent, dbModel.NotificationEvent)
		}
	}
}

////////////////////////////////////////
// TREATMENT TO CREATE/UPDATE
////////////////////////////////////////

///////
// Food Intake
///////

func TreatmentToFoodIntake(source nightscoutmodels.Treatment) *sitedb.FoodIntake {
	// Check some needed fields are assigned
	if isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions) {
		return ptr(nightscoutmodels.PartialTreatmentToFoodIntake(source))
	}
	return nil
}

///////
// Glucose Reading
///////

func TreatmentToGlucoseReading(source nightscoutmodels.Treatment) *sitedb.GlucoseReading {
	target := nightscoutmodels.PartialTreatmentToGlucoseReading(source)
	target.ReadingType = "sgv"
	target.Units = source.Units

	if source.Glucose != nil {
		target.Mgdl = target.Units.ToMGDL(float64(*source.Glucose))
	}

	if source.GlucoseType != nil {
		switch *source.GlucoseType {
		case nightscoutmodels.Sensor:
			target.ReadingType = "sgv"
		case nightscoutmodels.Finger:
			target.ReadingType = "mbg"
		case nightscoutmodels.Meter:
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

func TreatmentToInsulinDose(source nightscoutmodels.Treatment) *sitedb.InsulinDose {
	if isAssigned(source.Insulin) {
		return ptr(nightscoutmodels.PartialTreatmentToInsulinDose(source))
	}
	return nil
}

///////
// NotificationEvents
///////

func TreatmentToNotificationEvent(source nightscoutmodels.Treatment, eventType nightscoutmodels.TreatmentEventType) *sitedb.NotificationEvent {
	target := nightscoutmodels.PartialTreatmentToNotificationEvent(source)

	if eventType == nightscoutmodels.Announcement {
		target.Severity = 0
		target.NotificationType = "announcement"
		target.Title = "Announcement"
	} else if eventType == nightscoutmodels.Question {
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

func TreatmentToProfileSwitch(source nightscoutmodels.Treatment) *sitedb.ProfileSwitch {
	target := nightscoutmodels.PartialTreatmentToProfileSwitch(source)

	// TODO: This is the bit that will need more work, with the json of profile
	if source.Profile != nil {
		target.Name = source.Profile.String()
	}

	if source.ProfileJSON != nil {
		var profile nightscoutmodels.Profile
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

func TreatmentToTempBasal(source nightscoutmodels.Treatment) *sitedb.TempBasal {
	if isAssigned(source.Absolute, source.Rate, source.Percent, source.DurationMinutes) {
		return ptr(nightscoutmodels.PartialTreatmentToTempBasal(source))
	}
	return nil
}

///////
// Temp Targets
///////

func TreatmentToTempTarget(source nightscoutmodels.Treatment) *sitedb.TempTarget {

	if len(source.CorrectionRange) == 2 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[1]
	} else if len(source.CorrectionRange) == 1 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[0]
	}

	target := nightscoutmodels.PartialTreatmentToTempTarget(source)

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

func TreatmentToComponentEvent(source nightscoutmodels.Treatment) *sitedb.ComponentEvent {
	if isAssigned(source.PumpID, source.PumpSerial, source.PumpType) {
		return ptr(nightscoutmodels.PartialTreatmentToComponentEvent(source))
	}
	return nil
}

func TreatmentToActivity(source nightscoutmodels.Treatment) *sitedb.Activity {
	if isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate) {
		return ptr(nightscoutmodels.PartialTreatmentToActivity(source))
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

func DBModelToTreatment(source *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) nightscoutmodels.Treatment {
	treatment := nightscoutmodels.EventGroupToTreatment(source.EventGroup)

	nightscoutmodels.UpdateGlucoseReadingToTreatment(source.GlucoseReading, siteContext, treatment)
	nightscoutmodels.UpdateInsulinDoseToTreatment(source.InsulinDose, treatment)
	nightscoutmodels.UpdateFoodIntakeToTreatment(source.FoodIntake, treatment)
	nightscoutmodels.UpdateTempBasalToTreatment(source.TempBasal, treatment)
	nightscoutmodels.UpdateTempTargetToTreatment(source.TempTarget, treatment)
	nightscoutmodels.UpdateProfileSwitchToTreatment(source.ProfileSwitch, siteContext, treatment)
	nightscoutmodels.UpdateComponentEventToTreatment(source.ComponentEvent, treatment)
	nightscoutmodels.UpdateActivityToTreatment(source.Activity, treatment)
	nightscoutmodels.UpdateNotificationEventToTreatment(source.NotificationEvent, treatment)

	return *treatment
}
