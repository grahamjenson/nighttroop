package collections

import (
	"context"
	"database/sql"
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

func (c TreatmentCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.EventGroupsBundle, error) {

	createParams := TreatmentEventGroupBundle(ctx, nsModel)

	if overrideIdentifier != nil {
		createParams.EventGroup.ID = *overrideIdentifier
	}

	dbModel, err := createUpdateEventGroup(ctx, q, createParams, cc)
	if err != nil {
		return nil, err
	}

	newEventGroup, err := q.CreateEventGroup(ctx, nightscout.EventGroupCreate(createParams.EventGroup, cc))
	if err != nil {
		slog.Error("failed to create event group in database", "error", err, "createParams", createParams)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", createParams)
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	dbModel.EventGroup = *newEventGroup

	return dbModel, nil
}

func (c TreatmentCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, site.GetEventGroupParams{ID: identifier})
	if err != nil {
		return nil, err
	}
	egb := site.GetEventGroupToEventGroupsBundle(*dbModel)
	return &egb, nil
}

func (c TreatmentCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, identifier string, dbModel *site.EventGroupsBundle, uc nightscout.ConverterContext) (*site.EventGroupsBundle, error) {
	/////
	// Delete all related models
	/////

	/////
	// Create the new models using the patches merged with existing values
	//////

	/////
	// Update Event Group Referencing the new models
	/////

	// createSubsParams := TreatmentEventGroupCreate(ctx, uc, nsModel)

	// updateParams.EventGroup.ID = identifier
	// eventGroup := nightscout.TreatmentToUpdateEventGroup(treatment, cc)

	// // Finaly Update the EventGroup
	// dbModel, err := q.UpdateEventGroup(ctx, bundle.EventGroup)
	// if err != nil {
	// 	slog.Error("failed in updateEventGroup to update event group in database", "error", err, "updateParams", bundle.EventGroup)
	// 	return nil, err
	// }
	// if dbModel == nil {
	// 	slog.Error("failed in updateEventGroup to update event group in database, dbModel is nil", "updateParams", bundle.EventGroup)
	// 	return nil, err
	// }

	// if err != nil {
	// 	return nil, err
	// }

	// if dbModel == nil || dbModel.EventGroup.ID == "" {
	// 	slog.Error("failed to update event group in database, EventGroup is nil")
	// 	return nil, err
	// }

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
func createUpdateEventGroup(ctx context.Context, q *site.Queries, bundle site.EventGroupsBundle, cc nightscout.ConverterContext) (*site.EventGroupsBundle, error) {
	ret := site.EventGroupsBundle{}

	if bundle.GlucoseReading != nil {
		if bundle.EventGroup.GlucoseReadingID == nil {
			bundle.EventGroup.GlucoseReadingID = &bundle.GlucoseReading.ID
			dbModel, err := q.CreateGlucoseReading(ctx, nightscout.GlucoseReadingCreate(*bundle.GlucoseReading, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create glucose reading in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.GlucoseReading = dbModel
		} else {
			bundle.EventGroup.GlucoseReadingID = &bundle.GlucoseReading.ID
			dbModel, err := q.UpdateGlucoseReading(ctx, nightscout.GlucoseReadingUpdate(*bundle.GlucoseReading, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update glucose reading in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.GlucoseReading = dbModel
		}
	}

	if bundle.InsulinDose != nil {
		if bundle.EventGroup.InsulinDoseID == nil {
			bundle.EventGroup.InsulinDoseID = &bundle.InsulinDose.ID
			dbModel, err := q.CreateInsulinDose(ctx, nightscout.InsulinDoseCreate(*bundle.InsulinDose, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create insulin dose in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.InsulinDose = dbModel
		} else {
			bundle.EventGroup.InsulinDoseID = &bundle.InsulinDose.ID
			dbModel, err := q.UpdateInsulinDose(ctx, nightscout.InsulinDoseUpdate(*bundle.InsulinDose, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update insulin dose in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.InsulinDose = dbModel
		}

	}

	if bundle.FoodIntake != nil {
		if bundle.EventGroup.FoodIntakeID == nil {
			bundle.EventGroup.FoodIntakeID = &bundle.FoodIntake.ID
			dbModel, err := q.CreateFoodIntake(ctx, nightscout.FoodIntakeCreate(*bundle.FoodIntake, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create food intake in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.FoodIntake = dbModel
		} else {
			bundle.EventGroup.FoodIntakeID = &bundle.FoodIntake.ID
			dbModel, err := q.UpdateFoodIntake(ctx, nightscout.FoodIntakeUpdate(*bundle.FoodIntake, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update food intake in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.FoodIntake = dbModel
		}
	}

	if bundle.TempBasal != nil {
		if bundle.EventGroup.TempBasalID == nil {
			bundle.EventGroup.TempBasalID = &bundle.TempBasal.ID
			dbModel, err := q.CreateTempBasal(ctx, nightscout.TempBasalCreate(*bundle.TempBasal, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create temp basal in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.TempBasal = dbModel
		} else {
			bundle.EventGroup.TempBasalID = &bundle.TempBasal.ID
			dbModel, err := q.UpdateTempBasal(ctx, nightscout.TempBasalUpdate(*bundle.TempBasal, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update temp basal in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.TempBasal = dbModel
		}
	}

	if bundle.TempTarget != nil {
		if bundle.EventGroup.TempTargetID == nil {
			bundle.EventGroup.TempTargetID = &bundle.TempTarget.ID
			dbModel, err := q.CreateTempTarget(ctx, nightscout.TempTargetCreate(*bundle.TempTarget, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create temp target in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.TempTarget = dbModel
		} else {
			bundle.EventGroup.TempTargetID = &bundle.TempTarget.ID
			dbModel, err := q.UpdateTempTarget(ctx, nightscout.TempTargetUpdate(*bundle.TempTarget, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update temp target in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.TempTarget = dbModel
		}
	}

	if bundle.ProfileSwitch != nil {
		if bundle.EventGroup.ProfileSwitchID == nil {
			bundle.EventGroup.ProfileSwitchID = &bundle.ProfileSwitch.ID
			dbModel, err := q.CreateProfileSwitch(ctx, nightscout.ProfileSwitchCreate(*bundle.ProfileSwitch, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create profile switch in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.ProfileSwitch = dbModel
		} else {
			bundle.EventGroup.ProfileSwitchID = &bundle.ProfileSwitch.ID
			dbModel, err := q.UpdateProfileSwitch(ctx, nightscout.ProfileSwitchUpdate(*bundle.ProfileSwitch, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update profile switch in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.ProfileSwitch = dbModel
		}
	}

	if bundle.ComponentEvent != nil {
		if bundle.EventGroup.ComponentEventID == nil {
			bundle.EventGroup.ComponentEventID = &bundle.ComponentEvent.ID
			dbModel, err := q.CreateComponentEvent(ctx, nightscout.ComponentEventCreate(*bundle.ComponentEvent, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create component event in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.ComponentEvent = dbModel
		} else {
			bundle.EventGroup.ComponentEventID = &bundle.ComponentEvent.ID
			dbModel, err := q.UpdateComponentEvent(ctx, nightscout.ComponentEventUpdate(*bundle.ComponentEvent, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update component event in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.ComponentEvent = dbModel
		}
	}

	if bundle.Activity != nil {
		if bundle.EventGroup.ActivityID == nil {
			bundle.EventGroup.ActivityID = &bundle.Activity.ID
			dbModel, err := q.CreateActivity(ctx, nightscout.ActivityCreate(*bundle.Activity, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create activity in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.Activity = dbModel
		} else {
			bundle.EventGroup.ActivityID = &bundle.Activity.ID
			dbModel, err := q.UpdateActivity(ctx, nightscout.ActivityUpdate(*bundle.Activity, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update activity in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.Activity = dbModel
		}
	}

	if bundle.Note != nil {
		if bundle.EventGroup.NoteID == nil {
			bundle.EventGroup.NoteID = &bundle.Note.ID
			dbModel, err := q.CreateNote(ctx, nightscout.NoteCreate(*bundle.Note, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.Note = dbModel
		} else {
			bundle.EventGroup.NoteID = &bundle.Note.ID
			dbModel, err := q.UpdateNote(ctx, nightscout.NoteUpdate(*bundle.Note, cc))
			if err != nil || dbModel == nil {
				slog.Error("failed to update note in database", "error", err, "dbModel", dbModel)
				return nil, err
			}
			ret.Note = dbModel
		}
	}

	return &ret, nil
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
			slog.Error("failed to delete component event by ID", "error", err, "deviceStateID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.HardDeleteActivity(ctx, site.HardDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NoteID != nil {
		if res, err := q.HardDeleteNote(ctx, site.HardDeleteNoteParams{ID: *eventGroup.NoteID}); err != nil {
			slog.Error("failed to delete note by ID", "error", err, "noteID", *eventGroup.NoteID, "result", res)
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
			slog.Error("failed to soft delete component event by ID", "error", err, "deviceStateID", *eventGroup.ComponentEventID, "result", res)
			return nil, err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.SoftDeleteActivity(ctx, site.SoftDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to soft delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return nil, err
		}
	}
	if eventGroup.NoteID != nil {
		if res, err := q.SoftDeleteNote(ctx, site.SoftDeleteNoteParams{ID: *eventGroup.NoteID}); err != nil {
			slog.Error("failed to soft delete note by ID", "error", err, "noteID", *eventGroup.NoteID, "result", res)
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
	note := ptr(nightscout.TreatmentToNote(treatment))

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

	if note.Content == "" || note.Reason == nil {
		note = nil
	}

	//////
	// Limit the values to only the ones that are relevant to the event type
	//////

	switch *treatment.EventType {
	case nightscout.Announcement:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Note:       note,
		}
	case nightscout.BGCheck:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			Note:           note,
		}
	case nightscout.BolusWizard:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			Note:           note,
		}
	case nightscout.CarbCorrection:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			FoodIntake: foodIntake,
			Note:       note,
		}
	case nightscout.ComboBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			Note:           note,
		}
	case nightscout.CorrectionBolus:
		return site.EventGroupsBundle{
			EventGroup:  eventGroup,
			InsulinDose: insulinDose,
			Note:        note,
		}
	case nightscout.DADAlert:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Note:       note,
		}
	case nightscout.Exercise:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Activity:   activity,
			Note:       note,
		}
	case nightscout.InsulinChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.MealBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			Note:           note,
		}
	case nightscout.Note:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Note:       note,
		}
	case nightscout.OpenAPSOffline:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.ProfileSwitch:
		return site.EventGroupsBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: profileSwitch,
			Note:          note,
		}
	case nightscout.PumpBatteryChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.Question:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			Note:       note,
		}
	case nightscout.SensorChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.SensorStart:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.SiteChange:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			ComponentEvent: componentEvent,
			Note:           note,
		}
	case nightscout.SnackBolus:
		return site.EventGroupsBundle{
			EventGroup:     eventGroup,
			GlucoseReading: glucoseReading,
			InsulinDose:    insulinDose,
			FoodIntake:     foodIntake,
			Note:           note,
		}
	case nightscout.TempBasal:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			TempBasal:  tempBasal,
			Note:       note,
		}
	case nightscout.TemporaryTarget:
		return site.EventGroupsBundle{
			EventGroup: eventGroup,
			TempTarget: tempTarget,
			Note:       note,
		}
	}

	return site.EventGroupsBundle{EventGroup: eventGroup}
}
