package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// READ

func EventGroupGet(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*site.EventGroupsBundle, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.GetEventGroup(ctx, site.GetEventGroupParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows || dbModel == nil {

		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get event group", "error", err)
		return nil, nsutils.ErrInternal
	}
	egb := site.GetEventGroupToEventGroupsBundle(*dbModel)
	return &egb, nil
}

func EventGroupSearch(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*site.EventGroupsBundle, *nsutils.HTTPError) {
	if query.TypeIn == nil {
		// TODO list all supportted types so we can limit to those that NS Treatment expects
	}

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

// CREATE
func EventGroupCreate(ctx context.Context, siteContext *httputils.SiteContext, bundle site.CreateEventGroupBundle) (*site.EventGroupsBundle, *nsutils.HTTPError) {
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	ret, err := createEventGroup(ctx, q, bundle)
	if err != nil {
		return nil, nsutils.ErrInternal
	}

	if ret == nil || ret.EventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", bundle)
		return nil, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create event group in database", "error", err, "dbModel", ret)
		return nil, nsutils.ErrInternal
	}

	return ret, nil
}

// UPDATE
func EventGroupUpdate(ctx context.Context, siteContext *httputils.SiteContext, bundle site.UpdateEventGroupBundle) (*site.EventGroupsBundle, *nsutils.HTTPError) {

	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}

	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	ret, err := updateEventGroup(ctx, q, bundle)
	if err != nil {
		slog.Error("failed in EventGroupUpdate to update event group in database", "error", err, "bundle", bundle)
		return nil, nsutils.ErrInternal
	}

	if ret == nil || ret.EventGroup.ID == "" {
		slog.Error("failed in EventGroupUpdate to update event group in database, EventGroup is nil", "bundle", bundle)
		return nil, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed in EventGroupUpdate to commit update event group in database", "error", err, "dbModel", ret)
		return nil, nsutils.ErrInternal
	}

	return ret, nil
}

func EventGroupReplace(ctx context.Context, siteContext *httputils.SiteContext, bundle site.CreateEventGroupBundle, canCreate bool) (*site.EventGroupsBundle, bool, *nsutils.HTTPError) {
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}

	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	createingNewRecord := false
	// DELETE THE RECORD
	err = hardDeleteEventGroup(ctx, q, bundle.EventGroup.ID)
	if err != nil {
		notFoundError := err == sql.ErrNoRows

		allowedToContinue := canCreate && notFoundError
		if !allowedToContinue && notFoundError {
			slog.Error("failed to delete event group in database", "error", err, "allowedToContinue", allowedToContinue, "canCreate", canCreate, "notFoundError", notFoundError)
			return nil, false, nsutils.ErrNotFound
		}
		if !allowedToContinue {
			slog.Error("failed to delete event group in database", "error", err, "allowedToContinue", allowedToContinue, "canCreate", canCreate, "notFoundError", notFoundError)
			return nil, false, nsutils.ErrInternal
		}
		createingNewRecord = true
	}

	// RECREATE THE NEW RECORD
	ret, err := createEventGroup(ctx, q, bundle)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}

	if ret == nil || ret.EventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", bundle)
		return nil, false, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit update event group in database", "error", err, "dbModel", ret)
		return nil, false, nsutils.ErrInternal
	}

	return ret, createingNewRecord, nil
}

// DELETES

func EventGroupSoftDelete(ctx context.Context, siteContext *httputils.SiteContext, identifier string) *nsutils.HTTPError {
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	err = softDeleteEventGroup(ctx, q, identifier)
	if err != nil {
		return nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit soft delete event group in database", "error", err)
		return nsutils.ErrInternal
	}

	return nil
}

func EventGroupHardDelete(ctx context.Context, siteContext *httputils.SiteContext, identifier string) *nsutils.HTTPError {
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	err = hardDeleteEventGroup(ctx, q, identifier)
	if err != nil {
		return nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit hard delete event group in database", "error", err)
		return nsutils.ErrInternal
	}

	return nil
}

func EventGroupSoftDeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func EventGroupHardDeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

// //////////////////////
// Private Methods
// ///////////////////////
func updateEventGroup(ctx context.Context, q *site.Queries, bundle site.UpdateEventGroupBundle) (*site.EventGroupsBundle, error) {
	ret := site.EventGroupsBundle{}

	if bundle.GlucoseReading != nil {
		dbModel, err := q.UpdateGlucoseReading(ctx, *bundle.GlucoseReading)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update glucose reading in database", "error", err, "updateParams", *bundle.GlucoseReading)
			return nil, err
		}
		ret.GlucoseReading = dbModel
	}
	if bundle.InsulinDose != nil {
		dbModel, err := q.UpdateInsulinDose(ctx, *bundle.InsulinDose)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update insulin dose in database", "error", err, "updateParams", *bundle.InsulinDose)
			return nil, err
		}
		ret.InsulinDose = dbModel
	}
	if bundle.FoodIntake != nil {
		dbModel, err := q.UpdateFoodIntake(ctx, *bundle.FoodIntake)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update food intake in database", "error", err, "updateParams", *bundle.FoodIntake)
			return nil, err
		}
		ret.FoodIntake = dbModel
	}
	if bundle.TempBasal != nil {
		dbModel, err := q.UpdateTempBasal(ctx, *bundle.TempBasal)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update temp basal in database", "error", err, "updateParams", *bundle.TempBasal)
			return nil, err
		}
		ret.TempBasal = dbModel
	}
	if bundle.TempTarget != nil {
		dbModel, err := q.UpdateTempTarget(ctx, *bundle.TempTarget)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update temp target in database", "error", err, "updateParams", *bundle.TempTarget)
			return nil, err
		}
		ret.TempTarget = dbModel
	}
	if bundle.ProfileSwitch != nil {
		dbModel, err := q.UpdateProfileSwitch(ctx, *bundle.ProfileSwitch)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update profile switch in database", "error", err, "updateParams", *bundle.ProfileSwitch)
			return nil, err
		}
		ret.ProfileSwitch = dbModel
	}
	if bundle.DeviceState != nil {
		dbModel, err := q.UpdateDeviceState(ctx, *bundle.DeviceState)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update device state in database", "error", err, "updateParams", *bundle.DeviceState)
			return nil, err
		}
		ret.DeviceState = dbModel
	}
	if bundle.LoopState != nil {
		dbModel, err := q.UpdateLoopState(ctx, *bundle.LoopState)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update loop state in database", "error", err, "updateParams", *bundle.LoopState)
			return nil, err
		}
		ret.LoopState = dbModel
	}
	if bundle.PumpState != nil {
		dbModel, err := q.UpdatePumpState(ctx, *bundle.PumpState)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update pump state in database", "error", err, "updateParams", *bundle.PumpState)
			return nil, err
		}
		ret.PumpState = dbModel
	}
	if bundle.CgmState != nil {
		dbModel, err := q.UpdateCGMState(ctx, *bundle.CgmState)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update cgm state in database", "error", err, "updateParams", *bundle.CgmState)
			return nil, err
		}
		ret.CgmState = dbModel
	}
	if bundle.Activity != nil {
		dbModel, err := q.UpdateActivity(ctx, *bundle.Activity)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update activity in database", "error", err, "updateParams", *bundle.Activity)
			return nil, err
		}
		ret.Activity = dbModel
	}
	if bundle.Note != nil {
		dbModel, err := q.UpdateNote(ctx, *bundle.Note)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update note in database", "error", err, "updateParams", *bundle.Note)
			return nil, err
		}
		ret.Note = dbModel
	}

	// Finaly Update the EventGroup
	dbModel, err := q.UpdateEventGroup(ctx, bundle.EventGroup)
	if err != nil {
		slog.Error("failed in updateEventGroup to update event group in database", "error", err, "updateParams", bundle.EventGroup)
		return nil, err
	}
	if dbModel == nil {
		slog.Error("failed in updateEventGroup to update event group in database, dbModel is nil", "updateParams", bundle.EventGroup)
		return nil, err
	}
	ret.EventGroup = *dbModel

	return &ret, nil
}

