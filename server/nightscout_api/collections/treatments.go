package collections

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/collections/treatments"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentTypeCollection interface {
	Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError)
	Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError)
}

type TreatmentCollection struct {
	converter                func(c *site.SearchEventGroupsRow) nsutils.Record
	treatmentCollectionTypes map[nightscout.TreatmentEventType]TreatmentTypeCollection
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

func NewTreatmentCollection() TreatmentCollection {
	typeMap := map[nightscout.TreatmentEventType]TreatmentTypeCollection{
		nightscout.Announcement:      treatments.Announcement{},
		nightscout.BGCheck:           treatments.BGCheck{},
		nightscout.BolusWizard:       treatments.BolusWizard{},
		nightscout.CarbCorrection:    treatments.NewCarbsCorrection(),
		nightscout.ComboBolus:        treatments.ComboBolus{},
		nightscout.CorrectionBolus:   treatments.CorrectionBolus{},
		nightscout.DADAlert:          treatments.DADAlert{},
		nightscout.Exercise:          treatments.Exercise{},
		nightscout.InsulinChange:     treatments.InsulinChange{},
		nightscout.MealBolus:         treatments.MealBolus{},
		nightscout.Note:              treatments.Note{},
		nightscout.OpenAPSOffline:    treatments.OpenAPSOffline{},
		nightscout.ProfileSwitch:     treatments.ProfileSwitch{},
		nightscout.PumpBatteryChange: treatments.PumpBatteryChange{},
		nightscout.Question:          treatments.Question{},
		nightscout.SensorChange:      treatments.SensorChange{},
		nightscout.SensorStart:       treatments.SensorStart{},
		nightscout.SiteChange:        treatments.SiteChange{},
		nightscout.SnackBolus:        treatments.SnackBolus{},
		nightscout.TempBasal:         treatments.NewTempBasal(),
		nightscout.TemporaryTarget:   treatments.NewTempTarget(),
		nightscout.Unknown:           treatments.Unknown{},
	}

	return TreatmentCollection{
		converter:                nsutils.DBModelToRecord(nsutils.Treatments, nightscout.DBModelToTreatment),
		treatmentCollectionTypes: typeMap,
	}
}

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

	typeCollect, ok := c.treatmentCollectionTypes[et]
	if !ok {
		slog.Error("unsupported treatment event type", "eventType", et)
		return nil, nsutils.ErrInternal
	}
	record, rerr := typeCollect.Create(ctx, siteContext, treatmentModel)
	if rerr != nil {
		slog.Error("failed to create treatment in database", "error", rerr)
		return nil, rerr
	}
	if record == nil {
		slog.Error("treatment creation returned nil value", "eventType", et)
		return nil, nsutils.ErrInternal
	}

	return record, nil
}

func (c TreatmentCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.GetEventGroup(ctx, site.GetEventGroupParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows || dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get device status", "error", err)
		return nil, nsutils.ErrInternal
	}

	seg := site.GetEventGroupToSearchEventGroup(*dbModel)
	record := c.converter(&seg)
	return &record, nil
}

func (c TreatmentCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	if query.TypeIn == nil {
		// TODO list all supportted types so we can limit to those that NS Treatment expects
	}

	treatments, rerr := siteContext.SiteDB.Queries.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if rerr != nil {
		slog.Error("failed to query event groups from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(treatments, c.converter), nil
}

func (c TreatmentCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	for _, typeCollect := range c.treatmentCollectionTypes {
		if strings.HasPrefix(identifier, typeCollect.GetPrefix()) {
			record, rerr := typeCollect.Update(ctx, siteContext, identifier, patch)
			if rerr != nil {
				slog.Error("failed to update treatment in database", "error", rerr)
				return nil, rerr
			}
			if record == nil {
				slog.Error("treatment update returned nil value", "identifier", identifier)
				return nil, nsutils.ErrInternal
			}
			return record, nil
		}
	}

	slog.Error("unsupported treatment event type for identifier", "identifier", identifier)
	return nil, nsutils.ErrInternal
}

func (c TreatmentCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	for _, typeCollect := range c.treatmentCollectionTypes {
		if strings.HasPrefix(identifier, typeCollect.GetPrefix()) {
			rerr := typeCollect.Delete(ctx, siteContext, identifier, permanent)
			if rerr != nil {
				slog.Error("failed to delete treatment from database", "error", rerr)
				return rerr
			}
			return nil
		}
	}

	slog.Error("unsupported treatment event type for identifier", "identifier", identifier)
	return nsutils.ErrInternal
}

func (c TreatmentCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	nightscoutType := nightscout.TreatmentEventType(typeString)
	typeCollect, ok := c.treatmentCollectionTypes[nightscoutType]
	if !ok {
		slog.Error("unsupported treatment event type for deletion", "typeString", typeString)
		return 0, nsutils.ErrInternal
	}
	count, rerr := typeCollect.DeleteAll(ctx, siteContext, permanent)
	if rerr != nil {
		slog.Error("failed to delete treatments of type from database", "error", rerr)
		return 0, rerr
	}
	return count, nil
}
