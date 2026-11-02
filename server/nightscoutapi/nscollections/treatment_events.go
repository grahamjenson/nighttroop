package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

// treatmentEvent is one kind of event a treatment's event group can have, such as its glucose reading
type treatmentEvent interface {
	// fromTreatment sets the bundle's event of this kind from the treatment, nil when the treatment has none
	fromTreatment(treatment nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle)
	// merge sets the bundle's event of this kind to the update's when it has none, otherwise merges the update's into it
	merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle)
	// write syncs the bundle's event of this kind to the DB, setting the group's ID for it:
	// 1. Delete if the group has an ID for it but no event
	// 2. Create if the group has an event but no ID for it
	// 3. Update if the group has an ID and an event
	//
	// So an event is written again on every update of its group, even when it has not changed.
	write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error
	// hardDelete deletes the group's event of this kind, if it has one
	hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error
	// addToTreatment sets the treatment's fields from the bundle's event of this kind, if it has one
	addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment)
}

// eventKind is a treatmentEvent of events T, kept with create params C, update params U and hard delete params D
type eventKind[T interface{ GetID() string }, C, U any, D ~struct{ ID string }] struct {
	name string
	// of is the bundle's ID for its event of this kind, and the event
	of func(bundle *sitedb.EventGroupsBundle) (**string, **T)
	// prepare, when set, fills in the fields of the treatment this kind reads from its other fields
	prepare func(treatment nightscoutmodels.Treatment) nightscoutmodels.Treatment
	// has is whether the treatment has an event of this kind, always when it is nil
	has func(treatment nightscoutmodels.Treatment) bool
	// partial converts the treatment to its event of this kind, and complete, when set, sets what partial cannot
	partial  func(treatment nightscoutmodels.Treatment) T
	complete func(treatment nightscoutmodels.Treatment, event *T)
	// toTreatment sets the treatment's fields from the event, which may be nil
	toTreatment func(event *T, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment)
	mergeInto   func(source, target *T)
	toCreate    func(event T, cc nightscoutmodels.ConverterContext) C
	create      func(q *sitedb.Queries, ctx context.Context, params C) (*T, error)
	toUpdate    func(event T, cc nightscoutmodels.ConverterContext) U
	update      func(q *sitedb.Queries, ctx context.Context, params U) (*T, error)
	delete      func(q *sitedb.Queries, ctx context.Context, params D) (sql.Result, error)
}

func (k eventKind[T, C, U, D]) fromTreatment(treatment nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	_, event := k.of(bundle)
	*event = nil
	if k.prepare != nil {
		treatment = k.prepare(treatment)
	}
	if k.has != nil && !k.has(treatment) {
		return
	}
	converted := k.partial(treatment)
	if k.complete != nil {
		k.complete(treatment, &converted)
	}
	*event = &converted
}

func (k eventKind[T, C, U, D]) addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	_, event := k.of(bundle)
	k.toTreatment(*event, siteContext, treatment)
}

// withoutSite is a kind's toTreatment for a converter that does not need the site
func withoutSite[T any](convert func(event *T, treatment *nightscoutmodels.Treatment)) func(*T, *httputils.SiteContext, *nightscoutmodels.Treatment) {
	return func(event *T, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
		convert(event, treatment)
	}
}

func (k eventKind[T, C, U, D]) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	_, event := k.of(bundle)
	_, updated := k.of(&update)
	if *updated == nil {
		return
	}
	if *event == nil {
		*event = *updated
	} else {
		k.mergeInto(*updated, *event)
	}
}

func (k eventKind[T, C, U, D]) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	id, event := k.of(bundle)
	if *event == nil {
		if *id == nil {
			return nil
		}
		if _, err := k.delete(q, ctx, D{ID: **id}); err != nil {
			slog.ErrorContext(ctx, "failed to delete "+k.name+" from database", "error", err)
			return err
		}
		*id = nil
		return nil
	}

	verb, written, err := "create", (*T)(nil), error(nil)
	if *id == nil {
		written, err = k.create(q, ctx, k.toCreate(**event, cc))
	} else {
		verb = "update"
		written, err = k.update(q, ctx, k.toUpdate(**event, cc))
	}
	if err != nil || written == nil {
		slog.ErrorContext(ctx, "failed to "+verb+" "+k.name+" in database", "error", err, "dbModel", written)
		return err
	}
	*id, *event = ptr((*written).GetID()), written
	return nil
}

func (k eventKind[T, C, U, D]) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	id, _ := k.of(&sitedb.EventGroupsBundle{EventGroup: eventGroup})
	if *id == nil {
		return nil
	}
	res, err := k.delete(q, ctx, D{ID: **id})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete "+k.name+" by ID", "error", err, "id", **id, "result", res)
	}
	return err
}

