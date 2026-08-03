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

func EventGroupGet(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*site.GetEventGroupRow, *nsutils.HTTPError) {
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

	return dbModel, nil
}

func EventGroupSearch(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*site.GetEventGroupRow, *nsutils.HTTPError) {
	if query.TypeIn == nil {
		// TODO list all supportted types so we can limit to those that NS Treatment expects
	}

	eventGroups, rerr := siteContext.SiteDB.Queries.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if rerr != nil {
		slog.Error("failed to query event groups from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	res := []*site.GetEventGroupRow{}
	for _, eg := range eventGroups {
		if eg != nil {
			gEG := site.SearchEventGroupToGetEventGroup(*eg)
			res = append(res, &gEG)
		}
	}
	return res, nil

}

func EventGroupDelete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteEventGroup(ctx, site.HardDeleteEventGroupParams{ID: identifier})
		if err != nil {
			slog.Error("failed to permanently delete treatment event group by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for treatment event group delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Deleted treatment event group", "eid", identifier, "rowsAffected", rowsAffected)
		return nil

	}

	res, err := siteContext.SiteDB.Queries.SoftDeleteEventGroup(ctx, site.SoftDeleteEventGroupParams{ID: identifier, UpdatedByIdentityName: &siteContext.IdentityModel.Name})
	if err != nil {
		slog.Error("failed to soft delete treatment event group by ID", "error", err)
		return nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for treatment event group delete", "error", err)
		return nsutils.ErrInternal
	}
	if rowsAffected == 0 {
		return nsutils.ErrNotFound
	}
	slog.Info("Soft Deleted treatment event group", "eid", identifier, "rowsAffected", rowsAffected)

	return nil
}

func EventGroupDeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {

	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteEventGroupsByType(ctx, site.HardDeleteEventGroupsByTypeParams{GroupType: typeString})
		if err != nil {
			slog.Error("failed to permanently delete treatment event group by type", "error", err)
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for treatment event group delete", "error", err)
			return 0, nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return 0, nsutils.ErrNotFound
		}
		slog.Info("Deleted treatment event group", "typeString", typeString, "rowsAffected", rowsAffected)

		return int(rowsAffected), nil

	}

	res, err := siteContext.SiteDB.Queries.SoftDeleteEventGroupsByType(ctx, site.SoftDeleteEventGroupsByTypeParams{GroupType: typeString, UpdatedByIdentityName: &siteContext.IdentityModel.Name})
	if err != nil {
		slog.Error("failed to soft delete treatment event group by type", "error", err)
		return 0, nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for treatment event group delete", "error", err)
		return 0, nsutils.ErrInternal
	}
	if rowsAffected == 0 {
		return 0, nsutils.ErrNotFound
	}
	slog.Info("Soft Deleted treatment event group", "typeString", typeString, "rowsAffected", rowsAffected)

	return int(rowsAffected), nil
}

