
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

-- DateGTE
-- DateGT
-- DateLTE
-- DateLT
-- CreatedAtGTE
-- CreatedAtGT
-- CreatedAtLTE
-- CreatedAtLT
-- UpdatedGTE
-- UpdatedGT
-- UpdatedLTE
-- UpdatedLT
-- TypeEq
-- TypeNeq
-- TypeIn
-- TypeNin
-- IDEq
-- IDIn
-- DeviceEq
-- AppEq
-- Limit
-- Skip
-- SgvEQ
-- SgvGTE
-- SgvGT
-- SgvLTE
-- SgvLT
-- IncludeDeleted

-- name: SearchGlucoseReadings :many
SELECT * FROM glucose_readings
WHERE
    (CAST(sqlc.narg(DateGTE) AS INTEGER) IS NULL OR occurred_at >= sqlc.narg(DateGTE))
    AND (CAST(sqlc.narg(DateLTE) AS INTEGER) IS NULL OR occurred_at <= sqlc.narg(DateLTE))
    AND (CAST(sqlc.narg(DateGT) AS INTEGER) IS NULL OR occurred_at > sqlc.narg(DateGT))
    AND (CAST(sqlc.narg(DateLT) AS INTEGER) IS NULL OR occurred_at < sqlc.narg(DateLT))
    AND CAST(:SortOrder AS TEXT) IS NOT NULL
ORDER BY
    CASE WHEN :SortOrder = 'occurred_at' THEN occurred_at END ASC,
    CASE WHEN :SortOrder = '-occurred_at' THEN occurred_at END DESC,
    CASE WHEN :SortOrder = 'created_at' THEN created_at END ASC,
    CASE WHEN :SortOrder = '-created_at' THEN created_at END DESC,
    CASE WHEN :SortOrder = 'updated_at' THEN updated_at END ASC,
    CASE WHEN :SortOrder = '-updated_at' THEN updated_at END DESC,
    created_at ASC
LIMIT sqlc.arg(limit);

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