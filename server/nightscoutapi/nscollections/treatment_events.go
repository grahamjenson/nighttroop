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
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

// treatmentEvent is one kind of event a treatment's event group can have, such as its glucose reading
type treatmentEvent interface {
	// fromTreatment sets the bundle's event of this kind from the treatment, nil when the treatment has none
	fromTreatment(treatment nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle)
	// keep gives the bundle's event of this kind, when the stored group has one too, its identifier and what of it a
	// treatment does not have, so it is written over the stored one
	keep(bundle, stored *sitedb.EventGroupsBundle)
	// write writes the bundle's event of this kind to the DB, setting the group's ID for it: it creates one the group
	// has no ID for, and writes one it has an ID for over the stored one. So an event is written again on every update
	// of its group, even when it has not changed.
	write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error
	// drop deletes the stored group's event of this kind the bundle no longer has, once its group no longer refers to it
	drop(ctx context.Context, q *sitedb.Queries, bundle, stored *sitedb.EventGroupsBundle) error
	// hardDelete deletes the group's event of this kind, if it has one
	hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error
	// addToTreatment sets the treatment's fields from the bundle's event of this kind, if it has one
	addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment)
}

// eventKind is a treatmentEvent of events T, kept with create params C and hard delete params D
type eventKind[T interface{ GetID() string }, C any, D ~struct{ ID string }] struct {
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
	// kept, when set, gives the event what of the stored one a treatment does not have
	kept     func(stored, event *T)
	toCreate func(event T, cc nightscoutmodels.ConverterContext) C
	create   func(q *sitedb.Queries, ctx context.Context, params C) (*T, error)
	// rewrite writes the event over the stored one with the identifier
	rewrite func(q *sitedb.Queries, ctx context.Context, id string, params C) (*T, error)
	delete  func(q *sitedb.Queries, ctx context.Context, params D) (sql.Result, error)
}

