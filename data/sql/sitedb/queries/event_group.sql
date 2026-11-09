-------
-- CREATE 
--------
-- name: CreateEventGroup :one
INSERT INTO event_groups (
    id,
    group_type,
    glucose_reading_id,
    insulin_dose_id,
    food_intake_id,
    temp_basal_id,
    temp_target_id,
    profile_switch_id,
    component_event_id,
    activity_id,
    notification_event_id,

    notes,
    reason,
    duration_minutes,
    entered_by,
    uuid,
    mode,
    auto_forced,
    reasons,
    original_duration_ms,
    duration_type,
    sent_hash,

    occurred_at,
    utc_offset_minutes,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (
    sqlc.arg('ID'),
    sqlc.arg('GroupType'),
    sqlc.arg('GlucoseReadingID'),
    sqlc.arg('InsulinDoseID'),
    sqlc.arg('FoodIntakeID'),
    sqlc.arg('TempBasalID'),
    sqlc.arg('TempTargetID'),
    sqlc.arg('ProfileSwitchID'),
    sqlc.arg('ComponentEventID'),
    sqlc.arg('ActivityID'),
    sqlc.arg('NotificationEventID'),
    sqlc.arg('Notes'),
    sqlc.arg('Reason'),
    sqlc.arg('DurationMinutes'),
    sqlc.arg('EnteredBy'),
    NULLIF(CAST(sqlc.narg('Uuid') AS TEXT), ''),
    NULLIF(CAST(sqlc.narg('Mode') AS TEXT), ''),
    sqlc.narg('AutoForced'),
    NULLIF(CAST(sqlc.narg('Reasons') AS TEXT), ''),
    sqlc.narg('OriginalDurationMs'),
    NULLIF(CAST(sqlc.narg('DurationType') AS TEXT), ''),
    NULLIF(CAST(sqlc.narg('SentHash') AS TEXT), ''),
    sqlc.arg('OccurredAt'),
    sqlc.arg('UtcOffsetMinutes'),
    sqlc.arg('CreatedByAppName'),
    sqlc.arg('CreatedByDeviceName'),
    sqlc.arg('CreatedByIdentityName'),
    sqlc.arg('CreatedByIdentityID')
)
RETURNING *;

--------
-- READ
--------

-- name: GetEventGroupWithoutJoins :one
SELECT * from event_groups WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL);

-- name: GetEventGroupOfTypeWithoutJoins :many
SELECT * from event_groups WHERE group_type = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- The joins name the _vw views so sqlc types their columns as nullable, but the generated queries join the tables
-- themselves (see joinTablesNotViews in data/sql/gen/sqlc.go), so SQLite does not copy each table for every query
-- name: AllEventGroups :many
SELECT 
    -- It would be awesome if sqlc worked with embed null objects
    -- see here https://github.com/sqlc-dev/sqlc/issues/2348
    sqlc.embed(eg),
    
    sqlc.embed(gr),
    sqlc.embed(idose),
    sqlc.embed(fi),
    sqlc.embed(tb),
    sqlc.embed(tt),
    sqlc.embed(ps),
    sqlc.embed(ce),
    sqlc.embed(a),
    sqlc.embed(n)

FROM event_groups eg
LEFT JOIN glucose_readings_vw gr ON eg.glucose_reading_id = gr.id
LEFT JOIN insulin_doses_vw idose ON eg.insulin_dose_id = idose.id
LEFT JOIN food_intakes_vw fi ON eg.food_intake_id = fi.id
LEFT JOIN temp_basals_vw tb ON eg.temp_basal_id = tb.id
LEFT JOIN temp_targets_vw tt ON eg.temp_target_id = tt.id
LEFT JOIN profile_switches_vw ps ON eg.profile_switch_id = ps.id
LEFT JOIN component_events_vw ce ON eg.component_event_id = ce.id
LEFT JOIN activities_vw a ON eg.activity_id = a.id
LEFT JOIN notification_events_vw n ON eg.notification_event_id = n.id


WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) IS TRUE OR eg.deleted_at IS NULL)
ORDER BY eg.occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: EventGroupTotals :one
-- How many event groups AllEventGroups pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM event_groups WHERE deleted_at IS NULL;


-- name: GetEventGroup :one
SELECT
    -- It would be awesome if sqlc worked with embed null objects
    -- see here https://github.com/sqlc-dev/sqlc/issues/2348
    sqlc.embed(eg),
    
    sqlc.embed(gr),
    sqlc.embed(idose),
    sqlc.embed(fi),
    sqlc.embed(tb),
    sqlc.embed(tt),
    sqlc.embed(ps),
    sqlc.embed(ce),
    sqlc.embed(a),
    sqlc.embed(n)

