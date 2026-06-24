
-------
-- CREATE 
--------

-- name: CreateGlucoseReading :execresult
INSERT INTO glucose_readings (
    id,
    occurred_at,
    value,
    raw_json_data,
    utc_offset_minutes,
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: GetGlucoseReading :one
SELECT * FROM glucose_readings WHERE id = ?;

-- name: AllGlucoseReadings :many
SELECT * FROM glucose_readings ORDER BY occurred_at DESC LIMIT ?;

--------
-- UPDATE
--------

-- name: UpdateGlucoseReading :one
UPDATE glucose_readings SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    value = COALESCE(sqlc.narg(value), value),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------

-- name: DeleteGlucoseReading :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = ?;

-- name: DeletePermanentGlucoseReading :execresult
DELETE FROM glucose_readings WHERE id = ?;

-- name: DeleteGlucoseReadingsByType :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE type = ?;

-- name: DeletePermanentGlucoseReadingsByType :execresult
DELETE FROM glucose_readings
WHERE type = ?;