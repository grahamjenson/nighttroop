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
	carbConverter       func(c *site.Carb) nsutils.Record
	tempTargetConverter func(c *site.TempTarget) nsutils.Record
	tempBasalConverter  func(c *site.TempBasal) nsutils.Record
}

func NewTreatmentCollection() TreatmentCollection {
	return TreatmentCollection{
		carbConverter:       DBModelToRecord(nsutils.Treatments, nightscout.CarbsDBModelToTreatment),
		tempTargetConverter: DBModelToRecord(nsutils.Treatments, nightscout.TempTargetDBModelToTreatment),
		tempBasalConverter:  DBModelToRecord(nsutils.Treatments, nightscout.TempBasalDBModelToTreatment),
	}
}

// Announcement
// BGCheck
// SnackBolus
// CorrectionBolus
// ComboBolus
// TemporaryTarget
// MealBolus
// TempBasal
// Note
// Exercise
// SiteChange
// SensorStart
// SensorChange
// InsulinChange
// PumpBatteryChange
// Question
// ProfileSwitch
// CarbCorrection
// DADAlert
// OpenAPSOffline
// BolusWizard

func (c TreatmentCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(data, &treatmentModel)
	if err != nil {
		slog.Error("failed to map data to treatment model", "error", err)
		return nil, nsutils.ErrInternal
	}
	if treatmentModel.EventType == nil {
		slog.Error("treatment model event type is nil")
		return nil, nsutils.ErrInternal
	}
	et := *treatmentModel.EventType
	switch et {
	case nightscout.Announcement:
		break
	case nightscout.BGCheck:
		break
	case nightscout.SnackBolus:
		break
	case nightscout.CorrectionBolus:
		break
	case nightscout.ComboBolus:
		break
	case nightscout.TemporaryTarget:
		dbModel, err := siteContext.SiteDB.Queries.CreateTempTarget(ctx,
			nightscout.TreatmentToCreateTempTarget(treatmentModel, nightscout.CreateContext{
				IDPrefix:              et.ToIDPrefix(),
				CreatedByIdentityID:   siteContext.IdentityModel.ID,
				CreatedByIdentityName: siteContext.IdentityModel.Name,
			},
			),
		)
		if err != nil {
			slog.Error("failed to create temp target in database", "error", err)
			return nil, nsutils.ErrInternal
		}
		return ptr(c.tempTargetConverter(dbModel)), nil
	case nightscout.MealBolus:
		break
	case nightscout.TempBasal:
		dbModel, err := siteContext.SiteDB.Queries.CreateTempBasal(ctx,
			nightscout.TreatmentToCreateTempBasal(treatmentModel, nightscout.CreateContext{
				IDPrefix:              et.ToIDPrefix(),
				CreatedByIdentityID:   siteContext.IdentityModel.ID,
				CreatedByIdentityName: siteContext.IdentityModel.Name,
			},
			),
		)
		if err != nil {
			slog.Error("failed to create temp basal in database", "error", err)
			return nil, nsutils.ErrInternal
		}
		return ptr(c.tempBasalConverter(dbModel)), nil
	case nightscout.Note:
		break
	case nightscout.Exercise:
		break
	case nightscout.SiteChange:
		break
	case nightscout.SensorStart:
		break
	case nightscout.SensorChange:
		break
	case nightscout.InsulinChange:
		break
	case nightscout.PumpBatteryChange:
		break
	case nightscout.Question:
		break
	case nightscout.ProfileSwitch:
		break
	case nightscout.CarbCorrection:
		dbModel, err := siteContext.SiteDB.Queries.CreateCarb(ctx,
			nightscout.TreatmentToCreateCarb(treatmentModel, nightscout.CreateContext{
				IDPrefix:              et.ToIDPrefix(),
				CreatedByIdentityID:   siteContext.IdentityModel.ID,
				CreatedByIdentityName: siteContext.IdentityModel.Name,
			},
			),
		)
		if err != nil {
			slog.Error("failed to create carb in database", "error", err)
			return nil, nsutils.ErrInternal
		}
		return ptr(c.carbConverter(dbModel)), nil
	case nightscout.DADAlert:
		break
	case nightscout.OpenAPSOffline:
		break
	case nightscout.BolusWizard:
		break
	}

	slog.Error("unsupported treatment event type", "eventType", *treatmentModel.EventType)
	return nil, nsutils.ErrInternal

}

func (c TreatmentCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	records := []nsutils.Record{}
	carbs, err := siteContext.SiteDB.Queries.AllCarbs(ctx, nsquery.CarbsSearchAll(query))
	if err != nil {
		slog.Error("failed to query carbs from database", "error", err)
		return nil, nsutils.ErrInternal
	}
	records = append(records, CreateRecordSlice(carbs, c.carbConverter)...)

	return records, nil
}

func (c TreatmentCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