var (
	glucoseReadings = eventKind[sitedb.GlucoseReading, sitedb.CreateGlucoseReadingParams, sitedb.UpdateGlucoseReadingParams, sitedb.HardDeleteGlucoseReadingParams]{
		name: "glucose reading",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.GlucoseReading) {
			return &b.EventGroup.GlucoseReadingID, &b.GlucoseReading
		},
		has:     func(source nightscoutmodels.Treatment) bool { return isAssigned(source.Glucose) },
		partial: nightscoutmodels.PartialTreatmentToGlucoseReading,
		complete: func(source nightscoutmodels.Treatment, target *sitedb.GlucoseReading) {
			target.ReadingType = "sgv"
			target.Units = source.Units

			if source.Glucose != nil {
				target.Mgdl = target.Units.ToMGDL(float64(*source.Glucose))
			}

			if source.GlucoseType != nil {
				switch *source.GlucoseType {
				case nightscoutmodels.Sensor:
					target.ReadingType = "sgv"
				case nightscoutmodels.Finger:
					target.ReadingType = "mbg"
				case nightscoutmodels.Meter:
					target.ReadingType = "mbg"
				}
			}
		},
		toTreatment: nightscoutmodels.UpdateGlucoseReadingToTreatment, mergeInto: nightscoutmodels.UpdateToGlucoseReading,
		toCreate: nightscoutmodels.GlucoseReadingCreate, create: (*sitedb.Queries).CreateGlucoseReading,
		toUpdate: nightscoutmodels.GlucoseReadingUpdate, update: (*sitedb.Queries).UpdateGlucoseReading,
		delete: (*sitedb.Queries).HardDeleteGlucoseReading,
	}
	insulinDoses = eventKind[sitedb.InsulinDose, sitedb.CreateInsulinDoseParams, sitedb.UpdateInsulinDoseParams, sitedb.HardDeleteInsulinDoseParams]{
		name: "insulin dose",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.InsulinDose) {
			return &b.EventGroup.InsulinDoseID, &b.InsulinDose
		},
		has:         func(source nightscoutmodels.Treatment) bool { return isAssigned(source.Insulin) },
		partial:     nightscoutmodels.PartialTreatmentToInsulinDose,
		toTreatment: withoutSite(nightscoutmodels.UpdateInsulinDoseToTreatment), mergeInto: nightscoutmodels.UpdateToInsulinDose,
		toCreate: nightscoutmodels.InsulinDoseCreate, create: (*sitedb.Queries).CreateInsulinDose,
		toUpdate: nightscoutmodels.InsulinDoseUpdate, update: (*sitedb.Queries).UpdateInsulinDose,
		delete: (*sitedb.Queries).HardDeleteInsulinDose,
	}
	foodIntakes = eventKind[sitedb.FoodIntake, sitedb.CreateFoodIntakeParams, sitedb.UpdateFoodIntakeParams, sitedb.HardDeleteFoodIntakeParams]{
		name: "food intake",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.FoodIntake) {
			return &b.EventGroup.FoodIntakeID, &b.FoodIntake
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions)
		},
		partial:     nightscoutmodels.PartialTreatmentToFoodIntake,
		toTreatment: withoutSite(nightscoutmodels.UpdateFoodIntakeToTreatment), mergeInto: nightscoutmodels.UpdateToFoodIntake,
		toCreate: nightscoutmodels.FoodIntakeCreate, create: (*sitedb.Queries).CreateFoodIntake,
		toUpdate: nightscoutmodels.FoodIntakeUpdate, update: (*sitedb.Queries).UpdateFoodIntake,
		delete: (*sitedb.Queries).HardDeleteFoodIntake,
	}
	tempBasals = eventKind[sitedb.TempBasal, sitedb.CreateTempBasalParams, sitedb.UpdateTempBasalParams, sitedb.HardDeleteTempBasalParams]{
		name: "temp basal",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.TempBasal) {
			return &b.EventGroup.TempBasalID, &b.TempBasal
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.Absolute, source.Rate, source.TempBasalPercent, source.DurationMinutes)
		},
		partial: nightscoutmodels.PartialTreatmentToTempBasal,
		// AAPS sends its rate as rate, and Loop as absolute too
		complete: func(source nightscoutmodels.Treatment, target *sitedb.TempBasal) {
			if rate := nightscoutmodels.TempBasalValuePtr(source); rate != nil {
				target.Absolute = *rate
			}
		},
		toTreatment: withoutSite(nightscoutmodels.UpdateTempBasalToTreatment), mergeInto: nightscoutmodels.UpdateToTempBasal,
		toCreate: nightscoutmodels.TempBasalCreate, create: (*sitedb.Queries).CreateTempBasal,
		toUpdate: nightscoutmodels.TempBasalUpdate, update: (*sitedb.Queries).UpdateTempBasal,
		delete: (*sitedb.Queries).HardDeleteTempBasal,
	}
	tempTargets = eventKind[sitedb.TempTarget, sitedb.CreateTempTargetParams, sitedb.UpdateTempTargetParams, sitedb.HardDeleteTempTargetParams]{
		name: "temp target",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.TempTarget) {
			return &b.EventGroup.TempTargetID, &b.TempTarget
		},
		// Loop sends a correction range for the targets
		prepare: func(source nightscoutmodels.Treatment) nightscoutmodels.Treatment {
			if len(source.CorrectionRange) == 2 {
				source.LowTarget = source.CorrectionRange[0]
				source.HighTarget = source.CorrectionRange[1]
			} else if len(source.CorrectionRange) == 1 {
				source.LowTarget = source.CorrectionRange[0]
				source.HighTarget = source.CorrectionRange[0]
			}
			// An indefinite one lasts as long as a temp target can, a week
			if source.DurationMinutes == nil && source.DurationType != nil && *source.DurationType == "indefinite" {
				source.DurationMinutes = new(types.FlexibleInt64(7 * 24 * 60))
			}
			return source
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.HighTarget, source.LowTarget, source.DurationMinutes)
		},
		partial: nightscoutmodels.PartialTreatmentToTempTarget,
		complete: func(source nightscoutmodels.Treatment, target *sitedb.TempTarget) {
			if source.HighTarget != nil {
				target.HighTargetMgdl = source.HighTarget.Float64()
			} else if source.LowTarget != nil {
				// If Target Top is missing then we assume both targets are the same
				target.HighTargetMgdl = source.LowTarget.Float64()
			}

			if source.LowTarget != nil {
				target.LowTargetMgdl = source.LowTarget.Float64()
			} else if source.HighTarget != nil {
				// If Target Bottom is missing then we assume both targets are the same
				target.LowTargetMgdl = source.HighTarget.Float64()
			}
		},
		toTreatment: withoutSite(nightscoutmodels.UpdateTempTargetToTreatment), mergeInto: nightscoutmodels.UpdateToTempTarget,
		toCreate: nightscoutmodels.TempTargetCreate, create: (*sitedb.Queries).CreateTempTarget,
		toUpdate: nightscoutmodels.TempTargetUpdate, update: (*sitedb.Queries).UpdateTempTarget,
		delete: (*sitedb.Queries).HardDeleteTempTarget,
	}
	profileSwitches = eventKind[sitedb.ProfileSwitch, sitedb.CreateProfileSwitchParams, sitedb.UpdateProfileSwitchParams, sitedb.HardDeleteProfileSwitchParams]{
		name: "profile switch",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.ProfileSwitch) {
			return &b.EventGroup.ProfileSwitchID, &b.ProfileSwitch
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.Name, source.Profile, source.ProfileJSON)
		},
		partial: nightscoutmodels.PartialTreatmentToProfileSwitch,
		complete: func(source nightscoutmodels.Treatment, target *sitedb.ProfileSwitch) {
			// The profile's settings and schedules are kept from its JSON, converted to mg/dL, but not checked
			if source.Profile != nil {
				target.Name = source.Profile.String()
			}

			if source.ProfileJSON != nil {
				var profile nightscoutmodels.Profile
				err := json.Unmarshal([]byte(*source.ProfileJSON), &profile)
				if err == nil {

					profile.ConvertToMGDL()

					target.Dia = profile.Dia.Float64p()
					target.Timezone = profile.Timezone.Stringp()

					// Set units
					target.Units = profile.Units
					target.CarbsHr = profile.CarbsHr.Float64p()
					target.Delay = profile.Delay.Float64p()

					basalBytes, _ := json.Marshal(profile.Basal)
					carbRatioBytes, _ := json.Marshal(profile.CarbRatio)
					sensBytes, _ := json.Marshal(profile.Sens)
					targetLowBytes, _ := json.Marshal(profile.TargetLow)
					targetHighBytes, _ := json.Marshal(profile.TargetHigh)

					if basalBytes != nil {
						target.BasalRatesJson = ptr(string(basalBytes))
					} else {
						target.BasalRatesJson = ptr("[]")
					}

					if carbRatioBytes != nil {
						target.CarbRatiosJson = ptr(string(carbRatioBytes))
					} else {
						target.CarbRatiosJson = ptr("[]")
					}

					if sensBytes != nil {
						target.InsulinSensitivitiesJson = ptr(string(sensBytes))
					} else {
						target.InsulinSensitivitiesJson = ptr("[]")
					}

					if targetLowBytes != nil {
						target.TargetsLowJson = ptr(string(targetLowBytes))
					} else {
						target.TargetsLowJson = ptr("[]")
					}

					if targetHighBytes != nil {
						target.TargetsHighJson = ptr(string(targetHighBytes))
					} else {
						target.TargetsHighJson = ptr("[]")
					}
				}
			}
		},
		toTreatment: nightscoutmodels.UpdateProfileSwitchToTreatment, mergeInto: nightscoutmodels.UpdateToProfileSwitch,
		toCreate: nightscoutmodels.ProfileSwitchCreate, create: (*sitedb.Queries).CreateProfileSwitch,
		toUpdate: nightscoutmodels.ProfileSwitchUpdate, update: (*sitedb.Queries).UpdateProfileSwitch,
		delete: (*sitedb.Queries).HardDeleteProfileSwitch,
	}
	componentEvents = eventKind[sitedb.ComponentEvent, sitedb.CreateComponentEventParams, sitedb.UpdateComponentEventParams, sitedb.HardDeleteComponentEventParams]{
		name: "component event",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.ComponentEvent) {
			return &b.EventGroup.ComponentEventID, &b.ComponentEvent
		},
		// A change of a sensor, site or pump part always has one, and another treatment has one when it names the pump
		has: func(source nightscoutmodels.Treatment) bool {
			return slices.Contains(sitemodels.ComponentEventTypes, source.EventType.String()) ||
				isAssigned(source.PumpID, source.PumpSerial, source.PumpType)
		},
		partial:     nightscoutmodels.PartialTreatmentToComponentEvent,
		toTreatment: withoutSite(nightscoutmodels.UpdateComponentEventToTreatment), mergeInto: nightscoutmodels.UpdateToComponentEvent,
		toCreate: nightscoutmodels.ComponentEventCreate, create: (*sitedb.Queries).CreateComponentEvent,
		toUpdate: nightscoutmodels.ComponentEventUpdate, update: (*sitedb.Queries).UpdateComponentEvent,
		delete: (*sitedb.Queries).HardDeleteComponentEvent,
	}
	activities = eventKind[sitedb.Activity, sitedb.CreateActivityParams, sitedb.UpdateActivityParams, sitedb.HardDeleteActivityParams]{
		name: "activity",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.Activity) {
			return &b.EventGroup.ActivityID, &b.Activity
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate)
		},
		partial:     nightscoutmodels.PartialTreatmentToActivity,
		toTreatment: withoutSite(nightscoutmodels.UpdateActivityToTreatment), mergeInto: nightscoutmodels.UpdateToActivity,
		toCreate: nightscoutmodels.ActivityCreate, create: (*sitedb.Queries).CreateActivity,
		toUpdate: nightscoutmodels.ActivityUpdate, update: (*sitedb.Queries).UpdateActivity,
		delete: (*sitedb.Queries).HardDeleteActivity,
	}
	notificationEvents = eventKind[sitedb.NotificationEvent, sitedb.CreateNotificationEventParams, sitedb.UpdateNotificationEventParams, sitedb.HardDeleteNotificationEventParams]{
		name: "notification event",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.NotificationEvent) {
			return &b.EventGroup.NotificationEventID, &b.NotificationEvent
		},
		// Every treatment kept with a notification has one
		partial: nightscoutmodels.PartialTreatmentToNotificationEvent,
		complete: func(source nightscoutmodels.Treatment, target *sitedb.NotificationEvent) {
			if eventType := *source.EventType; eventType == nightscoutmodels.Announcement {
				target.Severity = 0
				target.NotificationType = "announcement"
				target.Title = "Announcement"
			} else if eventType == nightscoutmodels.Question {
				target.Severity = 1
				target.NotificationType = "question"
				target.Title = "Question"
			} else {
				target.Severity = 0
				target.NotificationType = "announcement"
				target.Title = "Treatment Event"
			}

			if source.Name != nil {
				target.Title = source.Name.String()
			}
		},
		toTreatment: withoutSite(nightscoutmodels.UpdateNotificationEventToTreatment), mergeInto: nightscoutmodels.UpdateToNotificationEvent,
		toCreate: nightscoutmodels.NotificationEventCreate, create: (*sitedb.Queries).CreateNotificationEvent,
		toUpdate: nightscoutmodels.NotificationEventUpdate, update: (*sitedb.Queries).UpdateNotificationEvent,
		delete: (*sitedb.Queries).HardDeleteNotificationEvent,
	}
)

