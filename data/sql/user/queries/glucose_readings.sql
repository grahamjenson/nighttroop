
-------
-- CREATE 
--------

-- name: CreateGlucoseReading :one
INSERT INTO glucose_readings (
    id,
    occurred_at,
    value,
    type,
    raw_json_data,
    utc_offset_minutes,
    created_by_app_name,
    created_by_device_name, 
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: GetGlucoseReading :one
SELECT * FROM glucose_readings WHERE id = ?;

-- name: AllGlucoseReadings :many
SELECT * FROM glucose_readings ORDER BY occurred_at DESC LIMIT ?;

-- name: SearchGlucoseReadings :many
SELECT * FROM glucose_readings
WHERE
    occurred_at >= COALESCE(sqlc.narg(occurred_at_start), occurred_at) AND occurred_at <= COALESCE(sqlc.narg(occurred_at_end), occurred_at)
ORDER BY
    CASE WHEN CAST(:sort_order AS TEXT) = 'occurred_at_asc' THEN occurred_at END ASC,
    CASE WHEN CAST(:sort_order AS TEXT) = 'occurred_at_desc' THEN occurred_at END DESC
LIMIT COALESCE(sqlc.narg(glimit), 1000);

--------
-- UPDATE
--------

-- name: UpdateGlucoseReading :one
UPDATE glucose_readings SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    value = COALESCE(sqlc.narg(value), value),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = ?,
    updated_by_identity_name = ?
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