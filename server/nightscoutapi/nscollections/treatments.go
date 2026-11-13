package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

type TreatmentCollection struct {
	named[nightscoutmodels.Treatment, sitedb.EventGroupsBundle]
}

func NewTreatmentCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Treatment, sitedb.EventGroupsBundle](TreatmentCollection{
		named[nightscoutmodels.Treatment, sitedb.EventGroupsBundle]{nsutils.Treatments, DBModelToTreatment},
	})
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

// Rows are its event group and its events, so insulin far outside real use refuses it, saying why, as xDrip+ drops a
// treatment it is refused, and every one queued after it
func (c TreatmentCollection) Rows(nsModel nightscoutmodels.Treatment) []sitedb.TableRow {
	if nsModel.EventType == nil {
		return nil
	}
	bundle := TreatmentEventGroupBundle(context.Background(), nsModel)
	rows := []sitedb.TableRow{{Table: "event_groups", Row: bundle.EventGroup}}
	for _, event := range treatmentEvents {
		if row, ok := event.row(&bundle); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func (c TreatmentCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.EventGroupsBundle, error) {

	tebg := TreatmentEventGroupBundle(ctx, nsModel)
	if err := fillTempBasal(ctx, q, nsModel, tebg.TempBasal); err != nil {
		return nil, err
	}
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
		slog.ErrorContext(ctx, "failed to create event group in database", "error", err, "createParams", tebg)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.ErrorContext(ctx, "failed to create event group in database, EventGroup is nil", "bundle", tebg)
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	tebg.EventGroup = *newEventGroup

	return &tebg, nil
}

// FallbackKey is a treatment's date, device, eventType, insulin and carbs, a zero being none, as Nightscout matches one
// sent without an identifier (treatment-fallback-key.js), of one stored without an app's id, xDrip+'s uuid or Loop's
// syncIdentifier. So a careportal note in the same second as xDrip+'s bolus is kept beside it rather than taking it
// over, and AAPS's Bolus Wizard's bolus and its carbs, sent as two Meal Boluses at the same date and deleted on their
// own by their identifiers, are two treatments.
func (c TreatmentCollection) FallbackKey() FallbackKey[nightscoutmodels.Treatment, sitedb.EventGroupsBundle] {
	type key = KeyField[nightscoutmodels.Treatment, sitedb.EventGroupsBundle]
	return FallbackKey[nightscoutmodels.Treatment, sitedb.EventGroupsBundle]{
		search: searchSameTimeTreatments,
		fields: []key{
			sameDate[nightscoutmodels.Treatment, sitedb.EventGroupsBundle](),
			sameDevice[nightscoutmodels.Treatment, sitedb.EventGroupsBundle](),
			{"eventType", func(sent nightscoutmodels.Treatment, stored *sitedb.EventGroupsBundle) bool {
				return sent.EventType == nil || sent.EventType.String() == stored.EventGroup.GroupType
			}},
			{"insulin", func(sent nightscoutmodels.Treatment, stored *sitedb.EventGroupsBundle) bool {
				var dose float64
				if stored.InsulinDose != nil {
					dose = stored.InsulinDose.Dose
				}
				return amount(sent.Insulin) == dose
			}},
			{"carbs", func(sent nightscoutmodels.Treatment, stored *sitedb.EventGroupsBundle) bool {
				var carbs float64
				if stored.FoodIntake != nil && stored.FoodIntake.Carbs != nil {
					carbs = *stored.FoodIntake.Carbs
				}
				return amount(sent.Carbs) == carbs
			}},
		},
		clientIDs: []func(*sitedb.EventGroupsBundle) *string{
			func(stored *sitedb.EventGroupsBundle) *string { return stored.EventGroup.Uuid },
			func(stored *sitedb.EventGroupsBundle) *string { return stored.EventGroup.SyncIdentifier },
		},
	}
}

// searchSameTimeTreatments finds the treatments of a type and device at the date of the one sent
func searchSameTimeTreatments(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, limit int64) ([]*sitedb.EventGroupsBundle, error) {
	eventGroups, err := q.SearchEventGroups(ctx, sitedb.SearchEventGroupsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		TypeEq:   nsModel.EventType.Stringp(),
		Limit:    limit,
	})
	bundles := make([]*sitedb.EventGroupsBundle, len(eventGroups))
	for i, eventGroup := range eventGroups {
		bundle := sitedb.GetEventGroupToEventGroupsBundle(sitedb.SearchEventGroupToGetEventGroup(*eventGroup))
		bundles[i] = &bundle
	}
	return bundles, err
}

// amount is a treatment's carbs or insulin, 0 when it has none
func amount(value *types.FlexibleFloat64) float64 {
	if value == nil {
		return 0
	}
	return value.Float64()
}

// Refreshes a treatment sent again with changes, by the hash of its JSON: Loop sends a temp basal or bolus again when it
// ends with what was delivered, and Trio a temp target that ended early with its real duration. One of another type,
// or stored without that hash, is left as it is.
func (c TreatmentCollection) Refreshes(nsModel nightscoutmodels.Treatment, stored *sitedb.EventGroupsBundle) bool {
	sent, resent := stored.EventGroup.SentHash, nightscoutmodels.SentHash(nsModel.SentJSON)
	return nsModel.EventType.String() == stored.EventGroup.GroupType && sent != nil && resent != nil && *sent != *resent
}

func (c TreatmentCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.EventGroupsBundle, error) {
	dbModel, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: identifier, IncludeDeleted: includeDeleted})
	if err != nil {
		return nil, err
	}

	egb := sitedb.GetEventGroupToEventGroupsBundle(*dbModel)

	return &egb, nil
}

