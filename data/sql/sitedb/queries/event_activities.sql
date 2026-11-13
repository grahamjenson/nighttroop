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
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (
    sqlc.arg('ID'),
    sqlc.arg('Name'),
    sqlc.arg('ActivityType'),
    sqlc.arg('DurationMinutes'),
    sqlc.narg('Calories'),
    sqlc.narg('Steps'),
    NULLIF(CAST(sqlc.narg('HeartRate') AS INTEGER), 0),
    sqlc.arg('Notes'),
    sqlc.arg('Reason'),
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

-- name: AllActivities :many
SELECT * FROM activities
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
AND occurred_at <= COALESCE(CAST(sqlc.narg('Before') AS INTEGER), 253402300799999)
AND (sqlc.narg('Before') IS NULL OR occurred_at < sqlc.narg('Before') OR id < CAST(sqlc.narg('BeforeID') AS TEXT))
ORDER BY occurred_at DESC, id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ActivityTotals :one
-- How many activities AllActivities pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM activities WHERE deleted_at IS NULL;


-- name: GetActivity :one
SELECT * FROM activities WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);


-- name: ActivityAscendingIDs :many
-- The activities in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date
SELECT id FROM activities
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ActivityHistoryIDs :many
-- The activities changed since Since, deleted ones too, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT id FROM activities WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at, occurred_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchActivities :many
SELECT * FROM activities
WHERE
    -- The date filters as one range, occurred_at's CHECK where unset, so SQLite reads only that range of its index
    occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
        AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))

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
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: RewriteActivity :one
-- RewriteActivity writes every column CreateActivity does over the activity with the ID, as a Nightscout update writes
-- its merged document. It keeps when it was created and who by: the identity given is who updated it.
UPDATE activities SET
    id = sqlc.arg('ID'),
    name = sqlc.arg('Name'),
    activity_type = sqlc.arg('ActivityType'),
    duration_minutes = sqlc.arg('DurationMinutes'),
    calories = sqlc.narg('Calories'),
    steps = sqlc.narg('Steps'),
    heart_rate = NULLIF(CAST(sqlc.narg('HeartRate') AS INTEGER), 0),
    notes = sqlc.arg('Notes'),
    reason = sqlc.arg('Reason'),
    occurred_at = sqlc.arg('OccurredAt'),
    utc_offset_minutes = sqlc.arg('UtcOffsetMinutes'),
    created_by_app_name = sqlc.arg('CreatedByAppName'),
    created_by_device_name = sqlc.arg('CreatedByDeviceName'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = CAST(sqlc.arg('CreatedByIdentityName') AS TEXT),
    updated_by_identity_id = sqlc.arg('CreatedByIdentityID')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateActivity :one
UPDATE activities SET
    name = COALESCE(sqlc.narg('Name'), name),
    activity_type = COALESCE(sqlc.narg('ActivityType'), activity_type),
    duration_minutes = COALESCE(sqlc.narg('DurationMinutes'), duration_minutes),
    calories = COALESCE(sqlc.narg('Calories'), calories),
    steps = COALESCE(sqlc.narg('Steps'), steps),
    heart_rate = COALESCE(NULLIF(CAST(sqlc.narg('HeartRate') AS INTEGER), 0), heart_rate),
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

-- name: SoftDeleteActivity :execresult
UPDATE activities SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteActivity :execresult
DELETE FROM activities WHERE id = ?;

-- name: SoftDeleteActivitiesByType :execresult
UPDATE activities SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = ?
WHERE activity_type = ? AND deleted_at IS NULL;

-- name: HardDeleteActivitiesByType :execresult
DELETE FROM activities WHERE activity_type = ?;