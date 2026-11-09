package nscollections

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"

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

func (c TreatmentCollection) ValidatePatch(nsModel nightscoutmodels.Treatment, dbModel *sitedb.EventGroupsBundle) *nsutils.HTTPError {
	if nsModel.EventType != nil && nsModel.EventType.String() != dbModel.EventGroup.GroupType {
		slog.Error("treatment event type is not good", "eventType", nsModel.EventType.String())
		return nsutils.BadDataBodyError
	}
	return nil
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

func (c TreatmentCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment) (*sitedb.EventGroupsBundle, error) {
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

// Refresh replaces a treatment sent again with the one sent, as Nightscout upserts a treatment on its created_at and
// eventType: Loop sends a temp basal or bolus again when it ends with what was delivered, and Trio a temp target that
// ended early with its real duration. One of another type, sent the same as it was, by the hash of its JSON, or
// stored without that hash, is left as it is.
func (c TreatmentCollection) Refresh(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Treatment, stored *sitedb.EventGroupsBundle, cc nightscoutmodels.ConverterContext) (*sitedb.EventGroupsBundle, bool, error) {
	sent, resent := stored.EventGroup.OtherAttributes.SentHash, nightscoutmodels.SentHash(nsModel.SentJSON)
	if nsModel.EventType.String() != stored.EventGroup.GroupType || sent == nil || resent == nil || *sent == *resent {
		return stored, false, nil
	}
	if _, err := hardDeleteEventGroup(ctx, q, stored.EventGroup.ID); err != nil {
		return nil, false, err
	}
	refreshed, err := c.CreateModel(q, ctx, nsModel, cc, &stored.EventGroup.ID)
	return refreshed, true, err
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
	if err := fillTempBasal(ctx, q, nsModel, dbModel.TempBasal); err != nil {
		return nil, err
	}

	// Do something with the update bundle to merge it into the dbModel

	err := createUpdateEventGroup(ctx, q, dbModel, cc)
	if err != nil {
		return nil, err
	}

	updateEventGroupParams := nightscoutmodels.EventGroupUpdate(dbModel.EventGroup, cc)
	newEventGroup, err := q.UpdateEventGroup(ctx, updateEventGroupParams)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update event group in database", "error", err)
		return nil, err
	}

	if newEventGroup == nil || newEventGroup.ID == "" {
		slog.ErrorContext(ctx, "failed to update event group in database, EventGroup is nil")
		return nil, fmt.Errorf("newEventGroup is nil or has empty ID")
	}

	dbModel.EventGroup = *newEventGroup

	return dbModel, nil
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
	for _, event := range treatmentEvents {
		event.addToTreatment(source, siteContext, treatment)
	}
	// Nighttroop keeps the insulin given and a temp basal's absolute rate, so they are what was programmed
	treatment.Timestamp = treatment.CreatedAt
	treatment.Programmed = treatment.Insulin
	if treatment.Absolute != nil {
		treatment.Temp = new(types.ShortString("absolute"))
	}
	return *treatment
}