// RewriteModel rewrites the group and its events: an event the treatment still has keeps its identifier and what of it
// a treatment does not have, one it no longer has is deleted once the group no longer refers to it, and one it now has
// is created. The hash it keeps is of the fields sent, as a create's is of the treatment sent.
func (c TreatmentCollection) RewriteModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, sent map[string]any, stored *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) (*sitedb.EventGroupsBundle, error) {
	bundle := TreatmentEventGroupBundle(ctx, nsModel)
	for _, event := range treatmentEvents {
		event.keep(&bundle, stored)
	}
	if err := fillTempBasal(ctx, q, nsModel, bundle.TempBasal); err != nil {
		return nil, err
	}
	if err := createUpdateEventGroup(ctx, q, &bundle, cc); err != nil {
		return nil, err
	}
	params := nightscoutmodels.EventGroupCreate(bundle.EventGroup, cc)
	sentJSON, _ := json.Marshal(sent)
	params.ID, params.SentHash = stored.EventGroup.ID, nightscoutmodels.SentHash(string(sentJSON))
	group, err := q.RewriteEventGroup(ctx, sitedb.RewriteEventGroupParams(params))
	if err != nil {
		slog.ErrorContext(ctx, "failed to rewrite event group in database", "error", err)
		return nil, err
	}
	bundle.EventGroup = *group
	for _, event := range treatmentEvents {
		if err := event.drop(ctx, q, &bundle, stored); err != nil {
			return nil, err
		}
	}
	return &bundle, nil
}

func (c TreatmentCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error) {
	if permanent {
		return hardDeleteEventGroup(ctx, q, identifier)
	}
	return softDeleteEventGroup(ctx, q, identifier)
}

// DeleteType deletes the treatments of the type, and when permanent those already soft deleted too. Each is deleted with
// its events one at a time, so a request deletes at most a search's worth, the newest first, and says how many: the
// rest are deleted by asking again.
func (c TreatmentCollection) DeleteType(q *sitedb.Queries, ctx context.Context, typeString string, permanent bool) (sql.Result, error) {
	eventGroups, err := q.GetEventGroupOfTypeWithoutJoins(ctx, sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: typeString, IncludeDeleted: permanent, Limit: apiquery.MaxQueryLimit})
	if err != nil {
		return nil, err
	}
	for _, eg := range eventGroups {
		if eg == nil {
			continue
		}
		if _, err := c.Delete(q, ctx, eg.ID, permanent); err != nil {
			return nil, err
		}
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: int64(len(eventGroups)), Err: nil}, nil
}

