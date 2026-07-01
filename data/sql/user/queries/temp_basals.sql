-------
-- CREATE 
--------

-- name: CreateTempBasal :one
INSERT INTO temp_basals (
    id,
    start,
    utc_offset_minutes,
    duration_minutes,
    value,
    raw_json_data, 
    created_by_app_name,
    created_by_device_name, 
    created_by_identity_id,
    created_by_identity_name
    )
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllTempBasals :many
SELECT * FROM temp_basals ORDER BY start DESC;


-- name: GetTempBasal :one
SELECT * FROM temp_basals WHERE id = ?;

--------
-- UPDATE
--------

-- name: UpdateTempBasal :one
UPDATE temp_basals SET
    start = COALESCE(sqlc.narg(start), start),
    duration_minutes = COALESCE(sqlc.narg(duration_minutes), duration_minutes),
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
