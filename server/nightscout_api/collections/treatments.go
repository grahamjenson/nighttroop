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
	converter func(c *site.EventGroupsBundle) nsutils.Record
}

// There is a list of incoming types, but we only support a few, and some we break apart into simpler types
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
		return nil, nsutils.BadDataBodyError
	}

	if treatmentModel.EventType == nil || *treatmentModel.EventType == nightscout.Unknown {
		slog.Error("treatment event type is nil")
		return nil, nsutils.BadDataBodyError
	}

	cc := nightscout.CreateContext{
		CreatedByIdentityID:   siteContext.IdentityModel.ID,
		CreatedByIdentityName: siteContext.IdentityModel.Name,
	}

	dbModel, rerr := EventGroupCreate(ctx, siteContext, TreatmentEventGroupCreate(ctx, cc, treatmentModel))
	if rerr != nil {
		slog.Error("failed to create treatment in database", "error", rerr)
		return nil, rerr
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
		return nil, nsutils.BadDataBodyError
	}

	cc := nightscout.UpdateContext{
		UpdatedByIdentityID:   siteContext.IdentityModel.ID,
		UpdatedByIdentityName: siteContext.IdentityModel.Name,
	}

	dbModel, rerr := EventGroupUpdate(ctx, siteContext, TreatmentEventGroupUpdate(ctx, cc, treatmentModel))
	if rerr != nil {
		slog.Error("failed to update treatment in database", "error", rerr)
		return nil, rerr
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
	//deviceState := nightscout.TreatmentToCreateDeviceState(treatment, cc)
	loopState := nightscout.TreatmentToCreateLoopState(treatment, cc)
	pumpState := nightscout.TreatmentToCreatePumpState(treatment, cc)
	cgmState := nightscout.TreatmentToCreateCgmState(treatment, cc)
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
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
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
			EventGroup: eventGroup,
			LoopState:  &loopState,
			Note:       &note,
		}
	case nightscout.ProfileSwitch:
		return site.CreateEventGroupBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: &profileSwitch,
			Note:          &note,
		}
	case nightscout.PumpBatteryChange:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
		}
	case nightscout.Question:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.SensorChange:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			CgmState:   &cgmState,
			Note:       &note,
		}
	case nightscout.SensorStart:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			CgmState:   &cgmState,
			Note:       &note,
		}
	case nightscout.SiteChange:
		return site.CreateEventGroupBundle{
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
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
	//deviceState := nightscout.TreatmentToUpdateDeviceState(treatment, cc)
	loopState := nightscout.TreatmentToUpdateLoopState(treatment, cc)
	pumpState := nightscout.TreatmentToUpdatePumpState(treatment, cc)
	cgmState := nightscout.TreatmentToUpdateCgmState(treatment, cc)
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
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
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
			EventGroup: eventGroup,
			LoopState:  &loopState,
			Note:       &note,
		}
	case nightscout.ProfileSwitch:
		return site.UpdateEventGroupBundle{
			EventGroup:    eventGroup,
			ProfileSwitch: &profileSwitch,
			Note:          &note,
		}
	case nightscout.PumpBatteryChange:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
		}
	case nightscout.Question:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			Note:       &note,
		}
	case nightscout.SensorChange:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			CgmState:   &cgmState,
			Note:       &note,
		}
	case nightscout.SensorStart:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			CgmState:   &cgmState,
			Note:       &note,
		}
	case nightscout.SiteChange:
		return site.UpdateEventGroupBundle{
			EventGroup: eventGroup,
			PumpState:  &pumpState,
			Note:       &note,
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