func createEventGroup(ctx context.Context, q *site.Queries, bundle site.CreateEventGroupBundle) (*site.EventGroupsBundle, error) {
	ret := site.EventGroupsBundle{}

	if bundle.GlucoseReading != nil {
		bundle.EventGroup.GlucoseReadingID = &bundle.GlucoseReading.ID
		dbModel, err := q.CreateGlucoseReading(ctx, *bundle.GlucoseReading)
		if err != nil || dbModel == nil {
			slog.Error("failed to create glucose reading in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.GlucoseReading = dbModel
	}
	if bundle.InsulinDose != nil {
		bundle.EventGroup.InsulinDoseID = &bundle.InsulinDose.ID
		dbModel, err := q.CreateInsulinDose(ctx, *bundle.InsulinDose)
		if err != nil || dbModel == nil {
			slog.Error("failed to create insulin dose in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.InsulinDose = dbModel
	}
	if bundle.FoodIntake != nil {
		bundle.EventGroup.FoodIntakeID = &bundle.FoodIntake.ID
		dbModel, err := q.CreateFoodIntake(ctx, *bundle.FoodIntake)
		if err != nil || dbModel == nil {
			slog.Error("failed to create food intake in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.FoodIntake = dbModel
	}
	if bundle.TempBasal != nil {
		bundle.EventGroup.TempBasalID = &bundle.TempBasal.ID
		dbModel, err := q.CreateTempBasal(ctx, *bundle.TempBasal)
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp basal in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.TempBasal = dbModel
	}
	if bundle.TempTarget != nil {
		bundle.EventGroup.TempTargetID = &bundle.TempTarget.ID
		dbModel, err := q.CreateTempTarget(ctx, *bundle.TempTarget)
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp target in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.TempTarget = dbModel
	}
	if bundle.ProfileSwitch != nil {
		bundle.EventGroup.ProfileSwitchID = &bundle.ProfileSwitch.ID
		dbModel, err := q.CreateProfileSwitch(ctx, *bundle.ProfileSwitch)
		if err != nil || dbModel == nil {
			slog.Error("failed to create profile switch in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.ProfileSwitch = dbModel
	}
	if bundle.DeviceState != nil {
		bundle.EventGroup.DeviceStateID = &bundle.DeviceState.ID
		dbModel, err := q.CreateDeviceState(ctx, *bundle.DeviceState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create device state in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.DeviceState = dbModel
	}
	if bundle.LoopState != nil {
		bundle.EventGroup.LoopStateID = &bundle.LoopState.ID
		dbModel, err := q.CreateLoopState(ctx, *bundle.LoopState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create loop state in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.LoopState = dbModel
	}
	if bundle.PumpState != nil {
		bundle.EventGroup.PumpStateID = &bundle.PumpState.ID
		dbModel, err := q.CreatePumpState(ctx, *bundle.PumpState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create pump state in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.PumpState = dbModel
	}
	if bundle.CgmState != nil {
		bundle.EventGroup.CgmStateID = &bundle.CgmState.ID
		dbModel, err := q.CreateCGMState(ctx, *bundle.CgmState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create cgm state in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.CgmState = dbModel
	}
	if bundle.Activity != nil {
		bundle.EventGroup.ActivityID = &bundle.Activity.ID
		dbModel, err := q.CreateActivity(ctx, *bundle.Activity)
		if err != nil || dbModel == nil {
			slog.Error("failed to create activity in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.Activity = dbModel
	}
	if bundle.Note != nil {
		bundle.EventGroup.NoteID = &bundle.Note.ID
		dbModel, err := q.CreateNote(ctx, *bundle.Note)
		if err != nil || dbModel == nil {
			slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.Note = dbModel
	}

	// Finaly Create the EventGroup
	dbModel, err := q.CreateEventGroup(ctx, bundle.EventGroup)
	if err != nil {
		slog.Error("failed to create event group in database", "error", err, "dbModel", dbModel)
		return nil, err
	}

	if dbModel == nil {
		slog.Error("failed to create event group in database, dbModel is nil", "bundle", bundle)
		return nil, nsutils.ErrInternal
	}

	ret.EventGroup = *dbModel

	return &ret, nil
}

func hardDeleteEventGroup(ctx context.Context, q *site.Queries, identifier string) error {
	eventGroup, err := q.HardDeleteEventGroup(ctx, site.HardDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to hard delete eventGroup event group by ID", "error", err)
		return err
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.HardDeleteGlucoseReading(ctx, site.HardDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.HardDeleteInsulinDose(ctx, site.HardDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.HardDeleteFoodIntake(ctx, site.HardDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.HardDeleteTempBasal(ctx, site.HardDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.HardDeleteTempTarget(ctx, site.HardDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.HardDeleteProfileSwitch(ctx, site.HardDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return err
		}
	}
	if eventGroup.DeviceStateID != nil {
		if res, err := q.HardDeleteDeviceState(ctx, site.HardDeleteDeviceStateParams{ID: *eventGroup.DeviceStateID}); err != nil {
			slog.Error("failed to delete device state by ID", "error", err, "deviceStateID", *eventGroup.DeviceStateID, "result", res)
			return err
		}
	}
	if eventGroup.LoopStateID != nil {
		if res, err := q.HardDeleteLoopState(ctx, site.HardDeleteLoopStateParams{ID: *eventGroup.LoopStateID}); err != nil {
			slog.Error("failed to delete loop state by ID", "error", err, "loopStateID", *eventGroup.LoopStateID, "result", res)
			return err
		}
	}
	if eventGroup.PumpStateID != nil {
		if res, err := q.HardDeletePumpState(ctx, site.HardDeletePumpStateParams{ID: *eventGroup.PumpStateID}); err != nil {
			slog.Error("failed to delete pump state by ID", "error", err, "pumpStateID", *eventGroup.PumpStateID, "result", res)
			return err
		}
	}
	if eventGroup.CgmStateID != nil {
		if res, err := q.HardDeleteCGMState(ctx, site.HardDeleteCGMStateParams{ID: *eventGroup.CgmStateID}); err != nil {
			slog.Error("failed to delete cgm state by ID", "error", err, "cgmStateID", *eventGroup.CgmStateID, "result", res)
			return err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.HardDeleteActivity(ctx, site.HardDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return err
		}
	}
	if eventGroup.NoteID != nil {
		if res, err := q.HardDeleteNote(ctx, site.HardDeleteNoteParams{ID: *eventGroup.NoteID}); err != nil {
			slog.Error("failed to delete note by ID", "error", err, "noteID", *eventGroup.NoteID, "result", res)
			return err
		}
	}

	return nil
}

func softDeleteEventGroup(ctx context.Context, q *site.Queries, identifier string) error {
	eventGroup, err := q.SoftDeleteEventGroup(ctx, site.SoftDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to soft delete eventGroup event group by ID", "error", err)
		return err
	}

	if eventGroup.GlucoseReadingID != nil {
		if res, err := q.SoftDeleteGlucoseReading(ctx, site.SoftDeleteGlucoseReadingParams{ID: *eventGroup.GlucoseReadingID}); err != nil {
			slog.Error("failed to soft delete glucose reading by ID", "error", err, "glucoseReadingID", *eventGroup.GlucoseReadingID, "result", res)
			return err
		}
	}
	if eventGroup.InsulinDoseID != nil {
		if res, err := q.SoftDeleteInsulinDose(ctx, site.SoftDeleteInsulinDoseParams{ID: *eventGroup.InsulinDoseID}); err != nil {
			slog.Error("failed to soft delete insulin dose by ID", "error", err, "insulinDoseID", *eventGroup.InsulinDoseID, "result", res)
			return err
		}
	}
	if eventGroup.FoodIntakeID != nil {
		if res, err := q.SoftDeleteFoodIntake(ctx, site.SoftDeleteFoodIntakeParams{ID: *eventGroup.FoodIntakeID}); err != nil {
			slog.Error("failed to soft delete food intake by ID", "error", err, "foodIntakeID", *eventGroup.FoodIntakeID, "result", res)
			return err
		}
	}
	if eventGroup.TempBasalID != nil {
		if res, err := q.SoftDeleteTempBasal(ctx, site.SoftDeleteTempBasalParams{ID: *eventGroup.TempBasalID}); err != nil {
			slog.Error("failed to soft delete temp basal by ID", "error", err, "tempBasalID", *eventGroup.TempBasalID, "result", res)
			return err
		}
	}
	if eventGroup.TempTargetID != nil {
		if res, err := q.SoftDeleteTempTarget(ctx, site.SoftDeleteTempTargetParams{ID: *eventGroup.TempTargetID}); err != nil {
			slog.Error("failed to soft delete temp target by ID", "error", err, "tempTargetID", *eventGroup.TempTargetID, "result", res)
			return err
		}
	}
	if eventGroup.ProfileSwitchID != nil {
		if res, err := q.SoftDeleteProfileSwitch(ctx, site.SoftDeleteProfileSwitchParams{ID: *eventGroup.ProfileSwitchID}); err != nil {
			slog.Error("failed to soft delete profile switch by ID", "error", err, "profileSwitchID", *eventGroup.ProfileSwitchID, "result", res)
			return err
		}
	}
	if eventGroup.DeviceStateID != nil {
		if res, err := q.SoftDeleteDeviceState(ctx, site.SoftDeleteDeviceStateParams{ID: *eventGroup.DeviceStateID}); err != nil {
			slog.Error("failed to soft delete device state by ID", "error", err, "deviceStateID", *eventGroup.DeviceStateID, "result", res)
			return err
		}
	}
	if eventGroup.LoopStateID != nil {
		if res, err := q.SoftDeleteLoopState(ctx, site.SoftDeleteLoopStateParams{ID: *eventGroup.LoopStateID}); err != nil {
			slog.Error("failed to soft delete loop state by ID", "error", err, "loopStateID", *eventGroup.LoopStateID, "result", res)
			return err
		}
	}
	if eventGroup.PumpStateID != nil {
		if res, err := q.SoftDeletePumpState(ctx, site.SoftDeletePumpStateParams{ID: *eventGroup.PumpStateID}); err != nil {
			slog.Error("failed to soft delete pump state by ID", "error", err, "pumpStateID", *eventGroup.PumpStateID, "result", res)
			return err
		}
	}
	if eventGroup.CgmStateID != nil {
		if res, err := q.SoftDeleteCGMState(ctx, site.SoftDeleteCGMStateParams{ID: *eventGroup.CgmStateID}); err != nil {
			slog.Error("failed to soft delete cgm state by ID", "error", err, "cgmStateID", *eventGroup.CgmStateID, "result", res)
			return err
		}
	}
	if eventGroup.ActivityID != nil {
		if res, err := q.SoftDeleteActivity(ctx, site.SoftDeleteActivityParams{ID: *eventGroup.ActivityID}); err != nil {
			slog.Error("failed to soft delete activity by ID", "error", err, "activityID", *eventGroup.ActivityID, "result", res)
			return err
		}
	}
	if eventGroup.NoteID != nil {
		if res, err := q.SoftDeleteNote(ctx, site.SoftDeleteNoteParams{ID: *eventGroup.NoteID}); err != nil {
			slog.Error("failed to soft delete note by ID", "error", err, "noteID", *eventGroup.NoteID, "result", res)
			return err
		}
	}

	return nil
}
