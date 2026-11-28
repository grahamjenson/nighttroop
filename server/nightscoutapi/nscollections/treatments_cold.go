package nscollections

import (
	"context"
	"database/sql"
	"slices"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/httputils"
)

// coldMembers are the tables whose cold events are a treatment's events, each with how a treatment is given a row of
// it to fill
var coldMembers = map[string]func(*sitedb.EventGroupsBundle) any{
	"glucose_readings": func(b *sitedb.EventGroupsBundle) any {
		b.GlucoseReading = &sitedb.GlucoseReading{}
		return b.GlucoseReading
	},
	"insulin_doses": func(b *sitedb.EventGroupsBundle) any { b.InsulinDose = &sitedb.InsulinDose{}; return b.InsulinDose },
	"food_intakes":  func(b *sitedb.EventGroupsBundle) any { b.FoodIntake = &sitedb.FoodIntake{}; return b.FoodIntake },
	"temp_targets":  func(b *sitedb.EventGroupsBundle) any { b.TempTarget = &sitedb.TempTarget{}; return b.TempTarget },
	"profile_switches": func(b *sitedb.EventGroupsBundle) any {
		b.ProfileSwitch = &sitedb.ProfileSwitch{}
		return b.ProfileSwitch
	},
	"activities": func(b *sitedb.EventGroupsBundle) any { b.Activity = &sitedb.Activity{}; return b.Activity },
	"notification_events": func(b *sitedb.EventGroupsBundle) any {
		b.NotificationEvent = &sitedb.NotificationEvent{}
		return b.NotificationEvent
	},
}

// coldTreatmentID is the ID of the treatment a cold event is of, "" when it is of none: an event keeps its
// treatment's ID, and a treatment with no event is its own
func coldTreatmentID(event map[string]any) string {
	if id, ok := event["group_id"].(string); ok {
		return id
	}
	if event["k"] == "event_groups" {
		id, _ := event["id"].(string)
		return id
	}
	return ""
}

// treatmentsCanMatch is whether a treatment made from a cold day can be what the query searches for: one searched for
// by its date and its type alone. A cold event keeps less than its rows did, so a search by another field is of rows
// only.
func treatmentsCanMatch(query nsquery.Query) bool {
	if query.IncludeDeleted {
		return false
	}
	query.TypeEq, query.TypeNeq, query.TypeIn, query.TypeNin = nil, nil, nil, nil
	return query.DatesOnly()
}

// SearchWithCold is a search for treatments, the newest first or the oldest, with those of the packaged days whose rows are deleted,
// and whether it was one: only where the server deletes rows, and only a search a cold event can match that reaches
// back before the hot days. The rows are read first, by rows, in the search's order. A treatment that is still a row
// is not made again, and no cold day is read that is past the last row of the page. A made treatment has what its events kept: a meal, a dose
// a person gave, a target, a switch, a note, a site change. A loop's doses and temp basals are series, not treatments.
func (c TreatmentCollection) SearchWithCold(q *sitedb.Queries, ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query, search func(*sitedb.Queries, context.Context, nsquery.Query) ([]*sitedb.EventGroupsBundle, error)) ([]*sitedb.EventGroupsBundle, bool, error) {
	cfg := siteContext.Config
	if cfg == nil || !cfg.RetentionDeleteRows || !treatmentsCanMatch(query) {
		return nil, false, nil
	}
	from, to := coldRange(query, cfg)
	if from >= to {
		return nil, false, nil
	}
	want := int(query.Skip + query.Limit)
	rowsQuery := query
	rowsQuery.Skip, rowsQuery.Limit = 0, int64(want)
	rows, err := search(q, ctx, rowsQuery)
	if err != nil {
		return nil, true, err
	}
	oldestFirst := query.SortOrder != nil && *query.SortOrder == nsquery.SortOccurredAtAsc
	if len(rows) == want && want > 0 {
		if oldestFirst {
			to = min(to, rows[want-1].EventGroup.OccurredAt+1)
		} else {
			from = max(from, rows[want-1].EventGroup.OccurredAt)
		}
	}
	inRows := map[string]bool{}
	for _, row := range rows {
		inRows[row.EventGroup.ID] = true
	}
	// A treatment can have more than one event, as a meal with its bolus has, so more events are read than treatments
	// are wanted, and each treatment is made once, from all its events that were read
	events, err := db.ColdEvents(ctx, q, from, to, 3*want, oldestFirst, func(event map[string]any) bool {
		id, groupType := coldTreatmentID(event), ""
		if t, ok := event["group_type"].(string); ok {
			groupType = t
		}
		return id != "" && !inRows[id] && typeMatches(query, groupType)
	})
	if err != nil {
		return nil, true, err
	}
	made := map[string]*sitedb.EventGroupsBundle{}
	merged := rows
	for _, event := range events {
		id := coldTreatmentID(event)
		bundle := made[id]
		if bundle == nil {
			bundle = &sitedb.EventGroupsBundle{}
			db.RowOf("event_groups", event, &bundle.EventGroup)
			bundle.EventGroup.ID = id
			made[id], merged = bundle, append(merged, bundle)
		}
		if member, ok := coldMembers[event["k"].(string)]; ok {
			db.RowOf(event["k"].(string), event, member(bundle))
		}
	}
	slices.SortStableFunc(merged, func(a, b *sitedb.EventGroupsBundle) int {
		if oldestFirst {
			a, b = b, a
		}
		return int(min(max(b.EventGroup.OccurredAt-a.EventGroup.OccurredAt, -1), 1))
	})
	return merged[min(int(query.Skip), len(merged)):min(want, len(merged))], true, nil
}

// ColdModel is none: a treatment's identifier does not say which day it is of, so a made treatment is not read or
// deleted by it
func (c TreatmentCollection) ColdModel(*sitedb.Queries, context.Context, *httputils.SiteContext, string, bool) (*sitedb.EventGroupsBundle, error) {
	return nil, sql.ErrNoRows
}
