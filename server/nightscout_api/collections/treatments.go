package collections

import (
	"context"
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
		nightscout.CarbCorrection:    treatments.CarbsCorrection{},
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
