package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	converter func(c *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewTreatmentCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Treatment, sitedb.EventGroupsBundle](TreatmentCollection{
		converter: nsutils.DBModelToRecord(nsutils.Treatments, DBModelToTreatment),
	})
}

func (c TreatmentCollection) Name() string {
	return "treatments"
}

func (c TreatmentCollection) Convert(dr *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c TreatmentCollection) Validate(nsModel nightscoutmodels.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	// Validate the input
	if nsModel.EventType == nil || *nsModel.EventType == nightscoutmodels.Unknown {
		eventTypeError := "eventType: null"
		if nsModel.EventType != nil {
			eventTypeError = "eventType: " + nsModel.EventType.String()
		}
		slog.Error("treatment EventType is empty or unknown", "error", eventTypeError)
		return nsutils.BadDataBodyErrorWithHint("EventType is nil or unknown: " + eventTypeError)
	}

	// Validate the change
	if dbModel == nil {
		return nil
	}

	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("cannot change event_type attribute on treatment", "existingEventType", dbModel.EventGroup.GroupType, "newEventType", nsModel.EventType.String())
		return nsutils.ErrCannotChangeEventType
	}

	return nil
}

func (c TreatmentCollection) ValidatePatch(nsModel nightscoutmodels.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}
	return nil
}

func (c TreatmentCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.EventGroupsBundle, error) {

	tebg := TreatmentEventGroupBundle(ctx, nsModel)
	err := createUpdateEventGroup(ctx, q, &tebg, cc)
	if err != nil {
		return nil, err
	}

	createEventGroupParams := nightscoutmodels.EventGroupCreate(tebg.EventGroup, cc)

	if overrideIdentifier != nil {
		createEventGroupParams.ID = *overrideIdentifier
	}
	newEventGroup, err := q.CreateEventGroup(ctx, createEventGroupParams)
	if err != nil {
		slog.Error("failed to create event group in database", "error", err, "createParams", tebg)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.Error("failed to create event group in database, EventGroup is nil", "bundle", tebg)
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	tebg.EventGroup = *newEventGroup

	return &tebg, nil
}

func (c TreatmentCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment) (*sitedb.EventGroupsBundle, error) {
	// Without a date every filter below is empty and would match the newest row
	if nsModel.Date == nil {
		return nil, nil
	}
	eventGroups, rerr := q.SearchEventGroups(ctx, sitedb.SearchEventGroupsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		TypeEq:   nsModel.EventType.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(eventGroups) > 0 {
		egb := sitedb.GetEventGroupToEventGroupsBundle(sitedb.SearchEventGroupToGetEventGroup(*eventGroups[0]))
		return &egb, nil
	}
	return nil, nil
}

func (c TreatmentCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: identifier, IncludeDeleted: includeDeleted})
	if err != nil {
		return nil, err
	}

	egb := sitedb.GetEventGroupToEventGroupsBundle(*dbModel)

	return &egb, nil
}

func (c TreatmentCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, identifier string, dbModel *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) (*sitedb.EventGroupsBundle, error) {
	// This will be a partial update

	nsModel.EventType = ptr(nightscoutmodels.TreatmentEventType(dbModel.EventGroup.GroupType)) // Ensure the event type is set to the existing value

	updateBundle := TreatmentEventGroupBundle(ctx, nsModel)

	MergeUpdateBundle(dbModel, updateBundle)

	// Do something with the update bundle to merge it into the dbModel

	err := createUpdateEventGroup(ctx, q, dbModel, cc)
	if err != nil {
		return nil, err
	}

	updateEventGroupParams := nightscoutmodels.EventGroupUpdate(dbModel.EventGroup, cc)
	newEventGroup, err := q.UpdateEventGroup(ctx, updateEventGroupParams)
	if err != nil {
		slog.Error("failed to update event group in database", "error", err)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.Error("failed to update event group in database, EventGroup is nil")
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	dbModel.EventGroup = *newEventGroup

	return dbModel, nil
}

func (c TreatmentCollection) HardDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return hardDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) SoftDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return softDeleteEventGroup(ctx, q, identifier)
}

