-------
-- CREATE 
--------
-- name: CreateActivity :one
INSERT INTO activities (
    id,
    name,
    activity_type,
    duration_minutes,
    calories,
    steps,
    heart_rate,
    notes,
    reason, 

    occurred_at,
    utc_offset_minutes,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllActivities :many
SELECT * FROM activities
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetActivity :one
SELECT * FROM activities WHERE id = ?;

-- name: SearchActivities :many
SELECT * FROM activities
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateEQ') AS INTEGER) IS NULL OR occurred_at = sqlc.narg('DateEQ'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR occurred_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR occurred_at < sqlc.narg('DateLT'))

    AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('CreatedAtGTE'))
    AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('CreatedAtLTE'))
    AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('CreatedAtGT'))
    AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('CreatedAtLT'))

    AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR updated_at >= sqlc.narg('UpdatedGTE'))
    AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR updated_at <= sqlc.narg('UpdatedLTE'))
    AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR updated_at > sqlc.narg('UpdatedGT'))
    AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR updated_at < sqlc.narg('UpdatedLT'))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

    AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR name = sqlc.narg('NameEq'))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    
    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR activity_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR activity_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR activity_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR activity_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR duration_minutes = sqlc.narg('DurationEQ'))
    AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR duration_minutes >= sqlc.narg('DurationGTE'))
    AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR duration_minutes > sqlc.narg('DurationGT'))
    AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR duration_minutes <= sqlc.narg('DurationLTE'))
    AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR duration_minutes < sqlc.narg('DurationLT'))
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateActivity :one
UPDATE activities SET
    name = COALESCE(sqlc.narg('Name'), name),
    activity_type = COALESCE(sqlc.narg('ActivityType'), activity_type),
    duration_minutes = COALESCE(sqlc.narg('DurationMinutes'), duration_minutes),
    calories = COALESCE(sqlc.narg('Calories'), calories),
    steps = COALESCE(sqlc.narg('Steps'), steps),
    heart_rate = COALESCE(sqlc.narg('HeartRate'), heart_rate),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),

    raw_json_data = json_patch(
        COALESCE(raw_json_data, '{}'),
        COALESCE(CAST(sqlc.narg('RawJsonData') AS TEXT), '{}')
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteActivity :execresult
UPDATE activities SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteActivity :execresult
DELETE FROM activities WHERE id = ?;

-- name: SoftDeleteActivitiesByType :execresult
UPDATE activities SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE activity_type = ? AND deleted_at IS NULL;

-- name: HardDeleteActivitiesByType :execresult
DELETE FROM activities WHERE activity_type = ?;