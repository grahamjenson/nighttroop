package db

import (
	"context"
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
// They are built from the rows alone, as no event row is deleted yet. A day with no event and nothing in force is
// written only to empty one that had some, and a day whose events are more than a cold day holds is not packaged.
func packageEventsDay(ctx context.Context, q *sitedb.Queries, start int64) error {
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
	slices.SortStableFunc(events, func(a, b map[string]any) int { return int(a["occurred_at"].(int64) - b["occurred_at"].(int64)) })
	date := time.UnixMilli(start).UTC().Format(time.DateOnly)
	before, err := inForce(ctx, q, start)
	if err != nil {
		return err
	}
	if len(events) == 0 && len(before) == 0 {
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
	return q.UpsertColdDay(ctx, sitedb.UpsertColdDayParams{Kind: retention.EventsKind, Date: date, Body: packed})
}
