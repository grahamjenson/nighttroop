package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// treatmentEvent is one kind of event a treatment's event group can have, such as its glucose reading. Each kind is a
// type of its own, sharing how its events are converted and kept through eventFrom, mergeEvent, writeEvent and
// hardDeleteEvent.
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
	// TODO: This means events are always updated, we should selectivly update.
	write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error
	// hardDelete deletes the group's event of this kind, if it has one
	hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error
	// addToTreatment sets the treatment's fields from the bundle's event of this kind, if it has one
	addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment)
}

// eventFrom is the treatment's event, converted by partial, when it has one, otherwise nil
func eventFrom[T any](has bool, partial func(nightscoutmodels.Treatment) T, treatment nightscoutmodels.Treatment) *T {
	if !has {
		return nil
	}
	return new(partial(treatment))
}

// mergeEvent sets the event to the update's when it has none, otherwise merges the update's into it
func mergeEvent[T any](event **T, update *T, mergeInto func(source, target *T)) {
	if update == nil {
		return
	}
	if *event == nil {
		*event = update
	} else {
		mergeInto(update, *event)
	}
}

// writeEvent syncs a group's event to the DB, see treatmentEvent.write, setting id, the group's ID for it. remove
// hard deletes it, and create and update write it.
func writeEvent[T interface{ GetID() string }](name string, id **string, event **T, remove func() error, create, update func(T) (*T, error)) error {
	if *event == nil {
		if *id == nil {
			return nil
		}
		if err := remove(); err != nil {
			return err
		}
		*id = nil
		return nil
	}

	verb, write := "create", create
	if *id != nil {
		verb, write = "update", update
	}
	written, err := write(**event)
	if err != nil || written == nil {
		slog.Error("failed to "+verb+" "+name+" in database", "error", err, "dbModel", written)
		return err
	}
	*id, *event = ptr((*written).GetID()), written
	return nil
}

// hardDeleteEvent deletes a group's event, by its ID, if it has one
func hardDeleteEvent(name string, id *string, remove func(id string) (sql.Result, error)) error {
	if id == nil {
		return nil
	}
	res, err := remove(*id)
	if err != nil {
		slog.Error("failed to delete "+name+" by ID", "error", err, "id", *id, "result", res)
	}
	return err
}

// The kinds of event, each with its own type
var (
	glucoseReadings    = glucoseReadingEvents{}
	insulinDoses       = insulinDoseEvents{}
	foodIntakes        = foodIntakeEvents{}
	tempBasals         = tempBasalEvents{}
	tempTargets        = tempTargetEvents{}
	profileSwitches    = profileSwitchEvents{}
	componentEvents    = componentEventEvents{}
	activities         = activityEvents{}
	notificationEvents = notificationEventEvents{}
)

type glucoseReadingEvents struct{}

func (glucoseReadingEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	bundle.GlucoseReading = nil
	if !isAssigned(source.Glucose) {
		return
	}
	target := nightscoutmodels.PartialTreatmentToGlucoseReading(source)
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
	bundle.GlucoseReading = &target
}

func (glucoseReadingEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.GlucoseReading, update.GlucoseReading, nightscoutmodels.UpdateToGlucoseReading)
}

func (k glucoseReadingEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("glucose reading", &bundle.EventGroup.GlucoseReadingID, &bundle.GlucoseReading,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.GlucoseReading) (*sitedb.GlucoseReading, error) {
			return q.CreateGlucoseReading(ctx, nightscoutmodels.GlucoseReadingCreate(event, cc))
		},
		func(event sitedb.GlucoseReading) (*sitedb.GlucoseReading, error) {
			return q.UpdateGlucoseReading(ctx, nightscoutmodels.GlucoseReadingUpdate(event, cc))
		})
}

func (glucoseReadingEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("glucose reading", eventGroup.GlucoseReadingID, func(id string) (sql.Result, error) {
		return q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: id})
	})
}

func (glucoseReadingEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateGlucoseReadingToTreatment(bundle.GlucoseReading, siteContext, treatment)
}

type insulinDoseEvents struct{}

func (insulinDoseEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	bundle.InsulinDose = eventFrom(isAssigned(source.Insulin), nightscoutmodels.PartialTreatmentToInsulinDose, source)
}

func (insulinDoseEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.InsulinDose, update.InsulinDose, nightscoutmodels.UpdateToInsulinDose)
}

func (k insulinDoseEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("insulin dose", &bundle.EventGroup.InsulinDoseID, &bundle.InsulinDose,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.InsulinDose) (*sitedb.InsulinDose, error) {
			return q.CreateInsulinDose(ctx, nightscoutmodels.InsulinDoseCreate(event, cc))
		},
		func(event sitedb.InsulinDose) (*sitedb.InsulinDose, error) {
			return q.UpdateInsulinDose(ctx, nightscoutmodels.InsulinDoseUpdate(event, cc))
		})
}

