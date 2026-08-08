package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	converter func(c *site.EventGroupsBundle) nsutils.Record
}

// There is a list of incoming types, but we only support a few, and some we break apart into simpler types
// Treatments are a constallation of DB types and no more.
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

func (c TreatmentCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, cc nightscout.CreateContext, replacing *site.EventGroupsBundle, overrideIdentifier *string) (*site.EventGroupsBundle, error) {
	if nsModel.EventType == nil || *nsModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is nil")
		return nil, nsutils.BadDataBodyError
	}

	createParams := TreatmentEventGroupCreate(ctx, cc, nsModel)

	if overrideIdentifier != nil {
		createParams.EventGroup.ID = *overrideIdentifier
	}

	dbModel, rerr := EventGroupCreate(ctx, q, createParams)

	return dbModel, rerr
}

func (c TreatmentCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, site.GetEventGroupParams{ID: identifier})
	if err != nil {
		return nil, err
	}
	egb := site.GetEventGroupToEventGroupsBundle(*dbModel)
	return &egb, nil
}

func (c TreatmentCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Treatment, identifier string, uc nightscout.UpdateContext) (*site.EventGroupsBundle, error) {
	if nsModel.EventType == nil || *nsModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is nil")
		return nil, nsutils.BadDataBodyError
	}

	return EventGroupUpdate(ctx, q, TreatmentEventGroupUpdate(ctx, uc, nsModel))
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

func (c TreatmentCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch)
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	treatments, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(treatments, c.converter), nil
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	if permanent {
		return EventGroupHardDelete(ctx, siteContext, identifier)
	}
	return EventGroupSoftDelete(ctx, siteContext, identifier)
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {

	if permanent {
		return EventGroupHardDeleteType(ctx, siteContext, typeString)
	}
	return EventGroupSoftDeleteType(ctx, siteContext, typeString)
}

// Replace will replace (delete, then insert) if it exists, or create it if it doesnt
func (c TreatmentCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(patch, &treatmentModel)
	if err != nil {
		slog.Error("failed to map patch to treatment model", "error", err)
		return nil, false, nsutils.BadDataBodyError
	}

	cc := nightscout.CreateContext{
		CreatedByIdentityID:   siteContext.IdentityModel.ID,
		CreatedByIdentityName: siteContext.IdentityModel.Name,
	}

	if treatmentModel.EventType == nil || *treatmentModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is nil")
		return nil, false, nsutils.BadDataBodyError
	}
	createParams := TreatmentEventGroupCreate(ctx, cc, treatmentModel)
	createParams.EventGroup.ID = identifier
	dbModel, created, rerr := EventGroupReplace(ctx, siteContext, createParams, canCreate)
	if rerr != nil {
		slog.Error("failed to update treatment in database", "error", rerr)
		return nil, false, rerr
	}
	if dbModel == nil {
		slog.Error("failed to update treatment in database, dbModel is nil")
		return nil, false, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), created, nil
}

func TreatmentEventGroupCreate(ctx context.Context, cc nightscout.CreateContext, treatment nightscout.Treatment) site.CreateEventGroupBundle {
	eventGroup := nightscout.TreatmentToCreateEventGroup(treatment, cc)

	glucoseReading := nightscout.TreatmentToCreateGlucoseReading(treatment, cc)
	insulinDose := nightscout.TreatmentToCreateInsulinDose(treatment, cc)
	foodIntake := nightscout.TreatmentToCreateFoodIntake(treatment, cc)
	tempBasal := nightscout.TreatmentToCreateTempBasal(treatment, cc)
	tempTarget := nightscout.TreatmentToCreateTempTarget(treatment, cc)
	profileSwitch := nightscout.TreatmentToCreateProfileSwitch(treatment, cc)
	componentEvent := nightscout.TreatmentToCreateComponentEvent(treatment, cc)
	activity := nightscout.TreatmentToCreateActivity(treatment, cc)
	note := nightscout.TreatmentToCreateNote(treatment, cc)

	switch *treatment.EventType {
	case nightscout.Announcement:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.BGCheck:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			Note:           &note,
		}
	case nightscout.BolusWizard:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.CarbCorrection:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			FoodIntake: &foodIntake,
			Note:       &note,
		}
	case nightscout.ComboBolus:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.CorrectionBolus:
		return site.CreateEventGroupBundle{
			EventGroup:  eventGroup,
			InsulinDose: &insulinDose,
			Note:        &note,
		}
	case nightscout.DADAlert:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.Exercise:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Activity:   &activity,
			Note:       &note,
		}
	case nightscout.InsulinChange:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.MealBolus:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.Note:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.OpenAPSOffline:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.ProfileSwitch:
		return site.CreateEventGroupBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: &profileSwitch,
			Note:          &note,
		}
	case nightscout.PumpBatteryChange:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.Question:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.SensorChange:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SensorStart:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SiteChange:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SnackBolus:
		return site.CreateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.TempBasal:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			TempBasal:  &tempBasal,
			Note:       &note,
		}
	case nightscout.TemporaryTarget:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			TempTarget: &tempTarget,
			Note:       &note,
		}
	}

	return site.CreateEventGroupBundle{}
}