// treatmentEvents are every kind of event a group can have, in the order they are written
var treatmentEvents = []treatmentEvent{
	glucoseReadings, insulinDoses, foodIntakes, tempBasals, tempTargets, profileSwitches, componentEvents, activities,
	notificationEvents,
}

// treatmentTypeEvents are the kinds of event each type of treatment keeps; its other fields are dropped
var treatmentTypeEvents = map[nightscoutmodels.TreatmentEventType][]treatmentEvent{
	nightscoutmodels.Announcement:      {glucoseReadings, notificationEvents},
	nightscoutmodels.BGCheck:           {glucoseReadings},
	nightscoutmodels.BolusWizard:       {glucoseReadings, insulinDoses, foodIntakes, componentEvents},
	nightscoutmodels.CarbCorrection:    {foodIntakes},
	nightscoutmodels.ComboBolus:        {glucoseReadings, insulinDoses, foodIntakes, componentEvents},
	nightscoutmodels.CorrectionBolus:   {insulinDoses, glucoseReadings, componentEvents},
	nightscoutmodels.DADAlert:          {},
	nightscoutmodels.Exercise:          {activities},
	nightscoutmodels.InsulinChange:     {componentEvents},
	nightscoutmodels.MealBolus:         {glucoseReadings, insulinDoses, foodIntakes, componentEvents},
	nightscoutmodels.Note:              {glucoseReadings, insulinDoses, foodIntakes, profileSwitches, activities, componentEvents},
	nightscoutmodels.OpenAPSOffline:    {componentEvents},
	nightscoutmodels.ProfileSwitch:     {profileSwitches},
	nightscoutmodels.PumpBatteryChange: {componentEvents},
	nightscoutmodels.Question:          {notificationEvents},
	nightscoutmodels.SensorChange:      {componentEvents},
	nightscoutmodels.SensorStart:       {componentEvents},
	nightscoutmodels.SiteChange:        {componentEvents},
	nightscoutmodels.SnackBolus:        {glucoseReadings, insulinDoses, foodIntakes, componentEvents},
	nightscoutmodels.TempBasal:         {tempBasals, componentEvents},
	nightscoutmodels.TemporaryTarget:   {tempTargets},
	nightscoutmodels.TemporaryOverride: {tempTargets},
	nightscoutmodels.CannulaFill:       {componentEvents},
	nightscoutmodels.Prime:             {componentEvents},
	nightscoutmodels.ReservoirChange:   {componentEvents},
	nightscoutmodels.Calibration:       {glucoseReadings, componentEvents},
	nightscoutmodels.Bolus:             {insulinDoses, componentEvents},
	nightscoutmodels.SMB:               {insulinDoses},
	nightscoutmodels.ExternalInsulin:   {insulinDoses},
}

