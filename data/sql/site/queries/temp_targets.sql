




-------
-- CREATE 
--------


-- name: CreateTempTarget :one
INSERT INTO temp_targets (
    id, 
    start,
    utc_offset_minutes, 
    duration_minutes, 
    high_target, 
    low_target,
    raw_json_data, 
    created_by_app_name,
    created_by_device_name, 
    created_by_identity_id,
    created_by_identity_name
    ) 
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;


--------
-- READ
--------


-- name: AllTempTargets :many
SELECT * FROM temp_targets ORDER BY start DESC;

-- name: GetTempTarget :one
SELECT * FROM temp_targets WHERE id = ?;

--------
-- UPDATE
--------

-- name: UpdateTempTarget :one
UPDATE temp_targets SET
    start = COALESCE(sqlc.narg(start), start),
    duration_minutes = COALESCE(sqlc.narg(duration_minutes), duration_minutes),
    high_target = COALESCE(sqlc.narg(high_target), high_target),
    low_target = COALESCE(sqlc.narg(low_target), low_target),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = ?,
    updated_by_identity_name = ?
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------