func (insulinDoseEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("insulin dose", eventGroup.InsulinDoseID, func(id string) (sql.Result, error) {
		return q.HardDeleteInsulinDose(ctx, sitedb.HardDeleteInsulinDoseParams{ID: id})
	})
}

func (insulinDoseEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateInsulinDoseToTreatment(bundle.InsulinDose, treatment)
}

type foodIntakeEvents struct{}

func (foodIntakeEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	has := isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions)
	bundle.FoodIntake = eventFrom(has, nightscoutmodels.PartialTreatmentToFoodIntake, source)
}

func (foodIntakeEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.FoodIntake, update.FoodIntake, nightscoutmodels.UpdateToFoodIntake)
}

func (k foodIntakeEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("food intake", &bundle.EventGroup.FoodIntakeID, &bundle.FoodIntake,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.FoodIntake) (*sitedb.FoodIntake, error) {
			return q.CreateFoodIntake(ctx, nightscoutmodels.FoodIntakeCreate(event, cc))
		},
		func(event sitedb.FoodIntake) (*sitedb.FoodIntake, error) {
			return q.UpdateFoodIntake(ctx, nightscoutmodels.FoodIntakeUpdate(event, cc))
		})
}

func (foodIntakeEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("food intake", eventGroup.FoodIntakeID, func(id string) (sql.Result, error) {
		return q.HardDeleteFoodIntake(ctx, sitedb.HardDeleteFoodIntakeParams{ID: id})
	})
}

func (foodIntakeEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateFoodIntakeToTreatment(bundle.FoodIntake, treatment)
}

type tempBasalEvents struct{}

func (tempBasalEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	has := isAssigned(source.Absolute, source.Rate, source.Percent, source.DurationMinutes)
	bundle.TempBasal = eventFrom(has, nightscoutmodels.PartialTreatmentToTempBasal, source)
}

func (tempBasalEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.TempBasal, update.TempBasal, nightscoutmodels.UpdateToTempBasal)
}

func (k tempBasalEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("temp basal", &bundle.EventGroup.TempBasalID, &bundle.TempBasal,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.TempBasal) (*sitedb.TempBasal, error) {
			return q.CreateTempBasal(ctx, nightscoutmodels.TempBasalCreate(event, cc))
		},
		func(event sitedb.TempBasal) (*sitedb.TempBasal, error) {
			return q.UpdateTempBasal(ctx, nightscoutmodels.TempBasalUpdate(event, cc))
		})
}

func (tempBasalEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("temp basal", eventGroup.TempBasalID, func(id string) (sql.Result, error) {
		return q.HardDeleteTempBasal(ctx, sitedb.HardDeleteTempBasalParams{ID: id})
	})
}

func (tempBasalEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateTempBasalToTreatment(bundle.TempBasal, treatment)
}

type tempTargetEvents struct{}

func (tempTargetEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	bundle.TempTarget = nil
	// Loop sends a correction range for the targets
	if len(source.CorrectionRange) == 2 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[1]
	} else if len(source.CorrectionRange) == 1 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[0]
	}
	if !isAssigned(source.HighTarget, source.LowTarget, source.DurationMinutes) {
		return
	}
	target := nightscoutmodels.PartialTreatmentToTempTarget(source)
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
	bundle.TempTarget = &target
}

func (tempTargetEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.TempTarget, update.TempTarget, nightscoutmodels.UpdateToTempTarget)
}

func (k tempTargetEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("temp target", &bundle.EventGroup.TempTargetID, &bundle.TempTarget,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.TempTarget) (*sitedb.TempTarget, error) {
			return q.CreateTempTarget(ctx, nightscoutmodels.TempTargetCreate(event, cc))
		},
		func(event sitedb.TempTarget) (*sitedb.TempTarget, error) {
			return q.UpdateTempTarget(ctx, nightscoutmodels.TempTargetUpdate(event, cc))
		})
}

func (tempTargetEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("temp target", eventGroup.TempTargetID, func(id string) (sql.Result, error) {
		return q.HardDeleteTempTarget(ctx, sitedb.HardDeleteTempTargetParams{ID: id})
	})
}

func (tempTargetEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateTempTargetToTreatment(bundle.TempTarget, treatment)
}

type profileSwitchEvents struct{}

func (profileSwitchEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	bundle.ProfileSwitch = nil
	if !isAssigned(source.Name, source.Profile, source.ProfileJSON) {
		return
	}
	target := nightscoutmodels.PartialTreatmentToProfileSwitch(source)
	// TODO: This is the bit that will need more work, with the json of profile
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
	bundle.ProfileSwitch = &target
}

func (profileSwitchEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.ProfileSwitch, update.ProfileSwitch, nightscoutmodels.UpdateToProfileSwitch)
}

