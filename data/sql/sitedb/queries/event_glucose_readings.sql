
-------
-- CREATE 
--------

-- name: CreateGlucoseReading :one
INSERT INTO glucose_readings (
    id,
    mgdl,
    reading_type,
    units,
    measurement_method,

    delta,
    direction,
    noise,
    filtered,
    unfiltered,
    rssi,
    is_calibration,
    sensor_mgdl_at_pairing,

    notes,
    reason,
    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (
    sqlc.arg('ID'),
    sqlc.arg('Mgdl'),
    sqlc.arg('ReadingType'),
    sqlc.arg('Units'),
    sqlc.arg('MeasurementMethod'),

    sqlc.arg('Delta'),
    sqlc.arg('Direction'),
    NULLIF(CAST(sqlc.narg('Noise') AS REAL), 0),
    NULLIF(CAST(sqlc.narg('Filtered') AS REAL), 0),
    NULLIF(CAST(sqlc.narg('Unfiltered') AS REAL), 0),
    NULLIF(CAST(sqlc.narg('Rssi') AS REAL), 0),
    sqlc.narg('IsCalibration'),
    sqlc.narg('SensorMgdlAtPairing'),

    sqlc.arg('Notes'),
    sqlc.arg('Reason'),
    sqlc.arg('UtcOffsetMinutes'),
    sqlc.arg('OccurredAt'),
    sqlc.arg('CreatedByAppName'),
    sqlc.arg('CreatedByDeviceName'),
    sqlc.arg('CreatedByIdentityName'),
    sqlc.narg('CreatedByIdentityID')
)
RETURNING *;

--------
-- READ
--------

-- name: GetGlucoseReading :one
SELECT * FROM glucose_readings WHERE 
(CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
AND id = sqlc.arg('ID');

-- name: OldestGlucoseReadingAt :one
-- When the oldest reading not deleted was, or 0 when there are none
SELECT CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) FROM glucose_readings WHERE deleted_at IS NULL;

--
-- There are three queries here for performance
-- 1. All default order with limit
-- 2. Default order with search terms limit and skip
-- I wont optimize further or compromise on non-prepared statements (until it becomes necessary)!

-- name: AllGlucoseReadings :many
SELECT * FROM glucose_readings 
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GlucoseReadingTotals :one
-- How many glucose readings AllGlucoseReadings pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM glucose_readings WHERE deleted_at IS NULL;



-- name: GlucoseReadingAscendingIDs :many
-- The readings in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date
SELECT id FROM glucose_readings
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GlucoseReadingHistoryIDs :many
-- The readings changed since Since, deleted ones too, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT id FROM glucose_readings WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at, occurred_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchGlucoseReadings :many
SELECT * FROM glucose_readings
WHERE
    -- The date filters as one range, occurred_at's CHECK where unset, so SQLite reads only that range of its index
    occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
        AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    -- The exact times of readings, as xDrip4iOS deletes the readings a rebuild hides
    AND (CAST(sqlc.narg('DateIn') AS TEXT) IS NULL OR occurred_at IN (SELECT CAST(value AS INTEGER) FROM json_each(sqlc.narg('DateIn'))))

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

    AND (CAST(sqlc.narg('SgvEQ') AS INTEGER) IS NULL OR mgdl = sqlc.narg('SgvEQ'))
    AND (CAST(sqlc.narg('SgvGTE') AS INTEGER) IS NULL OR mgdl >= sqlc.narg('SgvGTE'))
    AND (CAST(sqlc.narg('SgvGT') AS INTEGER) IS NULL OR mgdl > sqlc.narg('SgvGT'))
    AND (CAST(sqlc.narg('SgvLTE') AS INTEGER) IS NULL OR mgdl <= sqlc.narg('SgvLTE'))
    AND (CAST(sqlc.narg('SgvLT') AS INTEGER) IS NULL OR mgdl < sqlc.narg('SgvLT'))
    
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: AlarmGlucoseReadings :many
-- The CGM readings back to Minutes before the newest, however old it is, newest first: no data is measured from the
-- newest, and an alarm lasts by its readings, not by the time now
SELECT * FROM glucose_readings
WHERE reading_type = 'sgv' AND deleted_at IS NULL
AND occurred_at >= (SELECT MAX(occurred_at) FROM glucose_readings WHERE reading_type = 'sgv' AND deleted_at IS NULL)
    - CAST(sqlc.arg('Minutes') AS INTEGER) * 60000
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);


--------
-- UPDATE
--------

-- name: UpdateGlucoseReading :one
UPDATE glucose_readings SET
    mgdl = COALESCE(sqlc.narg('Mgdl'), mgdl),

    delta = COALESCE(sqlc.narg('Delta'), delta),
    direction = COALESCE(sqlc.narg('Direction'), direction),
    noise = COALESCE(NULLIF(CAST(sqlc.narg('Noise') AS REAL), 0), noise),
    filtered = COALESCE(NULLIF(CAST(sqlc.narg('Filtered') AS REAL), 0), filtered),
    unfiltered = COALESCE(NULLIF(CAST(sqlc.narg('Unfiltered') AS REAL), 0), unfiltered),
    rssi = COALESCE(NULLIF(CAST(sqlc.narg('Rssi') AS REAL), 0), rssi),
    is_calibration = COALESCE(sqlc.narg('IsCalibration'), is_calibration),
    sensor_mgdl_at_pairing = COALESCE(sqlc.narg('SensorMgdlAtPairing'), sensor_mgdl_at_pairing),

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

-- name: SoftDeleteGlucoseReading :execresult
UPDATE glucose_readings SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteGlucoseReading :execresult
DELETE FROM glucose_readings WHERE id = ?;


-- SoftDeleteGlucoseReadingsByType deletes at most Limit of the type, the newest first
-- name: SoftDeleteGlucoseReadingsByType :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id IN (
    SELECT g.id FROM glucose_readings g WHERE g.reading_type = sqlc.arg('ReadingType') AND g.deleted_at IS NULL
    ORDER BY g.occurred_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER)
);

-- HardDeleteGlucoseReadingsByType deletes at most Limit of the type, soft deleted or not, the newest first
-- name: HardDeleteGlucoseReadingsByType :execresult
DELETE FROM glucose_readings WHERE id IN (
    SELECT g.id FROM glucose_readings g WHERE g.reading_type = sqlc.arg('ReadingType')
    ORDER BY g.occurred_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER)
);