-------
-- CREATE 
--------
-- name: CreateBolus :one
INSERT INTO bolus (
        id,
        occurred_at,
        utc_offset_minutes,
        value,
        type,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllBoluses :many
SELECT * FROM bolus ORDER BY occurred_at DESC; 

-- name: GetBolus :one
SELECT * FROM bolus WHERE id = ?;

--------
-- UPDATE
--------

-- name: UpdateBolus :one
UPDATE bolus SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    value = COALESCE(sqlc.narg(value), value),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------
