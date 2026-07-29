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
    device_state_id,
    loop_state_id,
    pump_state_id,
    cgm_state_id,
    activity_id,
    note_id,

    occurred_at,
    utc_offset_minutes,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllEventGroups :many
SELECT * FROM event_groups 
LEFT JOIN glucose_readings gr ON event_groups.glucose_reading_id = gr.id
LEFT JOIN insulin_doses idose ON event_groups.insulin_dose_id = idose.id
LEFT JOIN food_intakes fi ON event_groups.food_intake_id = fi.id
LEFT JOIN temp_basals tb ON event_groups.temp_basal_id = tb.id
LEFT JOIN temp_targets tt ON event_groups.temp_target_id = tt.id
LEFT JOIN profile_switches ps ON event_groups.profile_switch_id = ps.id
LEFT JOIN device_states ds ON event_groups.device_state_id = ds.id
LEFT JOIN loop_states ls ON event_groups.loop_state_id = ls.id
LEFT JOIN pump_states ps2 ON event_groups.pump_state_id = ps2.id
LEFT JOIN cgm_states cs ON event_groups.cgm_state_id = cs.id
LEFT JOIN activities a ON event_groups.activity_id = a.id
LEFT JOIN notes n ON event_groups.note_id = n.id
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetEventGroup :one
SELECT * FROM event_groups WHERE id = ?;

-- -- name: SearchWithDefaultSortEventGroups :many
-- SELECT * FROM event_groups
-- WHERE
--     (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
--     AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
--     AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR occurred_at > sqlc.narg('DateGT'))
--     AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR occurred_at < sqlc.narg('DateLT'))

--     AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('CreatedAtGTE'))
--     AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('CreatedAtLTE'))
--     AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('CreatedAtGT'))
--     AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('CreatedAtLT'))

--     AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR updated_at >= sqlc.narg('UpdatedGTE'))
--     AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR updated_at <= sqlc.narg('UpdatedLTE'))
--     AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR updated_at > sqlc.narg('UpdatedGT'))
--     AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR updated_at < sqlc.narg('UpdatedLT'))

--     AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
--     AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

--     AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR name = sqlc.narg('NameEq'))

--     AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
--     AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    
--     AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR activity_type = sqlc.narg('TypeEq'))
--     AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR activity_type != sqlc.narg('TypeNeq'))
--     AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR activity_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
--     AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR activity_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

--     AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR duration_minutes = sqlc.narg('DurationEQ'))
--     AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR duration_minutes >= sqlc.narg('DurationGTE'))
--     AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR duration_minutes > sqlc.narg('DurationGT'))
--     AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR duration_minutes <= sqlc.narg('DurationLTE'))
--     AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR duration_minutes < sqlc.narg('DurationLT'))
    
--     AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
-- ORDER BY occurred_at DESC
-- LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateEventGroup :one
UPDATE event_groups SET
    group_type = COALESCE(sqlc.narg('GroupType'), group_type),
    glucose_reading_id = COALESCE(sqlc.narg('GlucoseReadingID'), glucose_reading_id),
    insulin_dose_id = COALESCE(sqlc.narg('InsulinDoseID'), insulin_dose_id),
    food_intake_id = COALESCE(sqlc.narg('FoodIntakeID'), food_intake_id),
    temp_basal_id = COALESCE(sqlc.narg('TempBasalID'), temp_basal_id),
    temp_target_id = COALESCE(sqlc.narg('TempTargetID'), temp_target_id),
    profile_switch_id = COALESCE(sqlc.narg('ProfileSwitchID'), profile_switch_id),
    device_state_id = COALESCE(sqlc.narg('DeviceStateID'), device_state_id),
    loop_state_id = COALESCE(sqlc.narg('LoopStateID'), loop_state_id),
    pump_state_id = COALESCE(sqlc.narg('PumpStateID'), pump_state_id),
    cgm_state_id = COALESCE(sqlc.narg('CgmStateID'), cgm_state_id),
    activity_id = COALESCE(sqlc.narg('ActivityID'), activity_id),
    note_id = COALESCE(sqlc.narg('NoteID'), note_id),

    occurred_at = COALESCE(sqlc.narg('OccurredAt'), occurred_at),
    utc_offset_minutes = COALESCE(sqlc.narg('UtcOffsetMinutes'), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ? AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteEventGroup :one
UPDATE event_groups SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ? AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteEventGroup :one
DELETE FROM event_groups WHERE id = ? RETURNING *;