func TreatmentEventGroupUpdate(ctx context.Context, cc nightscout.UpdateContext, treatment nightscout.Treatment) site.UpdateEventGroupBundle {
	eventGroup := nightscout.TreatmentToUpdateEventGroup(treatment, cc)

	glucoseReading := nightscout.TreatmentToUpdateGlucoseReading(treatment, cc)
	insulinDose := nightscout.TreatmentToUpdateInsulinDose(treatment, cc)
	foodIntake := nightscout.TreatmentToUpdateFoodIntake(treatment, cc)
	tempBasal := nightscout.TreatmentToUpdateTempBasal(treatment, cc)
	tempTarget := nightscout.TreatmentToUpdateTempTarget(treatment, cc)
	profileSwitch := nightscout.TreatmentToUpdateProfileSwitch(treatment, cc)
	componentEvent := nightscout.TreatmentToUpdateComponentEvent(treatment, cc)
	activity := nightscout.TreatmentToUpdateActivity(treatment, cc)
	note := nightscout.TreatmentToUpdateNote(treatment, cc)

	switch *treatment.EventType {
	case nightscout.Announcement:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.BGCheck:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			Note:           &note,
		}
	case nightscout.BolusWizard:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.CarbCorrection:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			FoodIntake: &foodIntake,
			Note:       &note,
		}
	case nightscout.ComboBolus:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.CorrectionBolus:
		return site.UpdateEventGroupBundle{
			EventGroup:  eventGroup,
			InsulinDose: &insulinDose,
			Note:        &note,
		}
	case nightscout.DADAlert:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.Exercise:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Activity:   &activity,
			Note:       &note,
		}
	case nightscout.InsulinChange:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.MealBolus:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.Note:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.OpenAPSOffline:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.ProfileSwitch:
		return site.UpdateEventGroupBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: &profileSwitch,
			Note:          &note,
		}
	case nightscout.PumpBatteryChange:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.Question:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.SensorChange:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SensorStart:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SiteChange:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			ComponentEvent: &componentEvent,
			Note:           &note,
		}
	case nightscout.SnackBolus:
		return site.UpdateEventGroupBundle{
			EventGroup:     eventGroup,
			GlucoseReading: &glucoseReading,
			InsulinDose:    &insulinDose,
			FoodIntake:     &foodIntake,
			Note:           &note,
		}
	case nightscout.TempBasal:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			TempBasal:  &tempBasal,
			Note:       &note,
		}
	case nightscout.TemporaryTarget:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			TempTarget: &tempTarget,
			Note:       &note,
		}
	}

	return site.UpdateEventGroupBundle{}
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
func EventGroupCreate(ctx context.Context, q *site.Queries, bundle site.CreateEventGroupBundle) (*site.EventGroupsBundle, error) {

	ret, err := createEventGroup(ctx, q, bundle)
	if err != nil {
		return nil, err
	}

	if ret == nil || ret.EventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", bundle)
		return nil, err
	}

	return ret, nil
}

// UPDATE
func EventGroupUpdate(ctx context.Context, q *site.Queries, bundle site.UpdateEventGroupBundle) (*site.EventGroupsBundle, *nsutils.HTTPError) {

	ret, err := updateEventGroup(ctx, q, bundle)
	if err != nil {
		slog.Error("failed in EventGroupUpdate to update event group in database", "error", err, "bundle", bundle)
		return nil, nsutils.ErrInternal
	}

	if ret == nil || ret.EventGroup.ID == "" {
		slog.Error("failed in EventGroupUpdate to update event group in database, EventGroup is nil", "bundle", bundle)
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
	if bundle.ComponentEvent != nil {
		dbModel, err := q.UpdateComponentEvent(ctx, *bundle.ComponentEvent)
		if err != nil || dbModel == nil {
			slog.Error("failed in updateEventGroup to update component event in database", "error", err, "updateParams", *bundle.ComponentEvent)
			return nil, err
		}
		ret.ComponentEvent = dbModel
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
	if bundle.ComponentEvent != nil {
		bundle.EventGroup.ComponentEventID = &bundle.ComponentEvent.ID
		dbModel, err := q.CreateComponentEvent(ctx, *bundle.ComponentEvent)
		if err != nil || dbModel == nil {
			slog.Error("failed to create component event in database", "error", err, "dbModel", dbModel)
			return nil, err
		}
		ret.ComponentEvent = dbModel
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
	if eventGroup.ComponentEventID != nil {
		if res, err := q.HardDeleteComponentEvent(ctx, site.HardDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to delete component event by ID", "error", err, "deviceStateID", *eventGroup.ComponentEventID, "result", res)
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
	if eventGroup.ComponentEventID != nil {
		if res, err := q.SoftDeleteComponentEvent(ctx, site.SoftDeleteComponentEventParams{ID: *eventGroup.ComponentEventID}); err != nil {
			slog.Error("failed to soft delete component event by ID", "error", err, "deviceStateID", *eventGroup.ComponentEventID, "result", res)
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
