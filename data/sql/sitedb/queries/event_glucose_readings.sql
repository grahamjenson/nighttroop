
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
    trend,
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
    sqlc.narg('Trend'),
    NULLIF(CAST(sqlc.narg('Noise') AS REAL), 0),
    sqlc.narg('Filtered'),
    sqlc.narg('Unfiltered'),
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
-- With EntriesOnly, only an entry: a treatment's glucose, which its event group refers to, is the treatment's, not an
-- entry, as Nightscout keeps it, and a loop would take it for a CGM reading
SELECT * FROM glucose_readings WHERE 
(CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
AND (CAST(sqlc.arg('EntriesOnly') AS BOOL) = FALSE OR NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id))
AND glucose_readings.id = sqlc.arg('ID');

-- name: GetGlucoseReadingsByIDs :many
-- The entries with the identifiers in IDs, a JSON array, in no order, each found by its primary key in one query: not
-- a treatment's glucose (see GetGlucoseReading)
SELECT * FROM glucose_readings
WHERE id IN (SELECT value FROM json_each(CAST(sqlc.narg('IDs') AS TEXT)))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
    AND NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id);

-- name: OldestGlucoseReadingAt :one
-- When the oldest reading not deleted was, or 0 when there are none
SELECT CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) FROM glucose_readings WHERE deleted_at IS NULL;

--
-- There are three queries here for performance
-- 1. All default order with limit
-- 2. Default order with search terms limit and skip
-- I wont optimize further or compromise on non-prepared statements (until it becomes necessary)!

-- name: AllGlucoseReadings :many
-- Before pages by date from occurred_at's index: the rows older than a time, or than the row at it with BeforeID,
-- by time then id, so rows at the same time are each shown once. With EntriesOnly, only entries (see GetGlucoseReading).
SELECT * FROM glucose_readings 
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
AND (CAST(sqlc.arg('EntriesOnly') AS BOOL) = FALSE OR NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id))
AND occurred_at <= COALESCE(CAST(sqlc.narg('Before') AS INTEGER), 253402300799999)
AND (sqlc.narg('Before') IS NULL OR occurred_at < sqlc.narg('Before') OR id < CAST(sqlc.narg('BeforeID') AS TEXT))
ORDER BY occurred_at DESC, id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GlucoseReadingTotals :one
-- How many glucose readings AllGlucoseReadings pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM glucose_readings WHERE deleted_at IS NULL;



-- name: GlucoseReadingAscendingIDs :many
-- The entries in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date: not a
-- treatment's glucose (see GetGlucoseReading)
SELECT id FROM glucose_readings
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
    AND NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id)
ORDER BY occurred_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GlucoseReadingSortedIDs :many
-- The entries in the dates searched, by changed (SortOrder 3, -3 newest first) or created (2, -2), for srvModified or
-- srvCreated: not a treatment's glucose (see GetGlucoseReading)
SELECT id FROM (
    SELECT id, CASE CAST(sqlc.arg('SortOrder') AS INTEGER) WHEN 3 THEN updated_at WHEN -3 THEN -updated_at WHEN 2 THEN created_at ELSE -created_at END AS sort_key
    FROM glucose_readings
    WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
        AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
        AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
        AND NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id)
) ORDER BY sort_key, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GlucoseReadingHistoryIDs :many
-- The entries changed since Since, deleted ones too, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history: not a treatment's glucose (see GetGlucoseReading)
SELECT id FROM glucose_readings WHERE glucose_readings.updated_at >= sqlc.arg('Since') AND NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id)
ORDER BY updated_at, occurred_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchGlucoseReadings :many
-- With EntriesOnly, only entries (see GetGlucoseReading)
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
    AND (CAST(sqlc.arg('EntriesOnly') AS BOOL) = FALSE OR NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id))
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: TreatmentGlucoseReadingIDs :many
-- The readings from From to To that a treatment refers to: its glucose, which the site shows as a meter's reading, as
-- it is typed in or read for the treatment, not a CGM's
SELECT glucose_readings.id FROM glucose_readings JOIN event_groups ON event_groups.glucose_reading_id = glucose_readings.id
WHERE glucose_readings.occurred_at BETWEEN CAST(sqlc.arg('From') AS INTEGER) AND CAST(sqlc.arg('To') AS INTEGER)
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: AlarmGlucoseReadings :many
-- The CGM readings back to Minutes before the newest, however old it is, newest first: no data is measured from the
-- newest, and an alarm lasts by its readings, not by the time now. A treatment's glucose is not a CGM reading.
SELECT * FROM glucose_readings
WHERE reading_type = 'sgv' AND deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups AS treatment WHERE treatment.glucose_reading_id = glucose_readings.id)
AND occurred_at >= (SELECT MAX(newest.occurred_at) FROM glucose_readings newest WHERE newest.reading_type = 'sgv'
    AND newest.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM event_groups WHERE event_groups.glucose_reading_id = newest.id))
    - CAST(sqlc.arg('Minutes') AS INTEGER) * 60000
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);


