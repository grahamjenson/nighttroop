package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type SQLResultCustomResult struct {
	LastInsertIdInt int64
	RowsAffectedInt int64
	Err             error
}

func (s SQLResultCustomResult) LastInsertId() (int64, error) {
	return s.LastInsertIdInt, s.Err
}

func (s SQLResultCustomResult) RowsAffected() (int64, error) {
	return s.RowsAffectedInt, s.Err
}

type TreatmentCollection struct {
	converter func(c *site.EventGroupsBundle) nsutils.Record
}

func NewTreatmentCollection() TreatmentCollection {
	return TreatmentCollection{
		converter: nsutils.DBModelToRecord(nsutils.Treatments, nightscout.DBModelToTreatment),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c TreatmentCollection) Name() string {
	return "treatments"
}

func (c TreatmentCollection) Convert(dr *site.EventGroupsBundle) nsutils.Record {
	return c.converter(dr)
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
	return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
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
	return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c TreatmentCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c TreatmentCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	treatments, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(treatments, c.converter), nil
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

// Replace will replace (delete, then insert) if it exists, or create it if it doesnt
func (c TreatmentCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate)
}

func (c TreatmentCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	// We will always repalce because it is way harder otherwise
	return MetaPatch(c, ctx, siteContext, identifier, patch)
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

// createUpdateEventGroup goes through an EventGroupsBundle if the EventGroup has an ID it will update, if not it will attempt to create
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
		// Id exists no model
		_, err := q.HardDeleteNotificationEvent(ctx, site.HardDeleteNotificationEventParams{ID: *bundle.EventGroup.NotificationEventID})
		if err != nil {
			slog.Error("failed to delete note from database", "error", err)
			return err
		}
		bundle.EventGroup.NotificationEventID = nil
	} else if bundle.CreateNotificationEvent() {
		// model exists not id
		// If the note is not deleted, but a new note is provided, we need to create it
		dbModel, err := q.CreateNotificationEvent(ctx, nightscout.NotificationEventCreate(*bundle.NotificationEvent, cc))
		if err != nil || dbModel == nil {
			slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
			return err
		}
		bundle.EventGroup.NotificationEventID = &dbModel.ID
		bundle.NotificationEvent = dbModel
	} else if bundle.UpdateNotificationEvent() {
		// model and id exist

		updateParams := nightscout.NotificationEventUpdate(*bundle.NotificationEvent, cc)
		a, _ := json.Marshal(updateParams)
		b, _ := json.Marshal(bundle.NotificationEvent)
		fmt.Println("ASDSAFSAFSAFASF", string(a), "note", string(b))
		dbModel, err := q.UpdateNotificationEvent(ctx, updateParams)
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
		return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
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

	return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

func softDeleteEventGroup(ctx context.Context, q *site.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.SoftDeleteEventGroup(ctx, site.SoftDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to soft delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to soft delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
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

	return SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

func TreatmentEventGroupBundle(ctx context.Context, treatment nightscout.Treatment) site.EventGroupsBundle {
	eventGroup := nightscout.TreatmentToEventGroup(treatment)

	glucoseReading := ptr(nightscout.TreatmentToGlucoseReading(treatment))
	insulinDose := ptr(nightscout.TreatmentToInsulinDose(treatment))
	foodIntake := ptr(nightscout.TreatmentToFoodIntake(treatment))
	tempBasal := ptr(nightscout.TreatmentToTempBasal(treatment))
	tempTarget := ptr(nightscout.TreatmentToTempTarget(treatment))
	profileSwitch := ptr(nightscout.TreatmentToProfileSwitch(treatment))
	componentEvent := ptr(nightscout.TreatmentToComponentEvent(treatment))
	activity := ptr(nightscout.TreatmentToActivity(treatment))
	notificationEvent := ptr(nightscout.TreatmentToNotificationEvent(treatment))

	///////
	// VALIDATE THE CREATED VALUES
	///////

	if glucoseReading.ReadingType == "Sensor" {
		glucoseReading.ReadingType = "sgv"
	}

	insulinDose.DoseType = "bolus"

	/////
	// Filter the values
	/////

	if glucoseReading.Reading == 0 {
		glucoseReading = nil
	}

	if insulinDose.Dose == 0 {
		insulinDose = nil
	}

	if foodIntake.Name == nil && foodIntake.Category == nil && foodIntake.Subcategory == nil && foodIntake.Carbs == nil && foodIntake.Fat == nil && foodIntake.Protein == nil && foodIntake.Fibre == nil && foodIntake.Energy == nil && foodIntake.Portions == nil {
		foodIntake = nil
	}

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
		}
	case nightscout.CorrectionBolus:
		return site.EventGroupsBundle{
			EventGroup:  eventGroup,
			InsulinDose: insulinDose,
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
		}
	case nightscout.Note:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
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
			EventGroup: eventGroup,
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
		}
	case nightscout.TempBasal:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			TempBasal:  tempBasal,
		}
	case nightscout.TemporaryTarget:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			TempTarget: tempTarget,
		}
	}

	return site.EventGroupsBundle{EventGroup: eventGroup}
}