func (k profileSwitchEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("profile switch", &bundle.EventGroup.ProfileSwitchID, &bundle.ProfileSwitch,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.ProfileSwitch) (*sitedb.ProfileSwitch, error) {
			return q.CreateProfileSwitch(ctx, nightscoutmodels.ProfileSwitchCreate(event, cc))
		},
		func(event sitedb.ProfileSwitch) (*sitedb.ProfileSwitch, error) {
			return q.UpdateProfileSwitch(ctx, nightscoutmodels.ProfileSwitchUpdate(event, cc))
		})
}

func (profileSwitchEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("profile switch", eventGroup.ProfileSwitchID, func(id string) (sql.Result, error) {
		return q.HardDeleteProfileSwitch(ctx, sitedb.HardDeleteProfileSwitchParams{ID: id})
	})
}

func (profileSwitchEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateProfileSwitchToTreatment(bundle.ProfileSwitch, siteContext, treatment)
}

type componentEventEvents struct{}

func (componentEventEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	has := isAssigned(source.PumpID, source.PumpSerial, source.PumpType)
	bundle.ComponentEvent = eventFrom(has, nightscoutmodels.PartialTreatmentToComponentEvent, source)
}

func (componentEventEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.ComponentEvent, update.ComponentEvent, nightscoutmodels.UpdateToComponentEvent)
}

func (k componentEventEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("component event", &bundle.EventGroup.ComponentEventID, &bundle.ComponentEvent,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.ComponentEvent) (*sitedb.ComponentEvent, error) {
			return q.CreateComponentEvent(ctx, nightscoutmodels.ComponentEventCreate(event, cc))
		},
		func(event sitedb.ComponentEvent) (*sitedb.ComponentEvent, error) {
			return q.UpdateComponentEvent(ctx, nightscoutmodels.ComponentEventUpdate(event, cc))
		})
}

func (componentEventEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("component event", eventGroup.ComponentEventID, func(id string) (sql.Result, error) {
		return q.HardDeleteComponentEvent(ctx, sitedb.HardDeleteComponentEventParams{ID: id})
	})
}

func (componentEventEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateComponentEventToTreatment(bundle.ComponentEvent, treatment)
}

type activityEvents struct{}

func (activityEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	has := isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate)
	bundle.Activity = eventFrom(has, nightscoutmodels.PartialTreatmentToActivity, source)
}

func (activityEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.Activity, update.Activity, nightscoutmodels.UpdateToActivity)
}

func (k activityEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("activity", &bundle.EventGroup.ActivityID, &bundle.Activity,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.Activity) (*sitedb.Activity, error) {
			return q.CreateActivity(ctx, nightscoutmodels.ActivityCreate(event, cc))
		},
		func(event sitedb.Activity) (*sitedb.Activity, error) {
			return q.UpdateActivity(ctx, nightscoutmodels.ActivityUpdate(event, cc))
		})
}

func (activityEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("activity", eventGroup.ActivityID, func(id string) (sql.Result, error) {
		return q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: id})
	})
}

func (activityEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateActivityToTreatment(bundle.Activity, treatment)
}

type notificationEventEvents struct{}

// fromTreatment always sets a notification, as every treatment kept with one has one
func (notificationEventEvents) fromTreatment(source nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
	target := nightscoutmodels.PartialTreatmentToNotificationEvent(source)
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
	bundle.NotificationEvent = &target
}

func (notificationEventEvents) merge(bundle *sitedb.EventGroupsBundle, update sitedb.EventGroupsBundle) {
	mergeEvent(&bundle.NotificationEvent, update.NotificationEvent, nightscoutmodels.UpdateToNotificationEvent)
}

func (k notificationEventEvents) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	return writeEvent("notification event", &bundle.EventGroup.NotificationEventID, &bundle.NotificationEvent,
		func() error { return k.hardDelete(ctx, q, bundle.EventGroup) },
		func(event sitedb.NotificationEvent) (*sitedb.NotificationEvent, error) {
			return q.CreateNotificationEvent(ctx, nightscoutmodels.NotificationEventCreate(event, cc))
		},
		func(event sitedb.NotificationEvent) (*sitedb.NotificationEvent, error) {
			return q.UpdateNotificationEvent(ctx, nightscoutmodels.NotificationEventUpdate(event, cc))
		})
}

func (notificationEventEvents) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
	return hardDeleteEvent("notification event", eventGroup.NotificationEventID, func(id string) (sql.Result, error) {
		return q.HardDeleteNotificationEvent(ctx, sitedb.HardDeleteNotificationEventParams{ID: id})
	})
}

func (notificationEventEvents) addToTreatment(bundle *sitedb.EventGroupsBundle, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	nightscoutmodels.UpdateNotificationEventToTreatment(bundle.NotificationEvent, treatment)
}

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
	nightscoutmodels.CannulaFill:       {componentEvents},
	nightscoutmodels.Prime:             {componentEvents},
	nightscoutmodels.ReservoirChange:   {componentEvents},
	nightscoutmodels.Calibration:       {glucoseReadings, componentEvents},
	nightscoutmodels.Bolus:             {insulinDoses, componentEvents},
}
