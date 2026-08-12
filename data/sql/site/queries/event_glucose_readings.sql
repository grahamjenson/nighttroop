
-------
-- CREATE 
--------

-- name: CreateGlucoseReading :one
INSERT INTO glucose_readings (
    id,
    reading,
    reading_type,
    
    notes,
    reason,
    utc_offset_minutes,
    occurred_at,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (
    sqlc.arg('ID'),
    sqlc.arg('Reading'),
    sqlc.arg('ReadingType'),
    sqlc.arg('Notes'),
    sqlc.arg('Reason'),
    sqlc.arg('UtcOffsetMinutes'),
    sqlc.arg('OccurredAt'),
    sqlc.arg('RawJsonData'),
    sqlc.arg('CreatedByAppName'),
    sqlc.arg('CreatedByDeviceName'),
    sqlc.arg('CreatedByIdentityName')
)
RETURNING *;

--------
-- READ
--------

-- name: GetGlucoseReading :one
SELECT * FROM glucose_readings WHERE id = ?;

--
-- There are three queries here for performance
-- 1. All default order with limit
-- 2. Default order with search terms limit and skip
-- I wont optimize further or compromise on non-prepared statements (until it becomes necessary)!

-- name: AllGlucoseReadings :many
SELECT * FROM glucose_readings 
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: SearchGlucoseReadings :many
SELECT * FROM glucose_readings
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
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

    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR reading_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR reading_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR reading_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR reading_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))

    AND (CAST(sqlc.narg('SgvEQ') AS INTEGER) IS NULL OR reading = sqlc.narg('SgvEQ'))
    AND (CAST(sqlc.narg('SgvGTE') AS INTEGER) IS NULL OR reading >= sqlc.narg('SgvGTE'))
    AND (CAST(sqlc.narg('SgvGT') AS INTEGER) IS NULL OR reading > sqlc.narg('SgvGT'))
    AND (CAST(sqlc.narg('SgvLTE') AS INTEGER) IS NULL OR reading <= sqlc.narg('SgvLTE'))
    AND (CAST(sqlc.narg('SgvLT') AS INTEGER) IS NULL OR reading < sqlc.narg('SgvLT'))
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


--------
-- UPDATE
--------

-- name: UpdateGlucoseReading :one
UPDATE glucose_readings SET
    reading = COALESCE(sqlc.narg('Reading'), reading),

    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),
    
    occurred_at = COALESCE(sqlc.narg('OccurredAt'), occurred_at),
    utc_offset_minutes = COALESCE(sqlc.narg('UtcOffsetMinutes'), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteGlucoseReading :execresult
UPDATE glucose_readings SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteGlucoseReading :execresult
DELETE FROM glucose_readings WHERE id = ?;


-- name: SoftDeleteGlucoseReadingsByType :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE reading_type = ? AND deleted_at IS NULL;

-- name: HardDeleteGlucoseReadingsByType :execresult
DELETE FROM glucose_readings WHERE reading_type = ?;