func MergeUpdateBundle(dbModel *site.EventGroupsBundle, updateBundle site.EventGroupsBundle) {

	nightscout.UpdateToEventGroup(updateBundle.EventGroup, &dbModel.EventGroup)

	if updateBundle.GlucoseReading != nil {
		if dbModel.GlucoseReading == nil {
			dbModel.GlucoseReading = updateBundle.GlucoseReading
		} else {
			nightscout.UpdateToGlucoseReading(*updateBundle.GlucoseReading, dbModel.GlucoseReading)
		}
	}
	if updateBundle.InsulinDose != nil {
		if dbModel.InsulinDose == nil {
			dbModel.InsulinDose = updateBundle.InsulinDose
		} else {
			nightscout.UpdateToInsulinDose(*updateBundle.InsulinDose, dbModel.InsulinDose)
		}
	}
	if updateBundle.FoodIntake != nil {
		if dbModel.FoodIntake == nil {
			dbModel.FoodIntake = updateBundle.FoodIntake
		} else {
			nightscout.UpdateToFoodIntake(*updateBundle.FoodIntake, dbModel.FoodIntake)
		}
	}
	if updateBundle.TempBasal != nil {
		if dbModel.TempBasal == nil {
			dbModel.TempBasal = updateBundle.TempBasal
		} else {
			nightscout.UpdateToTempBasal(*updateBundle.TempBasal, dbModel.TempBasal)
		}
	}
	if updateBundle.TempTarget != nil {
		if dbModel.TempTarget == nil {
			dbModel.TempTarget = updateBundle.TempTarget
		} else {
			nightscout.UpdateToTempTarget(*updateBundle.TempTarget, dbModel.TempTarget)
		}
	}
	if updateBundle.ProfileSwitch != nil {
		if dbModel.ProfileSwitch == nil {
			dbModel.ProfileSwitch = updateBundle.ProfileSwitch
		} else {
			nightscout.UpdateToProfileSwitch(*updateBundle.ProfileSwitch, dbModel.ProfileSwitch)
		}
	}
	if updateBundle.ComponentEvent != nil {
		if dbModel.ComponentEvent == nil {
			dbModel.ComponentEvent = updateBundle.ComponentEvent
		} else {
			nightscout.UpdateToComponentEvent(*updateBundle.ComponentEvent, dbModel.ComponentEvent)
		}
	}
	if updateBundle.Activity != nil {
		if dbModel.Activity == nil {
			dbModel.Activity = updateBundle.Activity
		} else {
			nightscout.UpdateToActivity(*updateBundle.Activity, dbModel.Activity)
		}
	}
	if updateBundle.NotificationEvent != nil {
		if dbModel.NotificationEvent == nil {
			dbModel.NotificationEvent = updateBundle.NotificationEvent
		} else {
			nightscout.UpdateToNotificationEvent(*updateBundle.NotificationEvent, dbModel.NotificationEvent)
		}
	}
}