FROM event_groups eg
LEFT JOIN glucose_readings_vw gr ON eg.glucose_reading_id = gr.id
LEFT JOIN insulin_doses_vw idose ON eg.insulin_dose_id = idose.id
LEFT JOIN food_intakes_vw fi ON eg.food_intake_id = fi.id
LEFT JOIN temp_basals_vw tb ON eg.temp_basal_id = tb.id
LEFT JOIN temp_targets_vw tt ON eg.temp_target_id = tt.id
LEFT JOIN profile_switches_vw ps ON eg.profile_switch_id = ps.id
LEFT JOIN component_events_vw ce ON eg.component_event_id = ce.id
LEFT JOIN activities_vw a ON eg.activity_id = a.id
LEFT JOIN notification_events_vw n ON eg.notification_event_id = n.id

WHERE eg.id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR eg.deleted_at IS NULL);



-- name: EventGroupAscendingIDs :many
-- The treatments in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date
SELECT id FROM event_groups
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: EventGroupHistoryIDs :many
-- The treatments changed since Since, deleted ones too, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT id FROM event_groups WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at, occurred_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchEventGroups :many
SELECT
    -- It would be awesome if sqlc worked with embed null objects
    -- see here https://github.com/sqlc-dev/sqlc/issues/2348
    sqlc.embed(eg),
    
    sqlc.embed(gr),
    sqlc.embed(idose),
    sqlc.embed(fi),
    sqlc.embed(tb),
    sqlc.embed(tt),
    sqlc.embed(ps),
    sqlc.embed(ce),
    sqlc.embed(a),
    sqlc.embed(n)

FROM event_groups eg
LEFT JOIN glucose_readings_vw gr ON eg.glucose_reading_id = gr.id
LEFT JOIN insulin_doses_vw idose ON eg.insulin_dose_id = idose.id
LEFT JOIN food_intakes_vw fi ON eg.food_intake_id = fi.id
LEFT JOIN temp_basals_vw tb ON eg.temp_basal_id = tb.id
LEFT JOIN temp_targets_vw tt ON eg.temp_target_id = tt.id
LEFT JOIN profile_switches_vw ps ON eg.profile_switch_id = ps.id
LEFT JOIN component_events_vw ce ON eg.component_event_id = ce.id
LEFT JOIN activities_vw a ON eg.activity_id = a.id
LEFT JOIN notification_events_vw n ON eg.notification_event_id = n.id


