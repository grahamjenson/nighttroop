package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// maxEventsBody is the most bytes a cold day's body is, as its column takes
const maxEventsBody = 16384

// eventRows is the most rows of one table a day's events are read from
const eventRows = 5000

// anys are rows of any table as one type, with the error of the query that read them
func anys[T any](rows []*T, err error) ([]any, error) {
	all := make([]any, len(rows))
	for i, row := range rows {
		all[i] = row
	}
	return all, err
}

// eventSource is a table whose rows are events: how a day's are read, which of them are events, and the member of a
// treatment that refers to one
type eventSource struct {
	table  string
	rows   func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error)
	keep   func(row any) bool
	member func(group *sitedb.EventGroup) *string
}

// eventSources are the tables whose rows are events. Not every row is one: a CGM reading is in the glucose series,
// unless it has a note, and a loop's own dose is in the insulin series. A temp basal is never one.
var eventSources = []eventSource{
	{"glucose_readings", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: &from, DateLT: &to, Limit: eventRows}))
	}, func(row any) bool { r := row.(*sitedb.GlucoseReading); return r.ReadingType == "mbg" || r.Notes != nil },
		func(g *sitedb.EventGroup) *string { return g.GlucoseReadingID }},
	{"insulin_doses", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchInsulinDoses(ctx, sitedb.SearchInsulinDosesParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, func(row any) bool { r := row.(*sitedb.InsulinDose); return r.Automatic == nil || !*r.Automatic },
		func(g *sitedb.EventGroup) *string { return g.InsulinDoseID }},
	{"food_intakes", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchFoodIntakes(ctx, sitedb.SearchFoodIntakesParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, nil, func(g *sitedb.EventGroup) *string { return g.FoodIntakeID }},
	{"activities", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchActivities(ctx, sitedb.SearchActivitiesParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, nil, func(g *sitedb.EventGroup) *string { return g.ActivityID }},
	{"temp_targets", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchTempTargets(ctx, sitedb.SearchTempTargetsParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, nil, func(g *sitedb.EventGroup) *string { return g.TempTargetID }},
	{"profile_switches", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchProfileSwitches(ctx, sitedb.SearchProfileSwitchesParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, nil, func(g *sitedb.EventGroup) *string { return g.ProfileSwitchID }},
	{"notification_events", func(ctx context.Context, q *sitedb.Queries, from, to int64) ([]any, error) {
		return anys(q.SearchNotificationEvents(ctx, sitedb.SearchNotificationEventsParams{DateGTE: &from, DateLT: &to, IncludeDeleted: new(false), Limit: eventRows}))
	}, nil, func(g *sitedb.EventGroup) *string { return g.NotificationEventID }},
}

// eventOf is a row as an event: "k", its table, and each column its table's policy keeps that has a value, by the
// column's name
func eventOf(table string, row any) map[string]any {
	event, keep := map[string]any{"k": table}, retention.Policies[table].Keep
	value := reflect.Indirect(reflect.ValueOf(row))
	for i := range value.NumField() {
		limit, ok := sitedb.ColumnLimitOf(table, value.Type().Field(i).Name)
		field := value.Field(i)
		if !ok || !slices.Contains(keep, limit.Column) || (field.Kind() == reflect.Pointer && field.IsNil()) {
			continue
		}
		event[limit.Column] = reflect.Indirect(field).Interface()
	}
	return event
}

const (
	// inForceLookback is how far before a day a temp target that is still in force at its start can have started,
	// the longest one lasts, and inForceRows how many of the newest rows before a day are looked through
	inForceLookback = 7 * 24 * time.Hour
	inForceRows     = 100
)

// inForce are the events from before the UTC day that starts at start, in ms, that were still in force then, the
// oldest first: the newest profile switch with no end, a newer one with an end not yet reached, and the newest temp
// target not yet over. They are read from the rows, as no event row is deleted yet.
func inForce(ctx context.Context, q *sitedb.Queries, start int64) ([]map[string]any, error) {
	var events []map[string]any
	switches, err := q.SearchProfileSwitches(ctx, sitedb.SearchProfileSwitchesParams{DateLT: &start, IncludeDeleted: new(false), Limit: inForceRows})
	if err != nil {
		return nil, err
	}
	// Newest first: each with an end not yet reached, down to the newest with no end
	for _, row := range switches {
		forever := row.DurationMinutes == 0
		if !forever && row.OccurredAt+row.DurationMinutes*time.Minute.Milliseconds() <= start {
			continue
		}
		events = append(events, eventOf("profile_switches", row))
		if forever {
			break
		}
	}
	since := start - inForceLookback.Milliseconds()
	targets, err := q.SearchTempTargets(ctx, sitedb.SearchTempTargetsParams{DateGTE: &since, DateLT: &start, IncludeDeleted: new(false), Limit: 1})
	if err != nil {
		return nil, err
	}
	// Only the newest, as it took the place of those before it
	if len(targets) > 0 && targets[0].OccurredAt+targets[0].DurationMinutes*time.Minute.Milliseconds() > start {
		events = append(events, eventOf("temp_targets", targets[0]))
	}
	slices.SortStableFunc(events, func(a, b map[string]any) int { return int(a["occurred_at"].(int64) - b["occurred_at"].(int64)) })
	return events, nil
}

// packageEventsDay writes the events of the UTC day that starts at start, in ms, from its rows, the oldest first: each
// row that is an event, with what its treatment says of it (who entered it, as what) where it has none, and each
// treatment with no event, as a site change is, and before them what was in force when the day started (inForce).
// With merge, for a day whose rows the server deletes, it keeps the events of the day packaged before whose rows are
// deleted at once (mergeEvents). A day whose other rows are deleted is not packaged again (packageTreatmentDays). A
// day with no event and nothing in force is written only to empty one that had some, or always, to say the day was
// packaged, and a day whose events are more than a cold day holds is not packaged.
func packageEventsDay(ctx context.Context, q *sitedb.Queries, start int64, always, merge bool) error {
	groups, err := q.EventGroupsOfDay(ctx, sitedb.EventGroupsOfDayParams{From: start, To: start + dayMs})
	if err != nil {
		return err
	}
	used, events := map[string]bool{}, []map[string]any{}
	for _, source := range eventSources {
		rows, err := source.rows(ctx, q, start, start+dayMs)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if source.keep != nil && !source.keep(row) {
				continue
			}
			event := eventOf(source.table, row)
			for _, group := range groups {
				if member := source.member(group); member != nil && *member == event["id"] {
					used[group.ID] = true
					// The treatment's own ID, as an app knows the treatment by it, not by its event's
					event["group_id"] = group.ID
					for column, value := range eventOf("event_groups", group) {
						if _, has := event[column]; !has {
							event[column] = value
						}
					}
				}
			}
			events = append(events, event)
		}
	}
	for _, group := range groups {
		memberless := !used[group.ID] && !slices.ContainsFunc(eventSources, func(source eventSource) bool { return source.member(group) != nil })
		if memberless {
			events = append(events, eventOf("event_groups", group))
		}
	}
	date := time.UnixMilli(start).UTC().Format(time.DateOnly)
	before, err := inForce(ctx, q, start)
	if err == nil && merge {
		events, before, err = mergeEvents(ctx, q, date, events, before)
	}
	if err != nil {
		return err
	}
	slices.SortStableFunc(events, func(a, b map[string]any) int { return int(EventAt(a) - EventAt(b)) })
	if len(events) == 0 && len(before) == 0 && !always {
		if _, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: retention.EventsKind, Date: date}); err != nil {
			return nil
		}
	}
	packed, err := retention.PackEvents(retention.EventsDay{InForce: before, Events: events})
	if err != nil {
		return err
	}
	if len(packed) > maxEventsBody {
		slog.WarnContext(ctx, "A day has more events than a cold day holds, so it is not packaged", "date", date, "events", len(events), "bytes", len(packed))
		return nil
	}
	if _, read, err := retention.ReadEvents(packed); err != nil || len(read.Events) != len(events) || len(read.InForce) != len(before) {
		return fmt.Errorf("the packaged events do not read back as they were written: %v", err)
	}
	return upsertColdDay(ctx, q, retention.EventsKind, date, packed)
}

// atOnce are the tables whose rows older than the hot days are deleted as soon as their day is packaged, as what they
// were is only in the day's events, which a day packaged again keeps (mergeEvents): each with whether a row of it,
// deleted or not, has an ID. A treatment of them alone, or of none, as a site change, is too.
var atOnce = map[string]func(ctx context.Context, q *sitedb.Queries, id string) error{
	"activities": func(ctx context.Context, q *sitedb.Queries, id string) error {
		_, err := q.GetActivity(ctx, sitedb.GetActivityParams{ID: id, IncludeDeleted: true})
		return err
	},
	"temp_targets": func(ctx context.Context, q *sitedb.Queries, id string) error {
		_, err := q.GetTempTarget(ctx, sitedb.GetTempTargetParams{ID: id, IncludeDeleted: true})
		return err
	},
	"notification_events": func(ctx context.Context, q *sitedb.Queries, id string) error {
		_, err := q.GetNotificationEvent(ctx, sitedb.GetNotificationEventParams{ID: id, IncludeDeleted: true})
		return err
	},
	"event_groups": func(ctx context.Context, q *sitedb.Queries, id string) error {
		_, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: id, IncludeDeleted: true})
		return err
	},
}