--------
-- UPDATE
--------

-- name: RewriteGlucoseReading :one
-- RewriteGlucoseReading writes every column CreateGlucoseReading does over the glucose reading with the ID, as a
-- Nightscout update writes its merged document. It keeps when it was created and who by: the identity given is who
-- updated it.
UPDATE glucose_readings SET
    id = sqlc.arg('ID'),
    mgdl = sqlc.arg('Mgdl'),
    reading_type = sqlc.arg('ReadingType'),
    units = sqlc.arg('Units'),
    measurement_method = sqlc.arg('MeasurementMethod'),
    delta = sqlc.arg('Delta'),
    trend = sqlc.narg('Trend'),
    noise = NULLIF(CAST(sqlc.narg('Noise') AS REAL), 0),
    filtered = sqlc.narg('Filtered'),
    unfiltered = sqlc.narg('Unfiltered'),
    rssi = NULLIF(CAST(sqlc.narg('Rssi') AS REAL), 0),
    is_calibration = sqlc.narg('IsCalibration'),
    sensor_mgdl_at_pairing = sqlc.narg('SensorMgdlAtPairing'),
    notes = sqlc.arg('Notes'),
    reason = sqlc.arg('Reason'),
    utc_offset_minutes = sqlc.arg('UtcOffsetMinutes'),
    occurred_at = sqlc.arg('OccurredAt'),
    created_by_app_name = sqlc.arg('CreatedByAppName'),
    created_by_device_name = sqlc.arg('CreatedByDeviceName'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = CAST(sqlc.arg('CreatedByIdentityName') AS TEXT),
    updated_by_identity_id = sqlc.narg('CreatedByIdentityID')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateGlucoseReading :one
UPDATE glucose_readings SET
    mgdl = COALESCE(sqlc.narg('Mgdl'), mgdl),

    delta = COALESCE(sqlc.narg('Delta'), delta),
    trend = COALESCE(sqlc.narg('Trend'), trend),
    noise = COALESCE(NULLIF(CAST(sqlc.narg('Noise') AS REAL), 0), noise),
    filtered = COALESCE(sqlc.narg('Filtered'), filtered),
    unfiltered = COALESCE(sqlc.narg('Unfiltered'), unfiltered),
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


-- SoftDeleteGlucoseReadingsByType deletes at most Limit entries of the type, the newest first, not a treatment's glucose
-- name: SoftDeleteGlucoseReadingsByType :execresult
-- Its type is CAST, as SQLite prepares a query again on each run when a bare parameter equals a column it plans around (data/sql/README.md)
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id IN (
    SELECT g.id FROM glucose_readings g WHERE g.reading_type = CAST(sqlc.arg('ReadingType') AS TEXT) AND g.deleted_at IS NULL
        AND NOT EXISTS (SELECT 1 FROM event_groups WHERE event_groups.glucose_reading_id = g.id)
    ORDER BY g.occurred_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER)
);

-- HardDeleteGlucoseReadingsByType deletes at most Limit entries of the type, soft deleted or not, the newest first, not
-- a treatment's glucose
-- name: HardDeleteGlucoseReadingsByType :execresult
-- Its type is CAST, as SQLite prepares a query again on each run when a bare parameter equals a column it plans around (data/sql/README.md)
DELETE FROM glucose_readings WHERE id IN (
    SELECT g.id FROM glucose_readings g WHERE g.reading_type = CAST(sqlc.arg('ReadingType') AS TEXT)
        AND NOT EXISTS (SELECT 1 FROM event_groups WHERE event_groups.glucose_reading_id = g.id)
    ORDER BY g.occurred_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER)
);