func (k eventKind[T, C, D]) fromTreatment(treatment nightscoutmodels.Treatment, bundle *sitedb.EventGroupsBundle) {
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

func (k eventKind[T, C, D]) addToTreatment(bundle *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
	_, event := k.of(bundle)
	k.toTreatment(*event, siteContext, treatment)
}

// withoutSite is a kind's toTreatment for a converter that does not need the site
func withoutSite[T any](convert func(event *T, treatment *nightscoutmodels.Treatment)) func(*T, *httputils.SiteContext, *nightscoutmodels.Treatment) {
	return func(event *T, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
		convert(event, treatment)
	}
}

func (k eventKind[T, C, D]) keep(bundle, stored *sitedb.EventGroupsBundle) {
	id, event := k.of(bundle)
	storedID, storedEvent := k.of(stored)
	if *event == nil || *storedID == nil {
		return
	}
	*id = *storedID
	if k.kept != nil && *storedEvent != nil {
		k.kept(*storedEvent, *event)
	}
}

func (k eventKind[T, C, D]) write(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	id, event := k.of(bundle)
	if *event == nil {
		return nil
	}
	verb, written, err := "create", (*T)(nil), error(nil)
	if *id == nil {
		written, err = k.create(q, ctx, k.toCreate(**event, cc))
	} else {
		verb = "rewrite"
		written, err = k.rewrite(q, ctx, **id, k.toCreate(**event, cc))
	}
	if err != nil || written == nil {
		slog.ErrorContext(ctx, "failed to "+verb+" "+k.name+" in database", "error", err, "dbModel", written)
		return err
	}
	*id, *event = ptr((*written).GetID()), written
	return nil
}

func (k eventKind[T, C, D]) drop(ctx context.Context, q *sitedb.Queries, bundle, stored *sitedb.EventGroupsBundle) error {
	id, _ := k.of(bundle)
	storedID, _ := k.of(stored)
	if *storedID == nil || *id != nil {
		return nil
	}
	if _, err := k.delete(q, ctx, D{ID: **storedID}); err != nil {
		slog.ErrorContext(ctx, "failed to delete "+k.name+" from database", "error", err)
		return err
	}
	return nil
}

func (k eventKind[T, C, D]) hardDelete(ctx context.Context, q *sitedb.Queries, eventGroup sitedb.EventGroup) error {
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

// AAPS sends a profile switch's timeshift in ms, and reads it back so, with its originalProfileName, as Nightscout's
// profile functions read it; one from the careportal or Trio has no originalProfileName and is kept in minutes
func aapsTimeshift[N any, T interface{ ~int64 }](originalProfileName *N, timeshift *T) bool {
	return originalProfileName != nil && timeshift != nil
}

const msPerMinute = 60_000

// Loop's override with no end is indefinite, kept as long as a temp target can be, a week
const (
	indefinite        = "indefinite"
	indefiniteMinutes = 7 * 24 * 60
)

var (
	glucoseReadings = eventKind[sitedb.GlucoseReading, sitedb.CreateGlucoseReadingParams, sitedb.HardDeleteGlucoseReadingParams]{
		name: "glucose reading",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.GlucoseReading) {
			return &b.EventGroup.GlucoseReadingID, &b.GlucoseReading
		},
		// A glucose of 0 or less is no reading, as AAPS sends 0 with a Bolus Wizard that used no BG
		has:     func(source nightscoutmodels.Treatment) bool { return source.Glucose != nil && *source.Glucose > 0 },
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
				case nightscoutmodels.Finger, nightscoutmodels.Meter, nightscoutmodels.Manual:
					target.ReadingType = "mbg"
				}
			}
		},
		// A reading the site put in the group, such as a CGM's, keeps what a treatment does not have of it, and its type
		// when it has no glucose type to tell it
		kept: func(stored, event *sitedb.GlucoseReading) {
			event.Delta, event.Direction, event.Noise, event.Filtered = stored.Delta, stored.Direction, stored.Noise, stored.Filtered
			event.Unfiltered, event.Rssi, event.IsCalibration = stored.Unfiltered, stored.Rssi, stored.IsCalibration
			event.SensorMgdlAtPairing = stored.SensorMgdlAtPairing
			if event.MeasurementMethod == nil {
				event.ReadingType = stored.ReadingType
			}
		},
		toTreatment: nightscoutmodels.UpdateGlucoseReadingToTreatment,
		toCreate:    nightscoutmodels.GlucoseReadingCreate, create: (*sitedb.Queries).CreateGlucoseReading,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateGlucoseReadingParams) (*sitedb.GlucoseReading, error) {
			params.ID = id
			return q.RewriteGlucoseReading(ctx, sitedb.RewriteGlucoseReadingParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteGlucoseReading,
	}
	insulinDoses = eventKind[sitedb.InsulinDose, sitedb.CreateInsulinDoseParams, sitedb.HardDeleteInsulinDoseParams]{
		name: "insulin dose",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.InsulinDose) {
			return &b.EventGroup.InsulinDoseID, &b.InsulinDose
		},
		has:         func(source nightscoutmodels.Treatment) bool { return isAssigned(source.Insulin) },
		partial:     nightscoutmodels.PartialTreatmentToInsulinDose,
		toTreatment: withoutSite(nightscoutmodels.UpdateInsulinDoseToTreatment),
		toCreate:    nightscoutmodels.InsulinDoseCreate, create: (*sitedb.Queries).CreateInsulinDose,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateInsulinDoseParams) (*sitedb.InsulinDose, error) {
			params.ID = id
			return q.RewriteInsulinDose(ctx, sitedb.RewriteInsulinDoseParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteInsulinDose,
	}
	foodIntakes = eventKind[sitedb.FoodIntake, sitedb.CreateFoodIntakeParams, sitedb.HardDeleteFoodIntakeParams]{
		name: "food intake",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.FoodIntake) {
			return &b.EventGroup.FoodIntakeID, &b.FoodIntake
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions)
		},
		partial:     nightscoutmodels.PartialTreatmentToFoodIntake,
		toTreatment: withoutSite(nightscoutmodels.UpdateFoodIntakeToTreatment),
		toCreate:    nightscoutmodels.FoodIntakeCreate, create: (*sitedb.Queries).CreateFoodIntake,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateFoodIntakeParams) (*sitedb.FoodIntake, error) {
			params.ID = id
			return q.RewriteFoodIntake(ctx, sitedb.RewriteFoodIntakeParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteFoodIntake,
	}
	tempBasals = eventKind[sitedb.TempBasal, sitedb.CreateTempBasalParams, sitedb.HardDeleteTempBasalParams]{
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
				target.Rate = *rate
			}
		},
		toTreatment: withoutSite(nightscoutmodels.UpdateTempBasalToTreatment),
		toCreate:    nightscoutmodels.TempBasalCreate, create: (*sitedb.Queries).CreateTempBasal,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateTempBasalParams) (*sitedb.TempBasal, error) {
			params.ID = id
			return q.RewriteTempBasal(ctx, sitedb.RewriteTempBasalParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteTempBasal,
	}
	tempTargets = eventKind[sitedb.TempTarget, sitedb.CreateTempTargetParams, sitedb.HardDeleteTempTargetParams]{
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
			if source.DurationMinutes == nil && source.DurationType != nil && *source.DurationType == indefinite {
				source.DurationMinutes = new(types.FlexibleInt64(indefiniteMinutes))
			}
			return source
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.HighTarget, source.LowTarget, source.DurationMinutes)
		},
		partial: nightscoutmodels.PartialTreatmentToTempTarget,
		// The targets are kept in mg/dL, converted from the units sent, or guessed from them as glucose is, as AAPS sends
		// "mmol" and Nightscout reads one under 20 with no units as mmol/L
		complete: func(source nightscoutmodels.Treatment, target *sitedb.TempTarget) {
			if source.HighTarget != nil {
				target.HighTargetMgdl = source.Units.ToMGDL(source.HighTarget.Float64())
			} else if source.LowTarget != nil {
				// If Target Top is missing then we assume both targets are the same
				target.HighTargetMgdl = source.Units.ToMGDL(source.LowTarget.Float64())
			}

			if source.LowTarget != nil {
				target.LowTargetMgdl = source.Units.ToMGDL(source.LowTarget.Float64())
			} else if source.HighTarget != nil {
				// If Target Bottom is missing then we assume both targets are the same
				target.LowTargetMgdl = source.Units.ToMGDL(source.HighTarget.Float64())
			}
		},
		// One a week long is given back as indefinite when it was sent so, as LoopFollow shows Loop's that way, from its
		// duration_type; one changed since to end sooner is not
		toTreatment: func(event *sitedb.TempTarget, _ *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
			nightscoutmodels.UpdateTempTargetToTreatment(event, treatment)
			// Its targets are given back in mg/dL, whatever the site's units
			if event != nil {
				treatment.Units = new(consts.UnitsMGDL)
			}
			if event == nil || event.DurationMinutes != indefiniteMinutes || treatment.DurationType == nil || *treatment.DurationType != indefinite {
				treatment.DurationType = nil
			}
		},
		toCreate: nightscoutmodels.TempTargetCreate, create: (*sitedb.Queries).CreateTempTarget,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateTempTargetParams) (*sitedb.TempTarget, error) {
			params.ID = id
			return q.RewriteTempTarget(ctx, sitedb.RewriteTempTargetParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteTempTarget,
	}
	profileSwitches = eventKind[sitedb.ProfileSwitch, sitedb.CreateProfileSwitchParams, sitedb.HardDeleteProfileSwitchParams]{
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
			if aapsTimeshift(source.OriginalProfileName, source.Timeshift) {
				target.Timeshift = ptr(int64(*source.Timeshift) / msPerMinute)
			}

			if source.ProfileJSON != nil {
				var profile nightscoutmodels.Profile
				err := json.Unmarshal([]byte(*source.ProfileJSON), &profile)
				if err == nil {

					profile.ConvertToMGDL()

					// AAPS's effective profile switch has its insulin only in its profileJson's iCfg
					var withICfg struct {
						ICfg *nightscoutmodels.InsulinConfig `json:"iCfg"`
					}
					if target.IcfgJson == nil && json.Unmarshal([]byte(*source.ProfileJSON), &withICfg) == nil {
						target.IcfgJson = nightscoutmodels.InsulinConfigToJSON(withICfg.ICfg)
					}

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
		toTreatment: func(event *sitedb.ProfileSwitch, siteContext *httputils.SiteContext, treatment *nightscoutmodels.Treatment) {
			nightscoutmodels.UpdateProfileSwitchToTreatment(event, siteContext, treatment)
			if event != nil && aapsTimeshift(event.OriginalProfileName, event.Timeshift) {
				treatment.Timeshift = ptr(types.FlexibleInt64(*event.Timeshift * msPerMinute))
			}
		},
		toCreate: nightscoutmodels.ProfileSwitchCreate, create: (*sitedb.Queries).CreateProfileSwitch,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateProfileSwitchParams) (*sitedb.ProfileSwitch, error) {
			params.ID = id
			return q.RewriteProfileSwitch(ctx, sitedb.RewriteProfileSwitchParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteProfileSwitch,
	}
	activities = eventKind[sitedb.Activity, sitedb.CreateActivityParams, sitedb.HardDeleteActivityParams]{
		name: "activity",
		of: func(b *sitedb.EventGroupsBundle) (**string, **sitedb.Activity) {
			return &b.EventGroup.ActivityID, &b.Activity
		},
		has: func(source nightscoutmodels.Treatment) bool {
			return isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate)
		},
		partial:     nightscoutmodels.PartialTreatmentToActivity,
		toTreatment: withoutSite(nightscoutmodels.UpdateActivityToTreatment),
		toCreate:    nightscoutmodels.ActivityCreate, create: (*sitedb.Queries).CreateActivity,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateActivityParams) (*sitedb.Activity, error) {
			params.ID = id
			return q.RewriteActivity(ctx, sitedb.RewriteActivityParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteActivity,
	}
	notificationEvents = eventKind[sitedb.NotificationEvent, sitedb.CreateNotificationEventParams, sitedb.HardDeleteNotificationEventParams]{
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
		toTreatment: withoutSite(nightscoutmodels.UpdateNotificationEventToTreatment),
		toCreate:    nightscoutmodels.NotificationEventCreate, create: (*sitedb.Queries).CreateNotificationEvent,
		rewrite: func(q *sitedb.Queries, ctx context.Context, id string, params sitedb.CreateNotificationEventParams) (*sitedb.NotificationEvent, error) {
			params.ID = id
			return q.RewriteNotificationEvent(ctx, sitedb.RewriteNotificationEventParams(params))
		},
		delete: (*sitedb.Queries).HardDeleteNotificationEvent,
	}
)

// treatmentEvents are every kind of event a group can have, in the order they are written
var treatmentEvents = []treatmentEvent{
	glucoseReadings, insulinDoses, foodIntakes, tempBasals, tempTargets, profileSwitches, activities, notificationEvents,
}

// treatmentTypeEvents are the kinds of event each type of treatment keeps, see sitemodels.EventGroupTypeKinds; its other
// fields are dropped
var treatmentTypeEvents = func() map[nightscoutmodels.TreatmentEventType][]treatmentEvent {
	byField := map[string]treatmentEvent{"glucose_reading_id": glucoseReadings, "insulin_dose_id": insulinDoses,
		"food_intake_id": foodIntakes, "temp_basal_id": tempBasals, "temp_target_id": tempTargets,
		"profile_switch_id": profileSwitches, "activity_id": activities, "notification_event_id": notificationEvents}
	types := map[nightscoutmodels.TreatmentEventType][]treatmentEvent{}
	for t, fields := range sitemodels.EventGroupTypeKinds {
		for _, field := range fields {
			types[t] = append(types[t], byField[field])
		}
	}
	return types
}()

// percentSwitchMs is how long before a temp basal a profile switch with a duration is looked for, as one still in force,
// and percentSwitches the most switches looked at
const percentSwitchMs, percentSwitches = 7 * 24 * 60 * 60_000, 1000

// maxMultiplier is the most times the basal a temp basal keeps, as Nightscout's 1000 percent
const maxMultiplier = 11

// fillTempBasal sets what the treatment doesn't send of a temp basal's rate in U/hr and its multiplier, from the other
// and the profile's basal at its start, see sitemodels.BasalAt, keeping them as they are when no profile has a basal
// then
func fillTempBasal(ctx context.Context, q *sitedb.Queries, treatment nightscoutmodels.Treatment, tempBasal *sitedb.TempBasal) error {
	sentRate, sentPercent := nightscoutmodels.TempBasalValuePtr(treatment) != nil, treatment.TempBasalPercent != nil
	if tempBasal == nil || sentRate == sentPercent || (sentPercent && tempBasal.Multiplier == nil) {
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
	basal := sitemodels.BasalAt(at, location, defaults, switches)
	switch {
	case basal == nil:
	case sentPercent:
		tempBasal.Rate = consts.Round(*basal**tempBasal.Multiplier, 3)
	// A rate far over the basal, as a temp basal set over a basal of nearly nothing, has no multiplier kept
	case *basal > 0 && tempBasal.Rate <= maxMultiplier**basal:
		tempBasal.Multiplier = new(consts.Round(tempBasal.Rate / *basal, 5))
	}
	return nil
}