WHERE
    -- The date filters as one range, occurred_at's CHECK where unset, so SQLite reads only that range of its index
    eg.occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
        AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))

    AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR eg.created_at >= sqlc.narg('CreatedAtGTE'))
    AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR eg.created_at <= sqlc.narg('CreatedAtLTE'))
    AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR eg.created_at > sqlc.narg('CreatedAtGT'))
    AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR eg.created_at < sqlc.narg('CreatedAtLT'))

    AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR eg.updated_at >= sqlc.narg('UpdatedGTE'))
    AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR eg.updated_at <= sqlc.narg('UpdatedLTE'))
    AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR eg.updated_at > sqlc.narg('UpdatedGT'))
    AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR eg.updated_at < sqlc.narg('UpdatedLT'))

    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR eg.group_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR eg.group_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR eg.group_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR eg.group_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR eg.id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR eg.id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))
    
    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR eg.created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR eg.created_by_app_name = sqlc.narg('AppEq'))

    -- model specific stuff

    AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR fi.name = sqlc.narg('NameEq') OR a.name = sqlc.narg('NameEq') OR ps.name = sqlc.narg('NameEq'))


    -- glucose readings
    AND (CAST(sqlc.narg('SgvEQ') AS INTEGER) IS NULL OR gr.mgdl = sqlc.narg('SgvEQ'))
    AND (CAST(sqlc.narg('SgvGTE') AS INTEGER) IS NULL OR gr.mgdl >= sqlc.narg('SgvGTE'))
    AND (CAST(sqlc.narg('SgvGT') AS INTEGER) IS NULL OR gr.mgdl > sqlc.narg('SgvGT'))
    AND (CAST(sqlc.narg('SgvLTE') AS INTEGER) IS NULL OR gr.mgdl <= sqlc.narg('SgvLTE'))
    AND (CAST(sqlc.narg('SgvLT') AS INTEGER) IS NULL OR gr.mgdl < sqlc.narg('SgvLT'))
    
    -- food intake

    AND (CAST(sqlc.narg('CarbsEQ') AS INTEGER) IS NULL OR fi.carbs = sqlc.narg('CarbsEQ'))
    AND (CAST(sqlc.narg('CarbsGTE') AS INTEGER) IS NULL OR fi.carbs >= sqlc.narg('CarbsGTE'))
    AND (CAST(sqlc.narg('CarbsGT') AS INTEGER) IS NULL OR fi.carbs > sqlc.narg('CarbsGT'))
    AND (CAST(sqlc.narg('CarbsLTE') AS INTEGER) IS NULL OR fi.carbs <= sqlc.narg('CarbsLTE'))
    AND (CAST(sqlc.narg('CarbsLT') AS INTEGER) IS NULL OR fi.carbs < sqlc.narg('CarbsLT'))
    

    -- insulin doses
    AND (CAST(sqlc.narg('InsulinEQ') AS INTEGER) IS NULL OR idose.dose = sqlc.narg('InsulinEQ'))
    AND (CAST(sqlc.narg('InsulinGTE') AS INTEGER) IS NULL OR idose.dose >= sqlc.narg('InsulinGTE'))
    AND (CAST(sqlc.narg('InsulinGT') AS INTEGER) IS NULL OR idose.dose > sqlc.narg('InsulinGT'))
    AND (CAST(sqlc.narg('InsulinLTE') AS INTEGER) IS NULL OR idose.dose <= sqlc.narg('InsulinLTE'))
    AND (CAST(sqlc.narg('InsulinLT') AS INTEGER) IS NULL OR idose.dose < sqlc.narg('InsulinLT'))



    -- range values
    AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR a.duration_minutes = sqlc.narg('DurationEQ') OR tb.duration_minutes = sqlc.narg('DurationEQ') OR tt.duration_minutes = sqlc.narg('DurationEQ') OR ps.duration_minutes = sqlc.narg('DurationEQ'))
    AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR a.duration_minutes >= sqlc.narg('DurationGTE') OR tb.duration_minutes >= sqlc.narg('DurationGTE') OR tt.duration_minutes >= sqlc.narg('DurationGTE') OR ps.duration_minutes >= sqlc.narg('DurationGTE'))
    AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR a.duration_minutes > sqlc.narg('DurationGT') OR tb.duration_minutes > sqlc.narg('DurationGT') OR tt.duration_minutes > sqlc.narg('DurationGT') OR ps.duration_minutes > sqlc.narg('DurationGT'))
    AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR a.duration_minutes <= sqlc.narg('DurationLTE') OR tb.duration_minutes <= sqlc.narg('DurationLTE') OR tt.duration_minutes <= sqlc.narg('DurationLTE') OR ps.duration_minutes <= sqlc.narg('DurationLTE'))
    AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR a.duration_minutes < sqlc.narg('DurationLT') OR tb.duration_minutes < sqlc.narg('DurationLT') OR tt.duration_minutes < sqlc.narg('DurationLT') OR ps.duration_minutes < sqlc.narg('DurationLT'))

    -- Trio's carbs and temp targets have a carbs or a duration, and who entered one is as sent, else the identity
    AND (CAST(sqlc.narg('CarbsExists') AS BOOL) IS NULL OR (fi.carbs IS NOT NULL) = sqlc.narg('CarbsExists'))
    AND (CAST(sqlc.narg('DurationExists') AS BOOL) IS NULL OR (COALESCE(eg.duration_minutes, a.duration_minutes, tb.duration_minutes, tt.duration_minutes, ps.duration_minutes) IS NOT NULL) = sqlc.narg('DurationExists'))
    AND (CAST(sqlc.narg('EnteredByEq') AS TEXT) IS NULL OR COALESCE(eg.entered_by, eg.created_by_identity_name) = sqlc.narg('EnteredByEq'))
    AND (CAST(sqlc.narg('EnteredByNin') AS TEXT) IS NULL OR COALESCE(eg.entered_by, eg.created_by_identity_name) NOT IN (SELECT value FROM json_each(sqlc.narg('EnteredByNin'))))


    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) IS TRUE OR eg.deleted_at IS NULL)
ORDER BY eg.occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: UpdateEventGroup :one
UPDATE event_groups SET
    glucose_reading_id = sqlc.narg('GlucoseReadingID'),
    insulin_dose_id = sqlc.narg('InsulinDoseID'),
    food_intake_id = sqlc.narg('FoodIntakeID'),
    temp_basal_id = sqlc.narg('TempBasalID'),
    temp_target_id = sqlc.narg('TempTargetID'),
    profile_switch_id = sqlc.narg('ProfileSwitchID'),
    component_event_id = sqlc.narg('ComponentEventID'),
    activity_id = sqlc.narg('ActivityID'),
    notification_event_id = sqlc.narg('NotificationEventID'),

    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),
    entered_by = COALESCE(sqlc.narg('EnteredBy'), entered_by),
    uuid = COALESCE(NULLIF(CAST(sqlc.narg('Uuid') AS TEXT), ''), uuid),
    mode = COALESCE(NULLIF(CAST(sqlc.narg('Mode') AS TEXT), ''), mode),
    auto_forced = COALESCE(sqlc.narg('AutoForced'), auto_forced),
    reasons = COALESCE(NULLIF(CAST(sqlc.narg('Reasons') AS TEXT), ''), reasons),
    original_duration_ms = COALESCE(sqlc.narg('OriginalDurationMs'), original_duration_ms),
    duration_type = COALESCE(NULLIF(CAST(sqlc.narg('DurationType') AS TEXT), ''), duration_type),
    sent_hash = COALESCE(NULLIF(CAST(sqlc.narg('SentHash') AS TEXT), ''), sent_hash),

    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