// percentSwitchMs is how long before a temp basal a profile switch with a duration is looked for, as one still in force,
// and percentSwitches the most switches looked at
const percentSwitchMs, percentSwitches = 7 * 24 * 60 * 60_000, 1000

// percentRate sets the rate of a temp basal the treatment sends only as a percent from the profile's basal at its
// start, see sitemodels.PercentBasal, keeping its rate when no profile has a basal then
func percentRate(ctx context.Context, q *sitedb.Queries, treatment nightscoutmodels.Treatment, tempBasal *sitedb.TempBasal) error {
	if tempBasal == nil || tempBasal.Percent == nil || nightscoutmodels.TempBasalValuePtr(treatment) != nil {
		return nil
	}
	at := tempBasal.OccurredAt
	defaults, err := q.SearchSummaryDefaultProfiles(ctx, sitedb.SearchSummaryDefaultProfilesParams{Since: at, Until: at, Limit: 1})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find the profile for a temp basal's percent", "error", err)
		return err
	}
	switches, err := q.SearchSummaryProfileSwitches(ctx, sitedb.SearchSummaryProfileSwitchesParams{Since: at, Until: at, Earliest: at - percentSwitchMs, Limit: percentSwitches})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find the profile switches for a temp basal's percent", "error", err)
		return err
	}
	location := time.UTC
	if siteContext := httputils.GetSiteContext(ctx); siteContext != nil {
		location = siteContext.GetSiteTimezone()
	}
	slices.Reverse(switches)
	if rate := sitemodels.PercentBasal(at, *tempBasal.Percent, location, defaults, switches); rate != nil {
		tempBasal.Absolute = *rate
	}
	return nil
}