func (c TreatmentCollection) HardDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {

	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString, IncludeDeleted: true})
	if err != nil {
		return nil, err
	}

	for _, eg := range eventGroups {
		if eg != nil {
			if _, err := hardDeleteEventGroup(ctx, q, eg.ID); err != nil {
				return nil, err
			}

		}
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

func (c TreatmentCollection) SoftDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString})
	if err != nil {
		return nil, err
	}

	for _, eg := range eventGroups {
		if eg != nil {
			if _, err := softDeleteEventGroup(ctx, q, eg.ID); err != nil {
				return nil, err
			}

		}
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

func (c TreatmentCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.EventGroupsBundle, error) {
	eventGroups, err := q.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if err != nil {
		return nil, err
	}
	res := []*sitedb.EventGroupsBundle{}
	for _, eg := range eventGroups {
		if eg != nil {
			egb := sitedb.GetEventGroupToEventGroupsBundle(sitedb.SearchEventGroupToGetEventGroup(*eg))
			res = append(res, &egb)
		}
	}
	return res, nil
}

// //////////////////////
// Private Methods
// ///////////////////////

// createUpdateEventGroup syncs each of a bundle's events to the DB, see treatmentEvent.write
func createUpdateEventGroup(ctx context.Context, q *sitedb.Queries, bundle *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) error {
	for _, event := range treatmentEvents {
		if err := event.write(ctx, q, bundle, cc); err != nil {
			return err
		}
	}
	return nil
}

func hardDeleteEventGroup(ctx context.Context, q *sitedb.Queries, identifier string) (sql.Result, error) {
	eventGroup, err := q.HardDeleteEventGroup(ctx, sitedb.HardDeleteEventGroupParams{ID: identifier})
	if err != nil {
		slog.Error("failed to hard delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.Error("failed to hard delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
		return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 0, Err: nil}, sql.ErrNoRows
	}

	for _, event := range treatmentEvents {
		if err := event.hardDelete(ctx, q, *eventGroup); err != nil {
			return nil, err
		}
	}

	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

func softDeleteEventGroup(ctx context.Context, q *sitedb.Queries, identifier string) (sql.Result, error) {
	if err := sitedb.SoftDeleteEventGroupAndEvents(ctx, q, identifier, nil); err != nil {
		slog.Error("failed to soft delete event group by ID", "error", err, "identifier", identifier)
		return nil, err
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

// Converts an incomming treatment to EventGroups, keeping only the events relevant to its type
func TreatmentEventGroupBundle(ctx context.Context, treatment nightscoutmodels.Treatment) sitedb.EventGroupsBundle {
	bundle := sitedb.EventGroupsBundle{EventGroup: nightscoutmodels.TreatmentToEventGroup(treatment)}
	for _, event := range treatmentTypeEvents[*treatment.EventType] {
		event.fromTreatment(treatment, &bundle)
	}
	return bundle
}

// Merge a DB Event Group with one created from the
func MergeUpdateBundle(dbModel *sitedb.EventGroupsBundle, updateBundle sitedb.EventGroupsBundle) {
	nightscoutmodels.UpdateToEventGroup(updateBundle.EventGroup, &dbModel.EventGroup)
	for _, event := range treatmentEvents {
		event.merge(dbModel, updateBundle)
	}
}

////////////////////////////////////////
// TREATMENT TO CREATE/UPDATE
////////////////////////////////////////

///////
// Food Intake
///////

func TreatmentToFoodIntake(source nightscoutmodels.Treatment) *sitedb.FoodIntake {
	// Check some needed fields are assigned
	if isAssigned(source.FoodType, source.PreBolus, source.Category, source.Subcategory, source.Carbs, source.Fat, source.Protein, source.Fibre, source.Energy, source.Portions) {
		return ptr(nightscoutmodels.PartialTreatmentToFoodIntake(source))
	}
	return nil
}

///////
// Glucose Reading
///////

func TreatmentToGlucoseReading(source nightscoutmodels.Treatment) *sitedb.GlucoseReading {
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
	if isAssigned(source.Glucose) {
		return &target
	}
	return nil
}

///////
// Insulin Dose
///////

func TreatmentToInsulinDose(source nightscoutmodels.Treatment) *sitedb.InsulinDose {
	if isAssigned(source.Insulin) {
		return ptr(nightscoutmodels.PartialTreatmentToInsulinDose(source))
	}
	return nil
}

///////
// NotificationEvents
///////

func TreatmentToNotificationEvent(source nightscoutmodels.Treatment) *sitedb.NotificationEvent {
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

	// Always is assigned
	if source.Name != nil {
		target.Title = source.Name.String()
	}

	return &target
}

///////
// Profile Switch
///////

func TreatmentToProfileSwitch(source nightscoutmodels.Treatment) *sitedb.ProfileSwitch {
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
	if isAssigned(source.Name, source.Profile, source.ProfileJSON) {
		return &target
	}
	return nil
}

///////
// Temp Basal
///////

func TreatmentToTempBasal(source nightscoutmodels.Treatment) *sitedb.TempBasal {
	if isAssigned(source.Absolute, source.Rate, source.Percent, source.DurationMinutes) {
		return ptr(nightscoutmodels.PartialTreatmentToTempBasal(source))
	}
	return nil
}

///////
// Temp Targets
///////

func TreatmentToTempTarget(source nightscoutmodels.Treatment) *sitedb.TempTarget {

	if len(source.CorrectionRange) == 2 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[1]
	} else if len(source.CorrectionRange) == 1 {
		source.LowTarget = source.CorrectionRange[0]
		source.HighTarget = source.CorrectionRange[0]
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

	if isAssigned(source.HighTarget, source.LowTarget, source.DurationMinutes) {
		return &target
	}
	return nil
}

func TreatmentToComponentEvent(source nightscoutmodels.Treatment) *sitedb.ComponentEvent {
	if isAssigned(source.PumpID, source.PumpSerial, source.PumpType) {
		return ptr(nightscoutmodels.PartialTreatmentToComponentEvent(source))
	}
	return nil
}

func TreatmentToActivity(source nightscoutmodels.Treatment) *sitedb.Activity {
	if isAssigned(source.Name, source.ActivityType, source.Calories, source.Steps, source.HeartRate) {
		return ptr(nightscoutmodels.PartialTreatmentToActivity(source))
	}
	return nil
}

func isAssigned(values ...any) bool {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func:
			if !v.IsNil() {
				return true
			}
		default:
			return true
		}
	}

	return false
}

func DBModelToTreatment(source *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) nightscoutmodels.Treatment {
	treatment := nightscoutmodels.EventGroupToTreatment(source.EventGroup)

	nightscoutmodels.UpdateGlucoseReadingToTreatment(source.GlucoseReading, siteContext, treatment)
	nightscoutmodels.UpdateInsulinDoseToTreatment(source.InsulinDose, treatment)
	nightscoutmodels.UpdateFoodIntakeToTreatment(source.FoodIntake, treatment)
	nightscoutmodels.UpdateTempBasalToTreatment(source.TempBasal, treatment)
	nightscoutmodels.UpdateTempTargetToTreatment(source.TempTarget, treatment)
	nightscoutmodels.UpdateProfileSwitchToTreatment(source.ProfileSwitch, siteContext, treatment)
	nightscoutmodels.UpdateComponentEventToTreatment(source.ComponentEvent, treatment)
	nightscoutmodels.UpdateActivityToTreatment(source.Activity, treatment)
	nightscoutmodels.UpdateNotificationEventToTreatment(source.NotificationEvent, treatment)

	return *treatment
}