func (c TreatmentCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.EventGroupAscendingIDs(ctx, nsquery.EventGroupAscending(query))
}

func (c TreatmentCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.EventGroupHistoryIDs(ctx, sitedb.EventGroupHistoryIDsParams{Since: since, Limit: limit})
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

// createUpdateEventGroup writes each of a bundle's events to the DB, see treatmentEvent.write
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
		slog.ErrorContext(ctx, "failed to hard delete eventGroup event group by ID", "error", err)
		return nil, err
	}

	if eventGroup == nil {
		slog.ErrorContext(ctx, "failed to hard delete eventGroup event group by ID, eventGroup is nil", "identifier", identifier)
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
		slog.ErrorContext(ctx, "failed to soft delete event group by ID", "error", err, "identifier", identifier)
		return nil, err
	}
	return nsutils.SQLResultCustomResult{LastInsertIdInt: 0, RowsAffectedInt: 1, Err: nil}, nil
}

// Converts an incomming treatment to EventGroups, keeping only the events relevant to its type
func TreatmentEventGroupBundle(ctx context.Context, treatment nightscoutmodels.Treatment) sitedb.EventGroupsBundle {
	// A duration of 0 is until cancelled for one sent indefinite, and for a Temporary Override, as Nightscout's
	// careportal sends one, so it is kept as indefinite rather than ended as it starts
	isIndefinite := treatment.DurationType != nil && *treatment.DurationType == indefinite
	if d := treatment.DurationMinutes; d != nil && *d == 0 && (isIndefinite || *treatment.EventType == nightscoutmodels.TemporaryOverride) {
		treatment.DurationMinutes, treatment.DurationInMilliseconds = nil, nil
		treatment.DurationType = new(types.ShortString(indefinite))
	}
	bundle := sitedb.EventGroupsBundle{EventGroup: nightscoutmodels.TreatmentToEventGroup(treatment)}
	for _, event := range treatmentTypeEvents[*treatment.EventType] {
		event.fromTreatment(treatment, &bundle)
	}
	return bundle
}

// isAssigned is whether any of the values is set, which kinds of events use to tell if a treatment has one
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
	// Every treatment says the units its glucose is in, as the API's, so a client does not read a temp target's or a
	// therapy event's glucose in the site's units, as Nightscout's careportal sends units with every treatment
	treatment.Units = siteContext.GetSettings().GetUnits()
	for _, event := range treatmentEvents {
		event.addToTreatment(source, siteContext, treatment)
	}
	// The duration sent in milliseconds is given back, as AAPS reads it before the minutes, while it is within a minute
	// of the minutes kept, so not once those are changed on the site
	if ms := source.EventGroup.DurationMs; ms != nil {
		if minutes := treatment.DurationMinutes; minutes == nil || max(*ms-int64(*minutes)*60_000, int64(*minutes)*60_000-*ms) < 60_000 {
			treatment.DurationInMilliseconds = new(types.FlexibleInt64(*ms))
		}
	}
	// A duration is given back both in minutes and in milliseconds, whichever was kept, as a client may read either
	if treatment.DurationInMilliseconds == nil && treatment.DurationMinutes != nil {
		treatment.DurationInMilliseconds = new(*treatment.DurationMinutes * 60_000)
	}
	if treatment.DurationMinutes == nil && treatment.DurationInMilliseconds != nil {
		treatment.DurationMinutes = new((*treatment.DurationInMilliseconds + 30_000) / 60_000)
	}
	// The timestamp is when it happened in UTC, as created_at is written and every client sends it, whatever its
	// utcOffset: LoopFollow reads it before created_at and drops its offset, reading the rest as UTC
	treatment.Timestamp = new(types.DateString(types.ISOString(time.UnixMilli(source.EventGroup.OccurredAt))))
	// Nighttroop keeps the insulin given and a temp basal's absolute rate, so they are what was programmed
	treatment.Programmed = treatment.Insulin
	if treatment.Absolute != nil {
		treatment.Temp = new(types.ShortString("absolute"))
	}
	return *treatment
}