// mergeEvents are a day's events and what was in force when it started, built from its rows, with those of the day as
// it was packaged before whose rows were deleted at once (atOnce): an event of such a table that no row has the ID of,
// deleted or not, so one deleted by a person, or moved to another day, is not kept. Each event keeps its row's ID, so
// it is merged by it.
func mergeEvents(ctx context.Context, q *sitedb.Queries, date string, events, before []map[string]any) ([]map[string]any, []map[string]any, error) {
	packaged, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: retention.EventsKind, Date: date})
	if errors.Is(err, sql.ErrNoRows) {
		return events, before, nil
	}
	if err != nil {
		return nil, nil, err
	}
	_, day, err := retention.ReadEvents(packaged.Body)
	if err != nil {
		return nil, nil, err
	}
	// gone is whether an event packaged before is of a row deleted at once, and not one the day still has
	gone := func(event map[string]any, now []map[string]any) (bool, error) {
		exists, ok := atOnce[fmt.Sprint(event["k"])]
		if !ok || slices.ContainsFunc(now, func(e map[string]any) bool { return e["id"] == event["id"] }) {
			return false, nil
		}
		if err := exists(ctx, q, fmt.Sprint(event["id"])); !errors.Is(err, sql.ErrNoRows) {
			return false, err
		}
		return true, nil
	}
	// What was in force is read from the rows of the days before, so a temp target deleted at once is kept from it
	for _, merged := range []struct {
		into *[]map[string]any
		from []map[string]any
	}{{&events, day.Events}, {&before, day.InForce}} {
		now := *merged.into
		for _, event := range merged.from {
			keep, err := gone(event, now)
			if err != nil {
				return nil, nil, err
			}
			if keep {
				*merged.into = append(*merged.into, event)
			}
		}
	}
	slices.SortStableFunc(before, func(a, b map[string]any) int { return int(EventAt(a) - EventAt(b)) })
	return events, before, nil
}
