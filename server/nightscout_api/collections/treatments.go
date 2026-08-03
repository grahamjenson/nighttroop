package collections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	converter func(c *site.GetEventGroupRow) nsutils.Record
}

// There is a list of incoming types, but we only support a few, and some we break apart into simpler types

// Announcement
// BGCheck
// BolusWizard
// CarbCorrection
// ComboBolus
// CorrectionBolus
// DADAlert
// Exercise
// InsulinChange
// MealBolus
// Note
// OpenAPSOffline
// ProfileSwitch
// PumpBatteryChange
// Question
// SensorChange
// SensorStart
// SiteChange
// SnackBolus
// TempBasal
// TemporaryTarget

// Treatments are a constallation of DB types and no more.
func NewTreatmentCollection() TreatmentCollection {
	return TreatmentCollection{
		converter: nsutils.DBModelToRecord(nsutils.Treatments, nightscout.DBModelToTreatment),
	}
}

func (c TreatmentCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(data, &treatmentModel)
	if err != nil {
		slog.Error("failed to map data to treatment model", "error", err)
		return nil, nsutils.ErrInternal
	}

	if treatmentModel.EventType == nil || *treatmentModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is nil")
		return nil, nsutils.ErrInternal
	}

	cc := nightscout.CreateContext{
		IDPrefix:              "ns",
		CreatedByIdentityID:   siteContext.IdentityModel.ID,
		CreatedByIdentityName: siteContext.IdentityModel.Name,
	}

	dbModel, err := EventGroupCreate(ctx, siteContext, TreatmentEventGroupCreate(ctx, cc, treatmentModel))
	if err != nil {
		slog.Error("failed to create treatment in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create treatment in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), nil
}

// Update will not create a new DB record, and only update existing records on the event group.
func (c TreatmentCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(patch, &treatmentModel)
	if err != nil {
		slog.Error("failed to map patch to treatment model", "error", err)
		return nil, nsutils.ErrInternal
	}

	cc := nightscout.UpdateContext{
		UpdatedByIdentityID:   siteContext.IdentityModel.ID,
		UpdatedByIdentityName: siteContext.IdentityModel.Name,
	}

	dbModel, err := EventGroupUpdate(ctx, siteContext, TreatmentEventGroupUpdate(ctx, cc, treatmentModel))
	if err != nil {
		slog.Error("failed to update treatment in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to update treatment in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbModel)), nil
}

// ///////
// Default EventGroup stuff
// ///////

func (c TreatmentCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	seg, rerr := EventGroupGet(ctx, siteContext, identifier)
	if rerr != nil {
		return nil, rerr
	}
	return ptr(c.converter(seg)), nil
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	treatments, rerr := EventGroupSearch(ctx, siteContext, query)
	if rerr != nil {
		return nil, rerr
	}

	return nsutils.CreateRecordSlice(treatments, c.converter), nil
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return EventGroupDelete(ctx, siteContext, identifier, permanent)
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return EventGroupDeleteType(ctx, siteContext, typeString, permanent)
}

func TreatmentEventGroupCreate(ctx context.Context, cc nightscout.CreateContext, treatment nightscout.Treatment) site.CreateEventGroupBundle {

	eventGroup := nightscout.TreatmentToCreateEventGroup(treatment, cc)

	glucoseReading := nightscout.TreatmentToCreateGlucoseReading(treatment, cc)
	insulinDose := nightscout.TreatmentToCreateInsulinDose(treatment, cc)
	foodIntake := nightscout.TreatmentToCreateFoodIntake(treatment, cc)
	tempBasal := nightscout.TreatmentToCreateTempBasal(treatment, cc)
	tempTarget := nightscout.TreatmentToCreateTempTarget(treatment, cc)
	profileSwitch := nightscout.TreatmentToCreateProfileSwitch(treatment, cc)
	deviceState := nightscout.TreatmentToCreateDeviceState(treatment, cc)
	loopState := nightscout.TreatmentToCreateLoopState(treatment, cc)
	pumpState := nightscout.TreatmentToCreatePumpState(treatment, cc)
	cgmState := nightscout.TreatmentToCreateCgmState(treatment, cc)
	activity := nightscout.TreatmentToCreateActivity(treatment, cc)
	note := nightscout.TreatmentToCreateNote(treatment, cc)

	switch *treatment.EventType {
	case nightscout.Announcement:
	case nightscout.BGCheck:
	case nightscout.BolusWizard:
	case nightscout.CarbCorrection:
	case nightscout.ComboBolus:
	case nightscout.CorrectionBolus:
	case nightscout.DADAlert:
	case nightscout.Exercise:
	case nightscout.InsulinChange:
	case nightscout.MealBolus:
	case nightscout.Note:
	case nightscout.OpenAPSOffline:
	case nightscout.ProfileSwitch:
	case nightscout.PumpBatteryChange:
	case nightscout.Question:
	case nightscout.SensorChange:
	case nightscout.SensorStart:
	case nightscout.SiteChange:
	case nightscout.SnackBolus:
	case nightscout.TempBasal:
	case nightscout.TemporaryTarget:
	case nightscout.Unknown:
	}

	return site.CreateEventGroupBundle{
		EventGroup:     eventGroup,
		GlucoseReading: &glucoseReading,
		InsulinDose:    &insulinDose,
		FoodIntake:     &foodIntake,
		TempBasal:      &tempBasal,
		TempTarget:     &tempTarget,
		ProfileSwitch:  &profileSwitch,
		DeviceState:    &deviceState,
		LoopState:      &loopState,
		PumpState:      &pumpState,
		CgmState:       &cgmState,
		Activity:       &activity,
		Note:           &note,
	}
}

func TreatmentEventGroupUpdate(ctx context.Context, cc nightscout.UpdateContext, treatment nightscout.Treatment) site.UpdateEventGroupBundle {

	eventGroup := nightscout.TreatmentToUpdateEventGroup(treatment, cc)

	glucoseReading := nightscout.TreatmentToUpdateGlucoseReading(treatment, cc)
	insulinDose := nightscout.TreatmentToUpdateInsulinDose(treatment, cc)
	foodIntake := nightscout.TreatmentToUpdateFoodIntake(treatment, cc)
	tempBasal := nightscout.TreatmentToUpdateTempBasal(treatment, cc)
	tempTarget := nightscout.TreatmentToUpdateTempTarget(treatment, cc)
	profileSwitch := nightscout.TreatmentToUpdateProfileSwitch(treatment, cc)
	deviceState := nightscout.TreatmentToUpdateDeviceState(treatment, cc)
	loopState := nightscout.TreatmentToUpdateLoopState(treatment, cc)
	pumpState := nightscout.TreatmentToUpdatePumpState(treatment, cc)
	cgmState := nightscout.TreatmentToUpdateCgmState(treatment, cc)
	activity := nightscout.TreatmentToUpdateActivity(treatment, cc)
	note := nightscout.TreatmentToUpdateNote(treatment, cc)

	return site.UpdateEventGroupBundle{
		EventGroup:     eventGroup,
		GlucoseReading: &glucoseReading,
		InsulinDose:    &insulinDose,
		FoodIntake:     &foodIntake,
		TempBasal:      &tempBasal,
		TempTarget:     &tempTarget,
		ProfileSwitch:  &profileSwitch,
		DeviceState:    &deviceState,
		LoopState:      &loopState,
		PumpState:      &pumpState,
		CgmState:       &cgmState,
		Activity:       &activity,
		Note:           &note,
	}
}