func EventGroupCreate(ctx context.Context, siteContext *httputils.SiteContext, bundle site.CreateEventGroupBundle) (*site.GetEventGroupRow, *nsutils.HTTPError) {
	ret := site.GetEventGroupRow{}

	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}

	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	if bundle.GlucoseReading != nil {
		dbModel, err := q.CreateGlucoseReading(ctx, *bundle.GlucoseReading)
		if err != nil || dbModel == nil {
			slog.Error("failed to create glucose reading in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.GlucoseReading = *dbModel
	}
	if bundle.InsulinDose != nil {
		dbModel, err := q.CreateInsulinDose(ctx, *bundle.InsulinDose)
		if err != nil || dbModel == nil {
			slog.Error("failed to create insulin dose in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.InsulinDose = *dbModel
	}
	if bundle.FoodIntake != nil {
		dbModel, err := q.CreateFoodIntake(ctx, *bundle.FoodIntake)
		if err != nil || dbModel == nil {
			slog.Error("failed to create food intake in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.FoodIntake = *dbModel
	}
	if bundle.TempBasal != nil {
		dbModel, err := q.CreateTempBasal(ctx, *bundle.TempBasal)
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp basal in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.TempBasal = *dbModel
	}
	if bundle.TempTarget != nil {
		dbModel, err := q.CreateTempTarget(ctx, *bundle.TempTarget)
		if err != nil || dbModel == nil {
			slog.Error("failed to create temp target in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.TempTarget = *dbModel
	}
	if bundle.ProfileSwitch != nil {
		dbModel, err := q.CreateProfileSwitch(ctx, *bundle.ProfileSwitch)
		if err != nil || dbModel == nil {
			slog.Error("failed to create profile switch in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.ProfileSwitch = *dbModel
	}
	if bundle.DeviceState != nil {
		dbModel, err := q.CreateDeviceState(ctx, *bundle.DeviceState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create device state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.DeviceState = *dbModel
	}
	if bundle.LoopState != nil {
		dbModel, err := q.CreateLoopState(ctx, *bundle.LoopState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create loop state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.LoopState = *dbModel
	}
	if bundle.PumpState != nil {
		dbModel, err := q.CreatePumpState(ctx, *bundle.PumpState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create pump state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.PumpState = *dbModel
	}
	if bundle.CgmState != nil {
		dbModel, err := q.CreateCGMState(ctx, *bundle.CgmState)
		if err != nil || dbModel == nil {
			slog.Error("failed to create cgm state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.CgmState = *dbModel
	}
	if bundle.Activity != nil {
		dbModel, err := q.CreateActivity(ctx, *bundle.Activity)
		if err != nil || dbModel == nil {
			slog.Error("failed to create activity in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.Activity = *dbModel
	}
	if bundle.Note != nil {
		dbModel, err := q.CreateNote(ctx, *bundle.Note)
		if err != nil || dbModel == nil {
			slog.Error("failed to create note in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.Note = *dbModel
	}

	// Finaly Create the EventGroup
	dbModel, err := q.CreateEventGroup(ctx, bundle.EventGroup)
	if err != nil || dbModel == nil {
		slog.Error("failed to create event group in database", "error", err, "dbModel", dbModel)
		return nil, nsutils.ErrInternal
	}
	ret.EventGroup = *dbModel

	if err := tx.Commit(); err != nil {
		return nil, nsutils.ErrInternal
	}

	return &ret, nil
}

func EventGroupUpdate(ctx context.Context, siteContext *httputils.SiteContext, bundle site.UpdateEventGroupBundle) (*site.GetEventGroupRow, *nsutils.HTTPError) {
	ret := site.GetEventGroupRow{}

	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}

	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	if bundle.GlucoseReading != nil {
		dbModel, err := q.UpdateGlucoseReading(ctx, *bundle.GlucoseReading)
		if err != nil || dbModel == nil {
			slog.Error("failed to update glucose reading in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.GlucoseReading = *dbModel
	}
	if bundle.InsulinDose != nil {
		dbModel, err := q.UpdateInsulinDose(ctx, *bundle.InsulinDose)
		if err != nil || dbModel == nil {
			slog.Error("failed to update insulin dose in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.InsulinDose = *dbModel
	}
	if bundle.FoodIntake != nil {
		dbModel, err := q.UpdateFoodIntake(ctx, *bundle.FoodIntake)
		if err != nil || dbModel == nil {
			slog.Error("failed to update food intake in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.FoodIntake = *dbModel
	}
	if bundle.TempBasal != nil {
		dbModel, err := q.UpdateTempBasal(ctx, *bundle.TempBasal)
		if err != nil || dbModel == nil {
			slog.Error("failed to update temp basal in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.TempBasal = *dbModel
	}
	if bundle.TempTarget != nil {
		dbModel, err := q.UpdateTempTarget(ctx, *bundle.TempTarget)
		if err != nil || dbModel == nil {
			slog.Error("failed to update temp target in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.TempTarget = *dbModel
	}
	if bundle.ProfileSwitch != nil {
		dbModel, err := q.UpdateProfileSwitch(ctx, *bundle.ProfileSwitch)
		if err != nil || dbModel == nil {
			slog.Error("failed to update profile switch in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.ProfileSwitch = *dbModel
	}
	if bundle.DeviceState != nil {
		dbModel, err := q.UpdateDeviceState(ctx, *bundle.DeviceState)
		if err != nil || dbModel == nil {
			slog.Error("failed to update device state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.DeviceState = *dbModel
	}
	if bundle.LoopState != nil {
		dbModel, err := q.UpdateLoopState(ctx, *bundle.LoopState)
		if err != nil || dbModel == nil {
			slog.Error("failed to update loop state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.LoopState = *dbModel
	}
	if bundle.PumpState != nil {
		dbModel, err := q.UpdatePumpState(ctx, *bundle.PumpState)
		if err != nil || dbModel == nil {
			slog.Error("failed to update pump state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.PumpState = *dbModel
	}
	if bundle.CgmState != nil {
		dbModel, err := q.UpdateCGMState(ctx, *bundle.CgmState)
		if err != nil || dbModel == nil {
			slog.Error("failed to update cgm state in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.CgmState = *dbModel
	}
	if bundle.Activity != nil {
		dbModel, err := q.UpdateActivity(ctx, *bundle.Activity)
		if err != nil || dbModel == nil {
			slog.Error("failed to update activity in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.Activity = *dbModel
	}
	if bundle.Note != nil {
		dbModel, err := q.UpdateNote(ctx, *bundle.Note)
		if err != nil || dbModel == nil {
			slog.Error("failed to update note in database", "error", err, "dbModel", dbModel)
			return nil, nsutils.ErrInternal
		}
		ret.Note = *dbModel
	}

	// Finaly Update the EventGroup
	dbModel, err := q.UpdateEventGroup(ctx, bundle.EventGroup)
	if err != nil || dbModel == nil {
		slog.Error("failed to update event group in database", "error", err, "dbModel", dbModel)
		return nil, nsutils.ErrInternal
	}
	ret.EventGroup = *dbModel

	if err := tx.Commit(); err != nil {
		return nil, nsutils.ErrInternal
	}

	return &ret, nil
}