-- EventsToGroup are the events that can be put in the group with GroupID (empty for a new one): those not deleted, nor in
-- another group, from the window, and the group's own. Kind is the group's column for the event, and amount, minutes
-- and detail what tells it apart, like a dose and its type. Whether an event is in another group is looked up by the
-- group's unique index on its column, as a NOT IN list of every group's events is read again for each event.
-- name: EventsToGroup :many
-- Each argument is given once, in params, as sqlc numbers an argument used in more than one select wrongly
WITH params AS (
    SELECT CAST(sqlc.arg('GroupID') AS TEXT) AS group_id, CAST(sqlc.arg('From') AS INTEGER) AS from_ms,
        CAST(sqlc.arg('To') AS INTEGER) AS to_ms
)
-- sqlc types the columns by the first select, so this one, which has no rows, says which can be null
SELECT '' AS kind, '' AS id, 0 AS occurred_at, CAST(NULL AS REAL) AS amount, CAST(NULL AS INTEGER) AS minutes,
    CAST(NULL AS TEXT) AS detail WHERE FALSE
UNION ALL
SELECT 'glucose_reading_id', e.id, e.occurred_at, mgdl, NULL, reading_type FROM glucose_readings e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.glucose_reading_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT glucose_reading_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'insulin_dose_id', e.id, e.occurred_at, dose, NULL, dose_type FROM insulin_doses e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.insulin_dose_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT insulin_dose_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'food_intake_id', e.id, e.occurred_at, carbs, NULL, name FROM food_intakes e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.food_intake_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT food_intake_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'temp_basal_id', e.id, e.occurred_at, rate, duration_minutes, NULL FROM temp_basals e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.temp_basal_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT temp_basal_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'temp_target_id', e.id, e.occurred_at, low_target_mgdl, duration_minutes, high_target_mgdl FROM temp_targets e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.temp_target_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT temp_target_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'profile_switch_id', e.id, e.occurred_at, NULL, duration_minutes, name FROM profile_switches e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.profile_switch_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT profile_switch_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'component_event_id', e.id, e.occurred_at, NULL, NULL, event_type FROM component_events e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.component_event_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT component_event_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'activity_id', e.id, e.occurred_at, NULL, duration_minutes, COALESCE(name, activity_type) FROM activities e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.activity_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT activity_id FROM event_groups WHERE id = p.group_id))
UNION ALL
SELECT 'notification_event_id', e.id, e.occurred_at, NULL, NULL, title FROM notification_events e, params p
WHERE e.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups g WHERE g.notification_event_id = e.id AND g.id != p.group_id)
    AND (e.occurred_at BETWEEN p.from_ms AND p.to_ms OR e.id = (SELECT notification_event_id FROM event_groups WHERE id = p.group_id))
LIMIT 1000;

-- RegroupEventGroup sets the group's events to exactly those given, as the site's edit page picks them, keeping its
-- type, time, notes and reason when not given
-- name: RegroupEventGroup :one
UPDATE event_groups SET
    group_type = COALESCE(sqlc.narg('GroupType'), group_type),
    occurred_at = COALESCE(sqlc.narg('OccurredAt'), occurred_at),
    glucose_reading_id = sqlc.narg('GlucoseReadingID'),
    insulin_dose_id = sqlc.narg('InsulinDoseID'),
    food_intake_id = sqlc.narg('FoodIntakeID'),
    temp_basal_id = sqlc.narg('TempBasalID'),
    temp_target_id = sqlc.narg('TempTargetID'),
    profile_switch_id = sqlc.narg('ProfileSwitchID'),
    component_event_id = sqlc.narg('ComponentEventID'),
    activity_id = sqlc.narg('ActivityID'),
    notification_event_id = sqlc.narg('NotificationEventID'),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *;

--------
-- DELETE
--------

-- name: SoftDeleteEventGroup :one
UPDATE event_groups SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteEventGroup :one
DELETE FROM event_groups WHERE id = ? RETURNING *;


-- name: SoftDeleteEventGroupsByType :many
UPDATE event_groups SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = ?
WHERE group_type = ? AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteEventGroupsByType :many
DELETE FROM event_groups WHERE group_type = ? RETURNING *;
